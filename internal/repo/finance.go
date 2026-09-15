package repo

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrFinanceAssetNotFound = errors.New("finance asset not found")

type FinanceRepository interface {
	ListAssets(ctx context.Context, accountBookID int64) ([]tablemodel.Asset, error)
	FindAssetByID(ctx context.Context, assetID int64, accountBookID int64) (tablemodel.Asset, error)

	SumStatementAmountByTypes(ctx context.Context, accountBookID int64, statementTypes []string) (float64, error)
	ListSpecialCategoryByTypes(ctx context.Context, statementTypes []string) ([]tablemodel.Category, error)
	ListStatementSumsByTypes(ctx context.Context, accountBookID int64, statementTypes []string) ([]FinanceStatementTypeSumRecord, error)

	SumIncomeExpendByAsset(ctx context.Context, accountBookID int64, assetID int64) (FinanceIncomeExpendSumRecord, error)
	ListAssetTimeline(ctx context.Context, accountBookID int64, assetID int64) ([]FinanceTimelineRecord, error)
}

type FinanceStatementTypeSumRecord struct {
	StatementType string
	Amount        float64
}

type FinanceIncomeExpendSumRecord struct {
	Income float64
	Expend float64
}

type FinanceTimelineRecord struct {
	Year         int
	Month        int
	IncomeAmount float64
	ExpendAmount float64
}
