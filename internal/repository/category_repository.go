package repository

import (
	"context"
)

type CategoryRepository interface {
	ListParents(ctx context.Context, filter CategoryListFilter) ([]CategoryParentRecord, error)
	ListChildrenByParentIDs(ctx context.Context, filter CategoryListFilter, parentIDs []int64) ([]CategoryChildRecord, error)
	ListFrequentChildren(ctx context.Context, filter CategoryListFilter, limit int) ([]CategoryFrequentRecord, error)
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
}
