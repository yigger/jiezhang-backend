package repository

import (
	"context"
	"errors"
)

var ErrFinanceAssetNotFound = errors.New("finance asset not found")

type FinanceRepository interface {
	ListAssets(ctx context.Context, accountBookID int64) ([]FinanceAssetRecord, error)
	FindAssetByID(ctx context.Context, assetID int64, accountBookID int64) (FinanceAssetRecord, error)

	SumStatementAmountByTypes(ctx context.Context, accountBookID int64, statementTypes []string) (float64, error)
	ListSpecialCategoryByTypes(ctx context.Context, accountBookID int64, statementTypes []string) ([]FinanceSpecialCategoryRecord, error)
	ListStatementSumsByTypes(ctx context.Context, accountBookID int64, statementTypes []string) ([]FinanceStatementTypeSumRecord, error)

	SumIncomeExpendByAsset(ctx context.Context, accountBookID int64, assetID int64) (FinanceIncomeExpendSumRecord, error)
	ListAssetTimeline(ctx context.Context, accountBookID int64, assetID int64) ([]FinanceTimelineRecord, error)
}

type FinanceAssetRecord struct {
	ID       int64
	Name     string
	Amount   float64
	ParentID int64
	IconPath string
	Type     string
}

type FinanceSpecialCategoryRecord struct {
	SpecialType string
	ID          int64
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
