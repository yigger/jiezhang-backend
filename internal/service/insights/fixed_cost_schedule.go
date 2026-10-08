package insights

import (
	"context"
	"errors"
	"fmt"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"time"
)

var fixedCostZone = time.FixedZone("Asia/Shanghai", 8*60*60)

func fixedCostDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}
func fixedCostDue(month time.Time, day int) time.Time {
	last := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
	if day > last {
		day = last
	}
	return time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, time.Local)
}
func firstFixedCostDate(now time.Time, interval, day int) time.Time {
	today := fixedCostDate(now.In(fixedCostZone))
	due := fixedCostDue(today, day)
	if due.Before(today) {
		due = fixedCostDue(time.Date(today.Year(), today.Month()+time.Month(interval), 1, 0, 0, 0, 0, time.Local), day)
	}
	return due
}
func nextFixedCostDate(previous time.Time, interval, day int) time.Time {
	return fixedCostDue(time.Date(previous.Year(), previous.Month()+time.Month(interval), 1, 0, 0, 0, 0, time.Local), day)
}

// RunDueFixedCosts catches up only confirmed schedules. Each occurrence commits
// its statement, wallet delta, annotation, execution identity and next date together.
func (s *StorageService) RunDueFixedCosts(ctx context.Context, through time.Time) (int, error) {
	through = fixedCostDate(through)
	rules, err := s.storage.ListScheduledFixedCosts(ctx, through)
	if err != nil {
		return 0, err
	}
	count := 0
	var failures []error
	for _, candidate := range rules {
		for {
			if err := ctx.Err(); err != nil {
				return count, errors.Join(append(failures, err)...)
			}
			advanced, created := false, false
			err = s.storage.Transact(ctx, func(tx repo.InsightsStorageTx) error {
				book, e := tx.LockBook(candidate.AccountBookID)
				if e != nil {
					return e
				}
				rule, e := tx.FixedCost(candidate.AccountBookID, candidate.ID)
				if e != nil {
					return e
				}
				if !rule.Active || rule.NextRunDate == nil || fixedCostDate(*rule.NextRunDate).After(through) {
					return nil
				}
				if rule.IntervalMonths < 1 || rule.IntervalMonths > 12 || rule.DueDay < 1 || rule.DueDay > 31 || rule.AmountCents <= 0 || rule.AmountCents > maxCents || rule.CategoryID <= 0 || rule.AssetID <= 0 {
					return ErrInvalidInput
				}
				ledger, ok := tx.(repo.FixedCostLedgerTx)
				if !ok {
					return errors.New("recurring ledger unavailable")
				}
				due := fixedCostDate(*rule.NextRunDate)
				next := nextFixedCostDate(due, rule.IntervalMonths, rule.DueDay)
				exists, e := ledger.FixedCostRun(rule.ID, due)
				if e != nil {
					return e
				}
				if !exists {
					members, e := tx.Members(book.ID)
					if e != nil {
						return e
					}
					member, _ := authorized(book, members, rule.CreatorID)
					if !member {
						return ErrForbidden
					}
					category, e := tx.Category(book.ID, rule.CategoryID)
					if e != nil {
						return e
					}
					if category.Type != "expend" {
						return ErrInvalidInput
					}
					assets, e := tx.Assets(book.ID)
					if e != nil {
						return e
					}
					validAsset := false
					for _, a := range assets {
						if a.ID == rule.AssetID && a.ParentID != 0 {
							validAsset = true
						}
					}
					if !validAsset {
						return ErrInvalidInput
					}
					amount := float64(rule.AmountCents) / 100
					occurred := time.Date(due.Year(), due.Month(), due.Day(), 23, 55, 0, 0, time.Local)
					now := s.now()
					row := model.Statement{UserID: rule.CreatorID, AccountBookID: book.ID, CategoryID: rule.CategoryID, AssetID: rule.AssetID, Type: "expend", Amount: amount, Description: "固定开销 · " + rule.Name, CreatedAt: occurred, Year: due.Year(), Month: int(due.Month()), Day: due.Day(), TimeText: "23:55"}
					id, e := ledger.CreateFixedCostStatement(ctx, row, repo.BalanceEffect{Source: -amount})
					if e != nil {
						return e
					}
					annotation := model.InsightStatementAnnotation{AccountBookID: book.ID, CreatorID: rule.CreatorID, StatementID: id, FixedCostID: &rule.ID, CreatedAt: now, UpdatedAt: now}
					if e = tx.SaveAnnotation(&annotation); e != nil {
						return e
					}
					if e = ledger.SaveFixedCostRun(&model.InsightFixedCostRun{FixedCostID: rule.ID, DueDate: due, StatementID: id, CreatedAt: now}); e != nil {
						return e
					}
					created = true
				}
				rule.NextRunDate = &next
				rule.UpdatedAt = s.now()
				if e = tx.SaveFixedCost(&rule); e != nil {
					return e
				}
				advanced = true
				return nil
			})
			if err != nil {
				failures = append(failures, fmt.Errorf("fixed-cost rule %d: %w", candidate.ID, err))
				break
			}
			if created {
				count++
			}
			if !advanced {
				break
			}
		}
	}
	return count, errors.Join(failures...)
}
