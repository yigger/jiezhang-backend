package insights

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrMissingRecord = repo.ErrInsightRecordNotFound
var ErrInvalidInput = errors.New("分析数据无效")
var ErrForbidden = errors.New("没有权限修改此记录")

const maxCents int64 = 999999999999

type StorageService struct {
	rows    repo.InsightsRepository
	storage repo.InsightsStorage
	now     func() time.Time
}

func NewStorage(s repo.InsightsStorage, rows repo.InsightsRepository) *StorageService {
	return &StorageService{storage: s, rows: rows, now: time.Now}
}
func validName(s string) bool { return strings.TrimSpace(s) != "" && utf8.RuneCountInString(s) <= 100 }
func authorized(book model.AccountBook, members []model.AccountBookCollaborator, user int64) (bool, bool) {
	owner := book.UserID == user
	member := owner
	for _, m := range members {
		if m.UserID == user {
			member = true
			if m.Role == "owner" {
				owner = true
			}
		}
	}
	return member, owner
}

var projectIconPattern = regexp.MustCompile(`^jcon-[A-Za-z0-9-]{1,59}$`)
var projectColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func (s *StorageService) SaveProject(ctx context.Context, book, user int64, in types.InsightProjectInput) (model.InsightProject, error) {
	if !validName(in.Name) || in.BudgetCents < 0 || in.BudgetCents > maxCents || in.ID < 0 {
		return model.InsightProject{}, ErrInvalidInput
	}
	var result model.InsightProject
	err := s.storage.Transact(ctx, func(tx repo.InsightsStorageTx) error {
		b, e := tx.LockBook(book)
		if e != nil {
			return e
		}
		members, e := tx.Members(book)
		if e != nil {
			return e
		}
		member, owner := authorized(b, members, user)
		if !member {
			return ErrForbidden
		}
		now := s.now()
		result = model.InsightProject{AccountBookID: book, CreatorID: user, CreatedAt: now}
		if in.ID > 0 {
			result, e = tx.Project(book, in.ID)
			if e != nil {
				return e
			}
			if result.CreatorID != user && !owner {
				return ErrForbidden
			}
		}
		if in.Icon != nil {
			icon := strings.TrimSpace(*in.Icon)
			if icon != "" && !projectIconPattern.MatchString(icon) {
				return ErrInvalidInput
			}
			result.Icon = nil
			if icon != "" {
				result.Icon = &icon
			}
		}
		if in.Color != nil {
			color := strings.TrimSpace(*in.Color)
			if color != "" && !projectColorPattern.MatchString(color) {
				return ErrInvalidInput
			}
			result.Color = nil
			if color != "" {
				color = strings.ToUpper(color)
				result.Color = &color
			}
		}
		if in.ParticipantIDs != nil {
			valid := map[int64]bool{b.UserID: true}
			for _, m := range members {
				valid[m.UserID] = true
			}
			seen := map[int64]bool{}
			for _, id := range in.ParticipantIDs {
				if !valid[id] || seen[id] {
					return ErrInvalidInput
				}
				seen[id] = true
			}
			result.ParticipantIDs, e = json.Marshal(in.ParticipantIDs)
			if e != nil {
				return e
			}
		}
		if in.StartDate != nil {
			result.StartDate, e = parseProjectDate(*in.StartDate)
			if e != nil {
				return e
			}
		}
		if in.EndDate != nil {
			result.EndDate, e = parseProjectDate(*in.EndDate)
			if e != nil {
				return e
			}
		}
		if result.StartDate != nil && result.EndDate != nil && result.EndDate.Before(*result.StartDate) {
			return ErrInvalidInput
		}
		result.Name = strings.TrimSpace(in.Name)
		result.BudgetCents = in.BudgetCents
		result.Archived = in.Archived
		result.UpdatedAt = now
		return tx.SaveProject(&result)
	})
	return result, err
}
func (s *StorageService) SaveFixedCost(ctx context.Context, book, user int64, in types.InsightFixedCostInput) (model.InsightFixedCost, error) {
	if !validName(in.Name) || in.AmountCents <= 0 || in.AmountCents > maxCents || in.IntervalMonths < 1 || in.IntervalMonths > 12 || in.DueDay < 1 || in.DueDay > 31 || in.ID < 0 || in.CategoryID < 0 || in.AssetID < 0 || len(in.CandidateKey) > 255 {
		return model.InsightFixedCost{}, ErrInvalidInput
	}
	var result model.InsightFixedCost
	err := s.storage.Transact(ctx, func(tx repo.InsightsStorageTx) error {
		b, e := tx.LockBook(book)
		if e != nil {
			return e
		}
		members, e := tx.Members(book)
		if e != nil {
			return e
		}
		member, owner := authorized(b, members, user)
		if !member {
			return ErrForbidden
		}
		now := s.now()
		result = model.InsightFixedCost{AccountBookID: book, CreatorID: user, CreatedAt: now}
		if in.ID == 0 && in.CandidateKey != "" {
			existing, err := tx.FixedCosts(book)
			if err != nil {
				return err
			}
			for _, f := range existing {
				if f.CandidateKey == in.CandidateKey {
					result = f
					return nil
				}
			}
		}
		if in.ID > 0 {
			result, e = tx.FixedCost(book, in.ID)
			if e != nil {
				return e
			}
			if result.CreatorID != user && !owner {
				return ErrForbidden
			}
		}
		if in.CategoryID > 0 {
			category, err := tx.Category(book, in.CategoryID)
			if err != nil {
				return err
			}
			if category.Type != "expend" {
				return ErrInvalidInput
			}
		}
		if in.AssetID > 0 {
			assets, err := tx.Assets(book)
			if err != nil {
				return err
			}
			found := false
			for _, a := range assets {
				if a.ID == in.AssetID && a.ParentID != 0 {
					found = true
				}
			}
			if !found {
				return ErrInvalidInput
			}
		}
		if in.Active && (in.CategoryID <= 0 || in.AssetID <= 0) {
			return ErrInvalidInput
		}
		if in.Active && (!result.Active || result.NextRunDate == nil || result.IntervalMonths != in.IntervalMonths || result.DueDay != in.DueDay) {
			next := firstFixedCostDate(now, in.IntervalMonths, in.DueDay)
			if result.ID > 0 {
				if ledger, ok := tx.(repo.FixedCostLedgerTx); ok {
					for {
						exists, err := ledger.FixedCostRun(result.ID, next)
						if err != nil {
							return err
						}
						if !exists {
							break
						}
						next = nextFixedCostDate(next, in.IntervalMonths, in.DueDay)
					}
				}
			}
			result.NextRunDate = &next
		}
		result.Name = strings.TrimSpace(in.Name)
		result.AmountCents = in.AmountCents
		result.CategoryID = in.CategoryID
		result.AssetID = in.AssetID
		result.IntervalMonths = in.IntervalMonths
		result.DueDay = in.DueDay
		result.CandidateKey = in.CandidateKey
		result.Active = in.Active
		result.UpdatedAt = now
		return tx.SaveFixedCost(&result)
	})
	return result, err
}
func (s *StorageService) SaveAnnotation(ctx context.Context, book, user int64, in types.InsightAnnotationInput) error {
	if in.StatementID <= 0 {
		return ErrInvalidInput
	}
	return s.storage.Transact(ctx, func(tx repo.InsightsStorageTx) error {
		b, e := tx.LockBook(book)
		if e != nil {
			return e
		}
		members, e := tx.Members(book)
		if e != nil {
			return e
		}
		member, owner := authorized(b, members, user)
		if !member {
			return ErrForbidden
		}
		row, e := tx.LockStatement(book, in.StatementID)
		if e != nil {
			return e
		}
		if row.UserID != user && !owner {
			return ErrForbidden
		}
		if row.Type != "income" && row.Type != "expend" {
			return ErrInvalidInput
		}
		validMembers := map[int64]bool{b.UserID: true}
		for _, m := range members {
			validMembers[m.UserID] = true
		}
		if in.ConsumerID != nil && !validMembers[*in.ConsumerID] {
			return ErrInvalidInput
		}
		if in.PayerID != nil && !validMembers[*in.PayerID] {
			return ErrInvalidInput
		}
		if in.ProjectID != nil {
			p, e := tx.Project(book, *in.ProjectID)
			if e != nil {
				return e
			}
			if in.ConsumerID != nil {
				var participants []int64
				if len(p.ParticipantIDs) > 0 {
					if err := json.Unmarshal(p.ParticipantIDs, &participants); err != nil {
						return err
					}
				}
				if len(participants) > 0 {
					found := false
					for _, id := range participants {
						if id == *in.ConsumerID {
							found = true
						}
					}
					if !found {
						return ErrInvalidInput
					}
				}
			}
			if p.Archived {
				existing, found, err := tx.Annotation(book, row.ID)
				if err != nil {
					return err
				}
				if !found || existing.ProjectID == nil || *existing.ProjectID != p.ID {
					return ErrInvalidInput
				}
			}
		}
		if in.FixedCostID != nil {
			if _, e = tx.FixedCost(book, *in.FixedCostID); e != nil {
				return e
			}
		}
		seen := map[int64]bool{}
		var sum int64
		for _, a := range in.Allocations {
			if !validMembers[a.MemberID] || seen[a.MemberID] || a.AmountCents < 0 || a.AmountCents > maxCents {
				return ErrInvalidInput
			}
			seen[a.MemberID] = true
			sum += a.AmountCents
		}
		if len(in.Allocations) > 0 && (row.Type != "expend" || sum != cents(row.Amount)) {
			return ErrInvalidInput
		}
		data, e := json.Marshal(in.Allocations)
		if e != nil {
			return e
		}
		now := s.now()
		annotation := model.InsightStatementAnnotation{AccountBookID: book, CreatorID: user, StatementID: row.ID, ConsumerID: in.ConsumerID, ProjectID: in.ProjectID, PayerID: in.PayerID, FixedCostID: in.FixedCostID, Allocations: data, CreatedAt: now, UpdatedAt: now}
		return tx.SaveAnnotation(&annotation)
	})
}
func (s *StorageService) CapturePortfolio(ctx context.Context, book, user int64, note string) (model.InsightPortfolioSnapshot, error) {
	if utf8.RuneCountInString(note) > 255 {
		return model.InsightPortfolioSnapshot{}, ErrInvalidInput
	}
	var result model.InsightPortfolioSnapshot
	err := s.storage.Transact(ctx, func(tx repo.InsightsStorageTx) error {
		b, e := tx.LockBook(book)
		if e != nil {
			return e
		}
		members, e := tx.Members(book)
		if e != nil {
			return e
		}
		member, _ := authorized(b, members, user)
		if !member {
			return ErrForbidden
		}
		assets, e := tx.Assets(book)
		if e != nil {
			return e
		}
		balances := []types.InsightAssetBalance{}
		now := s.now()
		result = model.InsightPortfolioSnapshot{AccountBookID: book, CreatorID: user, Note: strings.TrimSpace(note), CreatedAt: now, UpdatedAt: now}
		for _, a := range assets {
			if a.Type != "deposit" && a.Type != "debt" {
				return ErrInvalidInput
			}
			if math.IsNaN(a.Amount) || math.IsInf(a.Amount, 0) || math.Abs(a.Amount*100) > float64(maxCents) {
				return ErrInvalidInput
			}
			amount := int64(math.Round(a.Amount * 100))
			balances = append(balances, types.InsightAssetBalance{ID: a.ID, Name: a.Name, Type: a.Type, AmountCents: amount})
			if a.Type == "deposit" {
				result.AssetsCents += amount
			} else {
				result.LiabilitiesCents += amount
			}
		}
		result.Balances, e = json.Marshal(balances)
		if e != nil {
			return e
		}
		return tx.CreateSnapshot(&result)
	})
	return result, err
}
func (s *StorageService) SaveMerchantAliases(ctx context.Context, book, user int64, in types.InsightMerchantAliasInput) error {
	if len(in.PayeeIDs) == 0 || len(in.PayeeIDs) > 100 || utf8.RuneCountInString(in.Name) > 100 {
		return ErrInvalidInput
	}
	name := strings.TrimSpace(in.Name)
	return s.storage.Transact(ctx, func(tx repo.InsightsStorageTx) error {
		b, e := tx.LockBook(book)
		if e != nil {
			return e
		}
		members, e := tx.Members(book)
		if e != nil {
			return e
		}
		member, _ := authorized(b, members, user)
		if !member {
			return ErrForbidden
		}
		seen := map[int64]bool{}
		for _, id := range in.PayeeIDs {
			if id <= 0 || seen[id] {
				return ErrInvalidInput
			}
			seen[id] = true
			if _, e = tx.Payee(book, id); e != nil {
				return e
			}
			now := s.now()
			row := model.InsightMerchantAlias{AccountBookID: book, CreatorID: user, PayeeID: id, Name: name, CreatedAt: now, UpdatedAt: now}
			if e = tx.SaveMerchantAlias(&row); e != nil {
				return e
			}
		}
		return nil
	})
}

func parseProjectDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, ErrInvalidInput
	}
	return &parsed, nil
}
