package repo

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrCategoryNotFound = errors.New("category not found")

type CategoryRepository interface {
	BatchGetAssets(context.Context, []int64) ([]tablemodel.Asset, error)
	ListByIDs(context.Context, []int64) ([]tablemodel.Category, error)
	ListParents(ctx context.Context, accountBookID int64, categoryType string) ([]tablemodel.Category, error)
	ListChildrenByParentIDs(ctx context.Context, accountBookID int64, categoryType string, parentIDs []int64) ([]tablemodel.Category, error)
	ListFrequentChildren(ctx context.Context, accountBookID int64, categoryType string, limit int) ([]tablemodel.Category, error)
	ListGuessedFrequentByStatementType(ctx context.Context, filter CategoryGuessFilter) ([]tablemodel.Category, error)

	ListByParent(ctx context.Context, accountBookID int64, categoryType string, parentID int64) ([]tablemodel.Category, error)
	FindByID(ctx context.Context, accountBookID int64, id int64) (tablemodel.Category, error)
	ListStatementAmountByCategoryIDs(ctx context.Context, accountBookID int64, categoryIDs []int64) ([]CategoryAmountRecord, error)
	ListStatementAmountByParentIDs(ctx context.Context, accountBookID int64, parentIDs []int64) ([]CategoryAmountRecord, error)
	ListStatementsByCategory(ctx context.Context, accountBookID int64, categoryID int64) ([]tablemodel.Statement, error)
	SumStatements(ctx context.Context, accountBookID int64, statementType string, categoryIDs []int64, year int, month int) (float64, error)
	CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error)
	FindBySpecialType(ctx context.Context, specialType string) (int64, error)
	Create(ctx context.Context, input tablemodel.Category) (int64, error)
	UpdateByID(ctx context.Context, id int64, accountBookID int64, input tablemodel.Category) error
	DeleteByID(ctx context.Context, id int64, accountBookID int64) error
}

type CategoryGuessFilter struct {
	AccountBookID int64
	StatementType string
	Now           time.Time
	Limit         int
}

type CategoryAmountRecord struct {
	CategoryID int64
	Amount     float64
}
