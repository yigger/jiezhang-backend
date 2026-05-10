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

type categoryParentRow struct {
	ID       int64  `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	IconPath string `gorm:"column:icon_path"`
}

type categoryChildRow struct {
	ID       int64  `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	IconPath string `gorm:"column:icon_path"`
	ParentID int64  `gorm:"column:parent_id"`
}

type frequentCategoryRow struct {
	ID         int64  `gorm:"column:id"`
	Name       string `gorm:"column:name"`
	IconPath   string `gorm:"column:icon_path"`
	ParentID   int64  `gorm:"column:parent_id"`
	ParentName string `gorm:"column:parent_name"`
}

func (r *CategoryRepository) List(ctx context.Context, filter repository.CategoryListFilter) (repository.StatementCategoriesResult, error) {
	baseParentQuery := r.db.WithContext(ctx).Table("categories c")
	baseChildQuery := r.db.WithContext(ctx).Table("categories c")
	baseFrequentQuery := r.db.WithContext(ctx).Table("categories c")

	if filter.AccountBookID > 0 {
		baseParentQuery = baseParentQuery.Where("c.account_book_id = ?", filter.AccountBookID)
		baseChildQuery = baseChildQuery.Where("c.account_book_id = ?", filter.AccountBookID)
		baseFrequentQuery = baseFrequentQuery.Where("c.account_book_id = ?", filter.AccountBookID)
	}
	if filter.Type != "" {
		baseParentQuery = baseParentQuery.Where("c.type = ?", filter.Type)
		baseChildQuery = baseChildQuery.Where("c.type = ?", filter.Type)
		baseFrequentQuery = baseFrequentQuery.Where("c.type = ?", filter.Type)
	}

	var parents []categoryParentRow
	if err := baseParentQuery.
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path").
		Where("c.parent_id = 0").
		Order("c.id ASC").
		Scan(&parents).Error; err != nil {
		return repository.StatementCategoriesResult{}, err
	}

	parentIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentIDs = append(parentIDs, p.ID)
	}

	childrenByParent := make(map[int64][]repository.StatementCategoryChildItem, len(parents))
	if len(parentIDs) > 0 {
		var childRows []categoryChildRow
		if err := baseChildQuery.
			Select("c.id AS id, c.name AS name, c.icon_path AS icon_path, c.parent_id AS parent_id").
			Where("c.parent_id IN ?", parentIDs).
			Order("c.id ASC").
			Scan(&childRows).Error; err != nil {
			return repository.StatementCategoriesResult{}, err
		}
		for _, row := range childRows {
			childrenByParent[row.ParentID] = append(childrenByParent[row.ParentID], repository.StatementCategoryChildItem{
				ID:       row.ID,
				Name:     row.Name,
				IconPath: row.IconPath,
			})
		}
	}

	categories := make([]repository.StatementCategoryTreeItem, 0, len(parents))
	for _, p := range parents {
		childs := childrenByParent[p.ID]
		if childs == nil {
			childs = []repository.StatementCategoryChildItem{}
		}
		categories = append(categories, repository.StatementCategoryTreeItem{
			ID:       p.ID,
			Name:     p.Name,
			IconPath: p.IconPath,
			Childs:   childs,
		})
	}

	var frequentRows []frequentCategoryRow
	if err := baseFrequentQuery.
		Joins("LEFT JOIN categories parent ON parent.id = c.parent_id").
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path, c.parent_id AS parent_id, COALESCE(parent.name, '') AS parent_name").
		Where("c.parent_id > 0").
		Where("c.frequent > 5").
		Order("c.frequent DESC").
		Limit(10).
		Scan(&frequentRows).Error; err != nil {
		return repository.StatementCategoriesResult{}, err
	}

	frequent := make([]repository.StatementFrequentCategoryItem, 0, len(frequentRows))
	for _, row := range frequentRows {
		frequent = append(frequent, repository.StatementFrequentCategoryItem{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: row.IconPath,
			Parent: repository.StatementCategoryParentItem{
				ID:   row.ParentID,
				Name: row.ParentName,
			},
		})
	}

	return repository.StatementCategoriesResult{
		Frequent:   frequent,
		Categories: categories,
	}, nil
}
