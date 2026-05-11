package repository

import (
	"context"
	"time"
)

type AssetRepository interface {
	ListParents(ctx context.Context, filter AssetFilter) ([]AssetParentRecord, error)
	ListChildrenByParentIDs(ctx context.Context, filter AssetFilter, parentIDs []int64) ([]AssetChildRecord, error)
	ListFrequentChildren(ctx context.Context, filter AssetFilter, limit int) ([]AssetFrequentRecord, error)
	ListGuessedFrequentByStatementTime(ctx context.Context, filter AssetGuessFilter) ([]AssetFrequentRecord, error)
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
