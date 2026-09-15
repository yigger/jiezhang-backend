package repo

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrAccountBookNotFound = errors.New("account book not found")

type AccountBookRepository interface {
	// Used by auth middleware.
	FindByID(ctx context.Context, id int64, userID int64) (tablemodel.AccountBook, error)

	ListAccessible(ctx context.Context, userID int64) ([]tablemodel.AccountBook, error)
	FindAccessibleByID(ctx context.Context, id int64, userID int64) (tablemodel.AccountBook, error)

	Create(ctx context.Context, input AccountBookCreateInput) (tablemodel.AccountBook, error)
	UpdateByID(ctx context.Context, id int64, input AccountBookUpdateInput) error
	DeleteByID(ctx context.Context, id int64) error

	SwitchDefaultByUserID(ctx context.Context, userID int64, accountBookID int64) error
}

type AccountBookUpdateInput struct {
	Name        string
	Description string
	AccountType int
}

type AccountBookCreateInput struct {
	UserID       int64
	UserNickname string
	Name         string
	Description  string
	AccountType  int
	Categories   map[string][]AccountBookCategoryTemplate
	Assets       []AccountBookAssetTemplate
}

type AccountBookCategoryTemplate struct {
	Name     string
	IconPath string
	Childs   []AccountBookCategoryChildTemplate
}

type AccountBookCategoryChildTemplate struct {
	Name     string
	IconPath string
}

type AccountBookAssetTemplate struct {
	Name     string
	IconPath string
	Type     string
	Childs   []AccountBookAssetChildTemplate
}

type AccountBookAssetChildTemplate struct {
	Name     string
	IconPath string
}
