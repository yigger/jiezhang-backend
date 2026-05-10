package repository

import (
	"context"
)

type CategoryListForStatement struct {
	ID       int64                      `json:"id"`
	Name     string                     `json:"name"`
	IconPath string                     `json:"iconPath"`
	Childs   []CategoryListForStatement `json:"childs"`
}

type CategoryRepository interface {
	List(ctx context.Context, filter CategoryListFilter) ([]CategoryListForStatement, error)
}

type CategoryListFilter struct {
	AccountBookID int64
	Type          string
}
