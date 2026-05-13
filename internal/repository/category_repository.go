package repository

import (
	"context"
	"errors"
	"time"
)

var ErrCategoryNotFound = errors.New("category not found")

type CategoryRepository interface {
	ListParents(ctx context.Context, filter CategoryListFilter) ([]CategoryParentRecord, error)
	ListChildrenByParentIDs(ctx context.Context, filter CategoryListFilter, parentIDs []int64) ([]CategoryChildRecord, error)
	ListFrequentChildren(ctx context.Context, filter CategoryListFilter, limit int) ([]CategoryFrequentRecord, error)
	ListGuessedFrequentByStatementType(ctx context.Context, filter CategoryGuessFilter) ([]CategoryFrequentRecord, error)

	ListByParent(ctx context.Context, filter CategoryListFilter, parentID int64) ([]CategoryManageRecord, error)
	FindByID(ctx context.Context, accountBookID int64, id int64) (CategoryManageRecord, error)
	ListStatementAmountByCategoryIDs(ctx context.Context, accountBookID int64, categoryIDs []int64) ([]CategoryAmountRecord, error)
	ListStatementAmountByParentIDs(ctx context.Context, accountBookID int64, parentIDs []int64) ([]CategoryAmountRecord, error)
	ListStatementsByCategory(ctx context.Context, accountBookID int64, categoryID int64) ([]CategoryStatementRecord, error)
	SumStatements(ctx context.Context, accountBookID int64, statementType string, categoryIDs []int64, year int, month int) (float64, error)
	CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error)
	Create(ctx context.Context, input CategoryWriteRecord) (int64, error)
	UpdateByID(ctx context.Context, id int64, accountBookID int64, input CategoryWriteRecord) error
	DeleteByID(ctx context.Context, id int64, accountBookID int64) error
}

type CategoryListFilter struct {
	AccountBookID int64
	Type          string
}

type CategoryParentRecord struct {
	ID       int64
	Name     string
	IconPath string
}

type CategoryChildRecord struct {
	ID       int64
	Name     string
	IconPath string
	ParentID int64
}

type CategoryFrequentRecord struct {
	ID         int64
	Name       string
	IconPath   string
	ParentID   int64
	ParentName string
	HasParent  bool
}

type CategoryGuessFilter struct {
	AccountBookID int64
	StatementType string
	Now           time.Time
	Limit         int
}

type CategoryManageRecord struct {
	ID       int64
	Name     string
	Order    int
	IconPath string
	ParentID int64
	Type     string
}

type CategoryAmountRecord struct {
	CategoryID int64
	Amount     float64
}

type CategoryStatementRecord struct {
	ID           int64
	Day          int
	Year         int
	Month        int
	Type         string
	CategoryName string
	IconPath     string
	Description  string
	Amount       float64
	CreatedAt    time.Time
	AssetName    string
}

type CategoryWriteRecord struct {
	UserID        int64
	AccountBookID int64
	Name          string
	ParentID      int64
	IconPath      string
	Type          string
}
