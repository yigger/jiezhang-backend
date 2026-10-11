package insights

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"testing"
	"time"
)

type fakeRepository struct {
	rows       []repo.InsightStatementRecord
	err        error
	book       int64
	start, end time.Time
}

func (f *fakeRepository) ListRows(_ context.Context, book int64, start, end time.Time) ([]repo.InsightStatementRecord, error) {
	f.book, f.start, f.end = book, start, end
	return f.rows, f.err
}
func entry(id int64, date, kind string, amount float64, merchant string, member int64) repo.InsightStatementRecord {
	at, _ := time.ParseInLocation("2006-01-02", date, time.FixedZone("CST", 8*3600))
	return repo.InsightStatementRecord{Statement: model.Statement{ID: id, CreatedAt: at, Type: kind, Amount: amount, UserID: member, CategoryID: 1, AssetID: 2, Description: "会员"}, MerchantName: merchant, CategoryName: "订阅"}
}
func TestReportScopeAndReconciliation(t *testing.T) {
	f := &fakeRepository{rows: []repo.InsightStatementRecord{entry(1, "2026-01-01", "expend", 0.1, " Shop ", 1), entry(2, "2026-02-01", "expend", 0.2, "SHOP", 2), entry(3, "2026-03-01", "income", 10, "", 2), entry(4, "2026-10-07", "expend", 999, "", 1), entry(5, "2026-01-01", "transfer", 99, "", 1), entry(6, "2025-12-31", "expend", 99, "", 1)}}
	s := New(f)
	s.now = func() time.Time { return time.Date(2026, 10, 6, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)) }
	r, err := s.Report(context.Background(), 42, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if f.book != 42 || f.start.Format("2006-01-02") != "2026-01-01" || f.end.Format("2006-01-02") != "2026-10-07" {
		t.Fatalf("scope: %+v", f)
	}
	if r.ExpendCents != 30 || r.IncomeCents != 1000 || r.BalanceCents != 970 || r.Count != 3 {
		t.Fatalf("totals: %+v", r.InsightAmounts)
	}
	if len(r.Months) != 12 || r.Months[10].Started || r.Months[9].Complete {
		t.Fatal("month coverage/completion")
	}
	var spend, income int64
	for _, m := range r.Months {
		spend += m.ExpendCents
		income += m.IncomeCents
	}
	if spend != r.ExpendCents || income != r.IncomeCents {
		t.Fatal("months do not reconcile")
	}
	if len(r.Merchants) != 1 || r.Merchants[0].Count != 2 || r.Merchants[0].AverageCents != 15 {
		t.Fatalf("merchant aggregation: %+v", r.Merchants)
	}
	if len(r.Members) != 2 {
		t.Fatal("recording members were conflated")
	}
	var memberSpend int64
	for _, m := range r.Members {
		memberSpend += m.ExpendCents
	}
	if memberSpend != spend {
		t.Fatal("members do not reconcile")
	}
}
func TestRecurringUsesLatestAmountAndRejectsVariableCosts(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	end := start.AddDate(1, 0, 0)
	rows := []repo.InsightStatementRecord{entry(1, "2026-01-01", "expend", 100, "", 1), entry(2, "2026-02-01", "expend", 110, "", 1), entry(3, "2026-03-01", "expend", 105, "", 1)}
	r := build(rows, 2026, start, end, end)
	if len(r.RecurringCandidates) != 1 || r.RecurringCandidates[0].AmountCents != 10500 || r.RecurringCandidates[0].AnnualCents != 126000 {
		t.Fatalf("latest estimate: %+v", r.RecurringCandidates)
	}
	rows[0].Amount = 10
	r = build(rows, 2026, start, end, end)
	if len(r.RecurringCandidates) != 0 {
		t.Fatal("variable costs presented as fixed")
	}
}
func TestEmptyAndInvalidYear(t *testing.T) {
	f := &fakeRepository{}
	s := New(f)
	s.now = func() time.Time { return time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC) }
	r, err := s.Report(context.Background(), 1, 2025)
	if err != nil || r.Merchants == nil || r.Members == nil || r.LargeExpenses == nil || r.RecurringCandidates == nil {
		t.Fatal("empty collections must be arrays")
	}
	if _, err = s.Report(context.Background(), 1, 2027); !errors.Is(err, ErrInvalidYear) {
		t.Fatal("future year accepted")
	}
	f.err = errors.New("storage unavailable")
	if _, err = s.Report(context.Background(), 1, 2026); !errors.Is(err, f.err) {
		t.Fatal("storage error swallowed")
	}
}
