package billimport

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

type planned struct {
	entry           entry
	category, asset int64
	result          int
}

func (s *Service) Import(ctx context.Context, input Input) (Report, error) {
	if input.UserID <= 0 || input.AccountBookID <= 0 || len(input.Files) == 0 || len(input.Files) > 20 {
		return Report{}, invalid("positive user/book IDs and 1-20 files are required")
	}
	var entries []entry
	for i, file := range input.Files {
		rows, err := s.parse(ctx, file, i+1)
		if err != nil {
			return Report{}, err
		}
		entries = append(entries, rows...)
		if len(entries) > MaxRows {
			return Report{}, invalid("too many rows in import batch")
		}
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Occurred.Before(entries[j].Occurred) })
	var report Report
	if !input.Apply {
		catalog, err := s.repository.Catalog(ctx, input.AccountBookID, input.UserID)
		if err != nil {
			return report, storageError(err)
		}
		report, _, _, err = makePlan(entries, input, catalog)
		return report, err
	}
	err := s.repository.WithinImportTransaction(ctx, input.AccountBookID, input.UserID, func(tx repo.StatementImportMutation, catalog repo.ImportCatalog) error {
		var resolver *resolver
		var items []planned
		var err error
		report, items, resolver, err = makePlan(entries, input, catalog)
		if err != nil {
			return err
		}
		if report.Errors > 0 {
			return invalid("resolve reported errors before applying")
		}
		ids := make(map[int64]int64)
		resolve := func(id int64) int64 {
			if id < 0 {
				return ids[id]
			}
			return id
		}
		for _, c := range resolver.categories {
			id := c.ID
			c.ID = 0
			c.ParentID = resolve(c.ParentID)
			ids[id], err = tx.CreateCategory(ctx, c)
			if err != nil {
				return err
			}
		}
		for _, a := range resolver.assets {
			id := a.ID
			a.ID = 0
			a.ParentID = resolve(a.ParentID)
			ids[id], err = tx.CreateAsset(ctx, a)
			if err != nil {
				return err
			}
		}
		for i := range report.Creations {
			report.Creations[i].ID = resolve(report.Creations[i].ID)
			report.Creations[i].ParentID = resolve(report.Creations[i].ParentID)
		}
		for _, p := range items {
			e := p.entry
			balanceApplied := false
			key, fingerprint := importKey(e.Source, e.ID), e.Fingerprint
			row := model.Statement{UserID: input.UserID, AccountBookID: input.AccountBookID, CategoryID: resolve(p.category), AssetID: resolve(p.asset), Type: e.Kind, Amount: float64(e.Cents) / 100, Description: description(e), TargetObject: truncate(e.Counterparty, 255), CreatedAt: e.Occurred, Year: e.Occurred.Year(), Month: int(e.Occurred.Month()), Day: e.Occurred.Day(), TimeText: e.Occurred.Format("15:04:05"), BalanceApplied: &balanceApplied, ImportKey: &key, ImportFingerprint: &fingerprint}
			_, err := tx.CreateStatement(ctx, row, repo.BalanceEffect{})
			if err != nil {
				return err
			}
			report.Rows[p.result].CategoryID, report.Rows[p.result].AssetID = row.CategoryID, row.AssetID
		}
		return nil
	})
	if err != nil {
		return report, storageError(err)
	}
	report.Applied, report.Imported = true, report.Ready
	for i := range report.Rows {
		if report.Rows[i].Status == "ready" {
			report.Rows[i].Status = "imported"
		}
	}
	return report, nil
}

func makePlan(entries []entry, input Input, catalog repo.ImportCatalog) (Report, []planned, *resolver, error) {
	report := Report{Rows: make([]RowResult, 0, len(entries)), Creations: []Creation{}}
	r, err := newResolver(catalog, input.UserID, input.AccountBookID)
	if err != nil {
		return report, nil, nil, err
	}
	known := make(map[string]string)
	for _, row := range catalog.Statements {
		if row.AccountBookID == input.AccountBookID && row.UserID == input.UserID && row.ImportKey != nil {
			fingerprint := ""
			if row.ImportFingerprint != nil {
				fingerprint = *row.ImportFingerprint
			}
			known[*row.ImportKey] = fingerprint
		}
	}
	var items []planned
	for _, e := range entries {
		result := RowResult{File: e.File, Row: e.Row, Source: e.Source}
		key := importKey(e.Source, e.ID)
		if fingerprint, ok := known[key]; ok {
			if fingerprint == e.Fingerprint {
				result.Status = "duplicate"
				report.Duplicate++
			} else {
				result.Status, result.Reason = "error", "source_order_changed"
				report.Errors++
			}
		} else if e.Skip != "" {
			result.Status, result.Reason = "skipped", e.Skip
			report.Skipped++
		} else if e.Invalid != "" {
			result.Status, result.Reason = "error", e.Invalid
			report.Errors++
		} else {
			category, err := r.category(categoryTarget(e, input.Mapping), e.Kind)
			var asset int64
			if err == nil {
				target, kind := assetTarget(e, input.Mapping)
				asset, err = r.asset(target, kind)
			}
			if err != nil {
				result.Status, result.Reason = "error", mappingError(err)
				report.Errors++
			} else {
				result.Status, result.CategoryID, result.AssetID = "ready", category, asset
				items = append(items, planned{entry: e, category: category, asset: asset, result: len(report.Rows)})
				report.Ready++
				known[key] = e.Fingerprint
			}
		}
		report.Rows = append(report.Rows, result)
	}
	for _, c := range r.categories {
		report.Creations = append(report.Creations, Creation{ID: c.ID, Kind: "category", Name: c.Name, Type: c.Type, ParentID: c.ParentID})
	}
	for _, a := range r.assets {
		report.Creations = append(report.Creations, Creation{ID: a.ID, Kind: "asset", Name: a.Name, Type: a.Type, ParentID: a.ParentID})
	}
	return report, items, r, nil
}

func storageError(err error) error {
	if errors.Is(err, ErrInvalid) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, repo.ErrAccountBookNotFound) {
		return invalid("account book not found or access denied")
	}
	return fmt.Errorf("%w: %w", ErrStorage, err)
}
func description(e entry) string {
	parts := []string{e.Product}
	if e.Note != "" && e.Note != "/" && e.Note != "-" {
		parts = append(parts, e.Note)
	}
	parts = append(parts, "["+e.Source+" 历史补录]")
	return strings.Join(parts, "\n")
}
func truncate(s string, length int) string {
	chars := []rune(s)
	if len(chars) > length {
		return string(chars[:length])
	}
	return s
}
