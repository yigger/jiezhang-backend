package repo

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrBudgetCategoryNotFound = errors.New("budget category not found")

type BudgetRepository interface {
	GetAccountBookBudget(ctx context.Context, accountBookID int64) (float64, error)
	SumExpendByMonth(ctx context.Context, accountBookID int64, year int, month int) (float64, error)

	ListExpendParentCategories(ctx context.Context, accountBookID int64) ([]tablemodel.Category, error)
	ListChildCategoriesByParentID(ctx context.Context, accountBookID int64, parentID int64) ([]tablemodel.Category, error)
	ListChildCategoryIDsByParentID(ctx context.Context, accountBookID int64, parentID int64) ([]int64, error)
	FindCategoryByID(ctx context.Context, accountBookID int64, categoryID int64) (tablemodel.Category, error)
	SumStatementsByCategoryIDsAndMonth(ctx context.Context, accountBookID int64, categoryIDs []int64, year int, month int) (float64, error)
	SumParentCategoryBudget(ctx context.Context, accountBookID int64) (float64, error)

	UpdateAccountBookBudget(ctx context.Context, accountBookID int64, amount float64) error
	UpdateCategoryBudget(ctx context.Context, accountBookID int64, categoryID int64, amount float64) error
}
