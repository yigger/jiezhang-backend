package repository

import (
	"context"
)

type StatementFrequentCategoryItem struct {
	ID       int64                       `json:"id"`
	Name     string                      `json:"name"`
	IconPath string                      `json:"icon_path"`
	Parent   StatementCategoryParentItem `json:"parent"`
}

type StatementCategoryParentItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementCategoryChildItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type StatementCategoryTreeItem struct {
	ID       int64                        `json:"id"`
	Name     string                       `json:"name"`
	IconPath string                       `json:"icon_path"`
	Childs   []StatementCategoryChildItem `json:"childs"`
}

type StatementCategoriesResult struct {
	Frequent   []StatementFrequentCategoryItem `json:"frequent"`
	Categories []StatementCategoryTreeItem     `json:"categories"`
}

type CategoryRepository interface {
	List(ctx context.Context, filter CategoryListFilter) (StatementCategoriesResult, error)
}

type CategoryListFilter struct {
	AccountBookID int64
	Type          string
}
