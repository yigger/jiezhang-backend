package repository

import (
	"context"
	"errors"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
)

var ErrAccountBookNotFound = errors.New("account book not found")

type AccountBookRepository interface {
	// Used by auth middleware.
	FindByID(ctx context.Context, id int64, userID int64) (domain.AccountBook, error)

	ListAccessible(ctx context.Context, userID int64) ([]AccountBookRecord, error)
	FindAccessibleByID(ctx context.Context, id int64, userID int64) (AccountBookRecord, error)

	Create(ctx context.Context, input AccountBookCreateInput) (AccountBookRecord, error)
	UpdateByID(ctx context.Context, id int64, input AccountBookUpdateInput) error
	DeleteByID(ctx context.Context, id int64) error

	SwitchDefaultByUserID(ctx context.Context, userID int64, accountBookID int64) error
}

type AccountBookRecord struct {
	ID          int64
	UserID      int64
	AccountType int
	Name        string
	Description string
	Budget      float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
