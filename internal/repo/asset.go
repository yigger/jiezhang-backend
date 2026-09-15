package repo

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrAssetNotFound = errors.New("asset not found")

type AssetRepository interface {
	ListByIDs(context.Context, []int64) ([]tablemodel.Asset, error)
	ListParents(ctx context.Context, accountBookID int64) ([]tablemodel.Asset, error)
	ListChildrenByParentIDs(ctx context.Context, accountBookID int64, parentIDs []int64) ([]tablemodel.Asset, error)
	ListFrequentChildren(ctx context.Context, accountBookID int64, limit int) ([]tablemodel.Asset, error)
	ListGuessedFrequentByStatementTime(ctx context.Context, filter AssetGuessFilter) ([]tablemodel.Asset, error)

	ListByParent(ctx context.Context, accountBookID int64, parentID int64) ([]tablemodel.Asset, error)
	FindByID(ctx context.Context, accountBookID int64, id int64) (tablemodel.Asset, error)
	CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error)
	Create(ctx context.Context, input tablemodel.Asset) (int64, error)
	UpdateByID(ctx context.Context, id int64, accountBookID int64, input tablemodel.Asset) error
	DeleteByID(ctx context.Context, id int64, accountBookID int64) error
	UpdateAmountByID(ctx context.Context, id int64, accountBookID int64, amount float64) error
}

type AssetGuessFilter struct {
	AccountBookID int64
	Now           time.Time
	Limit         int
}
