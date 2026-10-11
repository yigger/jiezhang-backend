package insights

import (
	"encoding/json"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
	"testing"
	"time"
)

func ptr(v int64) *int64 { return &v }
func TestWorkspaceLifetimeProjectPaymentAndBurdenReconcile(t *testing.T) {
	rows := []repo.InsightStatementRecord{entry(1, "2025-01-01", "expend", 100, "店", 1), entry(2, "2026-01-01", "expend", 50, "店", 2), entry(3, "2026-02-01", "income", 20, "", 2), entry(4, "2026-03-01", "expend", 10, "", 1)}
	split, _ := json.Marshal([]types.InsightAllocation{{MemberID: 1, AmountCents: 4000}, {MemberID: 2, AmountCents: 6000}})
	badSplit, _ := json.Marshal([]types.InsightAllocation{{MemberID: 1, AmountCents: 100}})
	annotations := []model.InsightStatementAnnotation{{StatementID: 1, ProjectID: ptr(7), PayerID: ptr(2), Allocations: split}, {StatementID: 2, ProjectID: ptr(7), PayerID: ptr(1), Allocations: badSplit}, {StatementID: 3, ProjectID: ptr(7)}}
	r := assembleWorkspace(model.AccountBook{ID: 8, UserID: 1}, 1, []model.InsightProject{{ID: 7, Name: "旅行", BudgetCents: 20000, CreatorID: 2}}, []model.InsightFixedCost{{ID: 9, Name: "订阅", AmountCents: 3000, IntervalMonths: 3, Active: true}}, annotations, nil, []model.AccountBookCollaborator{{UserID: 2, Remark: "小明"}}, rows)
	if len(r.Projects) != 1 || r.Projects[0].Summary.ExpendCents != 15000 || r.Projects[0].Summary.IncomeCents != 2000 || !r.Projects[0].CanEdit {
		t.Fatalf("project: %+v", r.Projects)
	}
	var paid, burden int64
	for _, p := range r.Payers {
		paid += p.ExpendCents
	}
	for _, p := range r.Burdens {
		burden += p.ExpendCents
	}
	if paid+r.UnknownPayerCents != 16000 || burden+r.UnallocatedCents != 16000 || r.InvalidSplitCount != 1 {
		t.Fatalf("reconciliation: %+v", r)
	}
	if r.MonthlyFixedCents != 1000 || r.AnnualFixedCents != 12000 {
		t.Fatal("quarterly estimate incorrect")
	}
	if len(r.Annotations) != 3 || r.Annotations[1].SplitValid {
		t.Fatal("stale split not flagged")
	}
}
func TestPortfolioChangesExplainNetDeltaAndMissingHistory(t *testing.T) {
	encode := func(a []types.InsightAssetBalance) json.RawMessage { b, _ := json.Marshal(a); return b }
	snapshots := []model.InsightPortfolioSnapshot{{ID: 1, CreatedAt: time.Now(), AssetsCents: 10000, LiabilitiesCents: 2000, Balances: encode([]types.InsightAssetBalance{{ID: 1, Name: "余额", Type: "deposit", AmountCents: 10000}, {ID: 2, Name: "负债", Type: "debt", AmountCents: 2000}})}, {ID: 2, CreatedAt: time.Now(), AssetsCents: 14000, LiabilitiesCents: 1000, Balances: encode([]types.InsightAssetBalance{{ID: 1, Name: "余额", Type: "deposit", AmountCents: 12000}, {ID: 2, Name: "负债", Type: "debt", AmountCents: 1000}, {ID: 3, Name: "新增", Type: "deposit", AmountCents: 2000}})}}
	points := portfolioPoints(snapshots)
	if len(points) != 2 || points[0].DeltaCents != nil || *points[1].DeltaCents != 5000 {
		t.Fatal("missing history synthesized or delta incorrect")
	}
	var delta int64
	for _, c := range points[1].Changes {
		delta += c.NetDeltaCents
	}
	if delta != *points[1].DeltaCents {
		t.Fatal("asset changes do not reconcile")
	}
	if points[1].Changes[2].BeforeCents != nil {
		t.Fatal("newly tracked asset must not invent a prior zero balance")
	}
	if len(portfolioPoints(nil)) != 0 {
		t.Fatal("empty history must remain empty")
	}
}
