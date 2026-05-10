package mysql

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) (*CategoryRepository, error) {
	return &CategoryRepository{db: db}, nil
}

func (r *CategoryRepository) List(ctx context.Context, filter repository.CategoryListFilter) ([]repository.CategoryListForStatement, error) {
	var rows []repository.CategoryListForStatement

	query := r.db.WithContext(ctx).
		Table("categories c")

	if filter.AccountBookID > 0 {
		query = query.Where("c.account_book_id = ?", filter.AccountBookID)
	}
	if filter.Type != "" {
		query = query.Where("c.type = ?", filter.Type)
	}
	err := query.Scan(&rows).Error

	items := make([]repository.CategoryListForStatement, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.CategoryListForStatement{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: row.IconPath,
			Childs:   []repository.CategoryListForStatement{},
		})
	}

	return items, err
}
