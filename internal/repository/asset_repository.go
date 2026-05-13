package repository

import (
	"context"
	"errors"
	"time"
)

var ErrAssetNotFound = errors.New("asset not found")

type AssetRepository interface {
	ListParents(ctx context.Context, filter AssetFilter) ([]AssetParentRecord, error)
	ListChildrenByParentIDs(ctx context.Context, filter AssetFilter, parentIDs []int64) ([]AssetChildRecord, error)
	ListFrequentChildren(ctx context.Context, filter AssetFilter, limit int) ([]AssetFrequentRecord, error)
	ListGuessedFrequentByStatementTime(ctx context.Context, filter AssetGuessFilter) ([]AssetFrequentRecord, error)

	ListByParent(ctx context.Context, accountBookID int64, parentID int64) ([]AssetManageRecord, error)
	FindByID(ctx context.Context, accountBookID int64, id int64) (AssetManageRecord, error)
	CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error)
	Create(ctx context.Context, input AssetWriteRecord) (int64, error)
	UpdateByID(ctx context.Context, id int64, accountBookID int64, input AssetWriteRecord) error
	DeleteByID(ctx context.Context, id int64, accountBookID int64) error
	UpdateAmountByID(ctx context.Context, id int64, accountBookID int64, amount float64) error
}

type AssetFilter struct {
	AccountBookID int64
	Type          string
}

type AssetParentRecord struct {
	ID       int64
	Name     string
	IconPath string
}

type AssetChildRecord struct {
	ID       int64
	Name     string
	IconPath string
	ParentID int64
}

type AssetFrequentRecord struct {
	ID         int64
	Name       string
	IconPath   string
	ParentID   int64
	ParentName string
	HasParent  bool
}

type AssetGuessFilter struct {
	AccountBookID int64
	Now           time.Time
	Limit         int
}

type AssetManageRecord struct {
	ID       int64
	Name     string
	Order    int
	IconPath string
	ParentID int64
	Type     string
	Amount   float64
	Remark   string
}

type AssetWriteRecord struct {
	CreatorID     int64
	AccountBookID int64
	Name          string
	Amount        float64
	ParentID      int64
	IconPath      string
	Remark        string
	Type          string
}
