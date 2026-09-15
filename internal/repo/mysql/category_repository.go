package mysql

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository { return &CategoryRepository{db: db} }

func (r *CategoryRepository) ListParents(ctx context.Context, accountBookID int64, categoryType string) ([]tablemodel.Category, error) {
	query := r.buildCategoryBaseQuery(ctx, accountBookID, categoryType)
	parents := make([]tablemodel.Category, 0)
	if err := query.
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path").
		Where("c.parent_id = 0").
		Order("c.`order` ASC, c.id ASC").
		Scan(&parents).Error; err != nil {
		return nil, err
	}
	return parents, nil
}

func (r *CategoryRepository) ListChildrenByParentIDs(ctx context.Context, accountBookID int64, categoryType string, parentIDs []int64) ([]tablemodel.Category, error) {
	if len(parentIDs) == 0 {
		return []tablemodel.Category{}, nil
	}
	query := r.buildCategoryBaseQuery(ctx, accountBookID, categoryType)
	rows := make([]tablemodel.Category, 0)
	if err := query.
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path, c.parent_id AS parent_id").
		Where("c.parent_id IN ?", parentIDs).
		Order("c.`order` ASC, c.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CategoryRepository) ListFrequentChildren(ctx context.Context, accountBookID int64, categoryType string, limit int) ([]tablemodel.Category, error) {
	if limit <= 0 {
		limit = 10
	}
	query := r.buildCategoryBaseQuery(ctx, accountBookID, categoryType)
	var frequentRows []tablemodel.Category
	if err := query.
		Joins("LEFT JOIN categories parent ON parent.id = c.parent_id").
		Select("c.*").
		Where("c.parent_id > 0").
		Where("c.frequent > 5").
		Order("c.frequent DESC").
		Limit(limit).
		Scan(&frequentRows).Error; err != nil {
		return nil, err
	}
	return frequentRows, nil
}

func (r *CategoryRepository) buildCategoryBaseQuery(ctx context.Context, accountBookID int64, categoryType string) *gorm.DB {
	query := r.db.WithContext(ctx).Table("categories c")
	if accountBookID > 0 {
		query = query.Where("c.account_book_id = ?", accountBookID)
	}
	if categoryType != "" {
		query = query.Where("c.type = ?", categoryType)
	}
	return query
}

func (r *CategoryRepository) ListGuessedFrequentByStatementType(ctx context.Context, filter repo.CategoryGuessFilter) ([]tablemodel.Category, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 3
	}

	windowStart := filter.Now.Add(-30 * time.Minute).Format("15:04:05")
	windowEnd := filter.Now.Add(30 * time.Minute).Format("15:04:05")

	var rows []tablemodel.Category
	err := r.db.WithContext(ctx).
		Table("categories c").
		Joins("INNER JOIN statements s ON s.category_id = c.id").
		Joins("LEFT JOIN categories parent ON parent.id = c.parent_id").
		Select("c.*").
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

	return rows, nil
}

func (r *CategoryRepository) ListByParent(ctx context.Context, accountBookID int64, categoryType string, parentID int64) ([]tablemodel.Category, error) {
	rows := make([]tablemodel.Category, 0)
	err := r.buildCategoryBaseQuery(ctx, accountBookID, categoryType).
		Select("c.id, c.name, c.`order`, c.icon_path, c.parent_id, c.type").
		Where("c.parent_id = ?", parentID).
		Order("c.`order` ASC, c.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, accountBookID int64, id int64) (tablemodel.Category, error) {
	var row tablemodel.Category
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("id, name, `order`, icon_path, parent_id, type").
		Where("id = ? AND account_book_id = ?", id, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.Category{}, repo.ErrCategoryNotFound
		}
		return tablemodel.Category{}, err
	}
	return row, nil
}

func (r *CategoryRepository) FindBySpecialType(ctx context.Context, specialType string) (int64, error) {
	var id int64
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("id").
		Where("special_type = ?", specialType).
		Take(&id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, repo.ErrCategoryNotFound
		}
		return 0, err
	}
	return id, nil
}

func (r *CategoryRepository) ListStatementAmountByCategoryIDs(ctx context.Context, accountBookID int64, categoryIDs []int64) ([]repo.CategoryAmountRecord, error) {
	if len(categoryIDs) == 0 {
		return []repo.CategoryAmountRecord{}, nil
	}
	rows := make([]repo.CategoryAmountRecord, 0)
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("category_id AS category_id, COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND category_id IN ?", accountBookID, categoryIDs).
		Group("category_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CategoryRepository) ListStatementAmountByParentIDs(ctx context.Context, accountBookID int64, parentIDs []int64) ([]repo.CategoryAmountRecord, error) {
	if len(parentIDs) == 0 {
		return []repo.CategoryAmountRecord{}, nil
	}
	rows := make([]repo.CategoryAmountRecord, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN categories c ON c.id = s.category_id").
		Select("c.parent_id AS category_id, COALESCE(SUM(s.amount), 0) AS amount").
		Where("s.account_book_id = ? AND c.parent_id IN ?", accountBookID, parentIDs).
		Group("c.parent_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CategoryRepository) ListStatementsByCategory(ctx context.Context, accountBookID int64, categoryID int64) ([]tablemodel.Statement, error) {
	rows := make([]tablemodel.Statement, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN categories c ON c.id = s.category_id").
		Select("s.*").
		Where("s.account_book_id = ? AND s.category_id = ?", accountBookID, categoryID).
		Order("s.year DESC, s.month DESC, s.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CategoryRepository) SumStatements(ctx context.Context, accountBookID int64, statementType string, categoryIDs []int64, year int, month int) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}

	query := r.db.WithContext(ctx).
		Table("statements").
		Select("COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ?", accountBookID)

	if statementType != "" {
		query = query.Where("type = ?", statementType)
	}
	if len(categoryIDs) > 0 {
		query = query.Where("category_id IN ?", categoryIDs)
	}
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	if month > 0 {
		query = query.Where("month = ?", month)
	}

	if err := query.Scan(&row).Error; err != nil {
		return 0, err
	}
	return row.Amount, nil
}

func (r *CategoryRepository) CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error) {
	var ownerCount int64
	if err := r.db.WithContext(ctx).
		Table("account_books").
		Where("id = ? AND user_id = ?", accountBookID, userID).
		Count(&ownerCount).Error; err != nil {
		return false, err
	}
	if ownerCount > 0 {
		return true, nil
	}

	var collaboratorCount int64
	if err := r.db.WithContext(ctx).
		Table("account_book_collaborators").
		Where("account_book_id = ? AND user_id = ? AND role IN ?", accountBookID, userID, []string{"owner", "admin"}).
		Count(&collaboratorCount).Error; err != nil {
		return false, err
	}
	return collaboratorCount > 0, nil
}

func (r *CategoryRepository) Create(ctx context.Context, input tablemodel.Category) (int64, error) {
	now := time.Now()
	row := input
	row.CreatedAt = now
	row.UpdatedAt = now
	if err := r.db.WithContext(ctx).Omit("Budget", "SpecialType", "Frequent").Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *CategoryRepository) UpdateByID(ctx context.Context, id int64, accountBookID int64, input tablemodel.Category) error {
	res := r.db.WithContext(ctx).
		Table("categories").
		Where("id = ? AND account_book_id = ?", id, accountBookID).
		Updates(map[string]interface{}{
			"name":       input.Name,
			"parent_id":  input.ParentID,
			"icon_path":  input.IconPath,
			"type":       input.Type,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) DeleteByID(ctx context.Context, id int64, accountBookID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target tablemodel.Category
		err := tx.Where("id = ? AND account_book_id = ?", id, accountBookID).Take(&target).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repo.ErrCategoryNotFound
			}
			return err
		}

		categoryIDs := []int64{target.ID}
		if target.ParentID == 0 {
			childIDs := make([]int64, 0)
			if err := tx.Table("categories").Where("account_book_id = ? AND parent_id = ?", accountBookID, target.ID).Pluck("id", &childIDs).Error; err != nil {
				return err
			}
			categoryIDs = append(categoryIDs, childIDs...)
		}

		if err := tx.Table("statements").Where("account_book_id = ? AND category_id IN ?", accountBookID, categoryIDs).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		if err := tx.Table("categories").Where("account_book_id = ? AND id IN ?", accountBookID, categoryIDs).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *CategoryRepository) ListByIDs(ctx context.Context, ids []int64) ([]tablemodel.Category, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []tablemodel.Category
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}

func (r *CategoryRepository) BatchGetAssets(ctx context.Context, ids []int64) ([]tablemodel.Asset, error) {
	return (&StatementRepository{db: r.db}).BatchGetAssets(ctx, ids)
}
