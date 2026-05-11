package mysql

import (
	"context"
	"database/sql"
	"time"

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
	ID         int64          `gorm:"column:id"`
	Name       string         `gorm:"column:name"`
	IconPath   string         `gorm:"column:icon_path"`
	ParentID   sql.NullInt64  `gorm:"column:parent_id"`
	ParentName sql.NullString `gorm:"column:parent_name"`
}

func (r *CategoryRepository) ListParents(ctx context.Context, filter repository.CategoryListFilter) ([]repository.CategoryParentRecord, error) {
	query := r.buildCategoryBaseQuery(ctx, filter)
	var parents []categoryParentRow
	if err := query.
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path").
		Where("c.parent_id = 0").
		Order("c.id ASC").
		Scan(&parents).Error; err != nil {
		return nil, err
	}
	records := make([]repository.CategoryParentRecord, 0, len(parents))
	for _, p := range parents {
		records = append(records, repository.CategoryParentRecord{
			ID:       p.ID,
			Name:     p.Name,
			IconPath: p.IconPath,
		})
	}
	return records, nil
}

func (r *CategoryRepository) ListChildrenByParentIDs(ctx context.Context, filter repository.CategoryListFilter, parentIDs []int64) ([]repository.CategoryChildRecord, error) {
	if len(parentIDs) == 0 {
		return []repository.CategoryChildRecord{}, nil
	}
	query := r.buildCategoryBaseQuery(ctx, filter)
	var rows []categoryChildRow
	if err := query.
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path, c.parent_id AS parent_id").
		Where("c.parent_id IN ?", parentIDs).
		Order("c.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	records := make([]repository.CategoryChildRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, repository.CategoryChildRecord{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: row.IconPath,
			ParentID: row.ParentID,
		})
	}
	return records, nil
}

func (r *CategoryRepository) ListFrequentChildren(ctx context.Context, filter repository.CategoryListFilter, limit int) ([]repository.CategoryFrequentRecord, error) {
	if limit <= 0 {
		limit = 10
	}
	query := r.buildCategoryBaseQuery(ctx, filter)
	var frequentRows []frequentCategoryRow
	if err := query.
		Joins("LEFT JOIN categories parent ON parent.id = c.parent_id").
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path, c.parent_id AS parent_id, COALESCE(parent.name, '') AS parent_name").
		Where("c.parent_id > 0").
		Where("c.frequent > 5").
		Order("c.frequent DESC").
		Limit(limit).
		Scan(&frequentRows).Error; err != nil {
		return nil, err
	}
	records := make([]repository.CategoryFrequentRecord, 0, len(frequentRows))
	for _, row := range frequentRows {
		records = append(records, repository.CategoryFrequentRecord{
			ID:         row.ID,
			Name:       row.Name,
			IconPath:   row.IconPath,
			ParentID:   row.ParentID.Int64,
			ParentName: row.ParentName.String,
			HasParent:  row.ParentID.Valid,
		})
	}
	return records, nil
}

func (r *CategoryRepository) buildCategoryBaseQuery(ctx context.Context, filter repository.CategoryListFilter) *gorm.DB {
	query := r.db.WithContext(ctx).Table("categories c")
	if filter.AccountBookID > 0 {
		query = query.Where("c.account_book_id = ?", filter.AccountBookID)
	}
	if filter.Type != "" {
		query = query.Where("c.type = ?", filter.Type)
	}
	return query
}

func (r *CategoryRepository) ListGuessedFrequentByStatementType(ctx context.Context, filter repository.CategoryGuessFilter) ([]repository.CategoryFrequentRecord, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 3
	}

	windowStart := filter.Now.Add(-30 * time.Minute).Format("15:04:05")
	windowEnd := filter.Now.Add(30 * time.Minute).Format("15:04:05")

	var rows []frequentCategoryRow
	err := r.db.WithContext(ctx).
		Table("categories c").
		Joins("INNER JOIN statements s ON s.category_id = c.id").
		Joins("LEFT JOIN categories parent ON parent.id = c.parent_id").
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path, parent.id AS parent_id, parent.name AS parent_name").
		Where("c.account_book_id = ?", filter.AccountBookID).
		Where("c.parent_id > 0").
		Where("c.frequent >= 5").
		Where("s.type = ?", filter.StatementType).
		Where("TIME(s.created_at) <= ? AND TIME(s.created_at) >= ?", windowEnd, windowStart).
		Group("c.id").
		Order("c.frequent DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	records := make([]repository.CategoryFrequentRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, repository.CategoryFrequentRecord{
			ID:         row.ID,
			Name:       row.Name,
			IconPath:   row.IconPath,
			ParentID:   row.ParentID.Int64,
			ParentName: row.ParentName.String,
			HasParent:  row.ParentID.Valid,
		})
	}
	return records, nil
}
