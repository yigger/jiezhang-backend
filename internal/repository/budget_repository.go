package repository

import (
	"context"
	"errors"
)

var ErrBudgetCategoryNotFound = errors.New("budget category not found")

type BudgetRepository interface {
	GetAccountBookBudget(ctx context.Context, accountBookID int64) (float64, error)
	SumExpendByMonth(ctx context.Context, accountBookID int64, year int, month int) (float64, error)

	ListExpendParentCategories(ctx context.Context, accountBookID int64) ([]BudgetCategoryRecord, error)
	ListChildCategoriesByParentID(ctx context.Context, accountBookID int64, parentID int64) ([]BudgetCategoryRecord, error)
	ListChildCategoryIDsByParentID(ctx context.Context, accountBookID int64, parentID int64) ([]int64, error)
	FindCategoryByID(ctx context.Context, accountBookID int64, categoryID int64) (BudgetCategoryRecord, error)
	SumStatementsByCategoryIDsAndMonth(ctx context.Context, accountBookID int64, categoryIDs []int64, year int, month int) (float64, error)
	SumParentCategoryBudget(ctx context.Context, accountBookID int64) (float64, error)

	UpdateAccountBookBudget(ctx context.Context, accountBookID int64, amount float64) error
	UpdateCategoryBudget(ctx context.Context, accountBookID int64, categoryID int64, amount float64) error
}

type BudgetCategoryRecord struct {
	ID       int64
	Name     string
	IconPath string
	ParentID int64
	Budget   float64
	Type     string
}
