package repo

import (
	"context"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

type SuperStatementFilter struct {
	AccountBookID int64
	Year          *int
	Month         *int
	AssetParentID *int64
	AssetID       *int64
	CategoryID    *int64
	OrderBy       string
}

type SuperStatementMonthSummaryRecord struct {
	Year         int
	Month        int
	ExpendAmount float64
	IncomeAmount float64
	Surplus      float64
}

type SuperStatementOverviewRecord struct {
	Expend float64
	Income float64
	Left   float64
}

type SuperStatementRepository interface {
	BatchLookup
	ListSimpleRows(ctx context.Context, filter SuperStatementFilter) ([]tablemodel.Statement, error)
	ListMonthSummaries(ctx context.Context, filter SuperStatementFilter) ([]SuperStatementMonthSummaryRecord, error)
	GetOverview(ctx context.Context, filter SuperStatementFilter) (SuperStatementOverviewRecord, error)
}

type SuperChartMonthSummary struct {
	Expend float64
	Income float64
}

type SuperChartDaySummary struct {
	Day    int
	Expend float64
	Income float64
}

type SuperChartPieParentItem struct {
	ParentID   int64
	ParentName string
	Data       float64
}

type SuperChartCategoryTopItem struct {
	CategoryID int64
	Name       string
	Data       float64
}

type SuperChartYearMonth struct {
	Year  int
	Month int
}

type SuperChartRepository interface {
	GetMonthSummary(ctx context.Context, accountBookID int64, year int, month int, includeRepayment bool) (SuperChartMonthSummary, error)
	ListDaySummaries(ctx context.Context, accountBookID int64, year int, month int) ([]SuperChartDaySummary, error)
	ListPieParents(ctx context.Context, accountBookID int64, year int, month int, statementType string) ([]SuperChartPieParentItem, error)
	ListCategoryTop(ctx context.Context, accountBookID int64, year int, month int) ([]SuperChartCategoryTopItem, error)
	ListYearMonths(ctx context.Context, accountBookID int64) ([]SuperChartYearMonth, error)
	SumExpendBetween(ctx context.Context, accountBookID int64, start time.Time, end time.Time) (float64, error)
}
