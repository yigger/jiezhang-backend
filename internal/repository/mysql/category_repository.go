package mysql

import (
	"context"
	"database/sql"
	"errors"
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

type categoryManageRow struct {
	ID       int64  `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	Order    int    `gorm:"column:order"`
	IconPath string `gorm:"column:icon_path"`
	ParentID int64  `gorm:"column:parent_id"`
	Type     string `gorm:"column:type"`
}

type categoryMutationRow struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID        int64     `gorm:"column:user_id"`
	AccountBookID int64     `gorm:"column:account_book_id"`
	Name          string    `gorm:"column:name"`
	ParentID      int64     `gorm:"column:parent_id"`
	Order         int       `gorm:"column:order"`
	IconPath      string    `gorm:"column:icon_path"`
	Type          string    `gorm:"column:type"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (categoryMutationRow) TableName() string {
	return "categories"
}

func (r *CategoryRepository) ListParents(ctx context.Context, filter repository.CategoryListFilter) ([]repository.CategoryParentRecord, error) {
	query := r.buildCategoryBaseQuery(ctx, filter)
	var parents []categoryParentRow
	if err := query.
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path").
		Where("c.parent_id = 0").
		Order("c.`order` ASC, c.id ASC").
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
		Order("c.`order` ASC, c.id ASC").
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

func (r *CategoryRepository) ListByParent(ctx context.Context, filter repository.CategoryListFilter, parentID int64) ([]repository.CategoryManageRecord, error) {
	rows := make([]categoryManageRow, 0)
	err := r.buildCategoryBaseQuery(ctx, filter).
		Select("c.id, c.name, c.`order`, c.icon_path, c.parent_id, c.type").
		Where("c.parent_id = ?", parentID).
		Order("c.`order` ASC, c.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.CategoryManageRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.CategoryManageRecord{
			ID:       row.ID,
			Name:     row.Name,
			Order:    row.Order,
			IconPath: row.IconPath,
			ParentID: row.ParentID,
			Type:     row.Type,
		})
	}
	return items, nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, accountBookID int64, id int64) (repository.CategoryManageRecord, error) {
	var row categoryManageRow
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("id, name, `order`, icon_path, parent_id, type").
		Where("id = ? AND account_book_id = ?", id, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.CategoryManageRecord{}, repository.ErrCategoryNotFound
		}
		return repository.CategoryManageRecord{}, err
	}
	return repository.CategoryManageRecord{
		ID:       row.ID,
		Name:     row.Name,
		Order:    row.Order,
		IconPath: row.IconPath,
		ParentID: row.ParentID,
		Type:     row.Type,
	}, nil
}

func (r *CategoryRepository) ListStatementAmountByCategoryIDs(ctx context.Context, accountBookID int64, categoryIDs []int64) ([]repository.CategoryAmountRecord, error) {
	if len(categoryIDs) == 0 {
		return []repository.CategoryAmountRecord{}, nil
	}
	rows := make([]repository.CategoryAmountRecord, 0)
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

func (r *CategoryRepository) ListStatementAmountByParentIDs(ctx context.Context, accountBookID int64, parentIDs []int64) ([]repository.CategoryAmountRecord, error) {
	if len(parentIDs) == 0 {
		return []repository.CategoryAmountRecord{}, nil
	}
	rows := make([]repository.CategoryAmountRecord, 0)
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

func (r *CategoryRepository) ListStatementsByCategory(ctx context.Context, accountBookID int64, categoryID int64) ([]repository.CategoryStatementRecord, error) {
	rows := make([]repository.CategoryStatementRecord, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN categories c ON c.id = s.category_id").
		Joins("LEFT JOIN assets a ON a.id = s.asset_id").
		Select([]string{
			"s.id AS id",
			"s.day AS day",
			"s.year AS year",
			"s.month AS month",
			"s.type AS type",
			"c.name AS category_name",
			"c.icon_path AS icon_path",
			"COALESCE(s.description, '') AS description",
			"s.amount AS amount",
			"s.created_at AS created_at",
			"COALESCE(a.name, '') AS asset_name",
		}).
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

func (r *CategoryRepository) Create(ctx context.Context, input repository.CategoryWriteRecord) (int64, error) {
	now := time.Now()
	row := categoryMutationRow{
		UserID:        input.UserID,
		AccountBookID: input.AccountBookID,
		Name:          input.Name,
		ParentID:      input.ParentID,
		IconPath:      input.IconPath,
		Type:          input.Type,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *CategoryRepository) UpdateByID(ctx context.Context, id int64, accountBookID int64, input repository.CategoryWriteRecord) error {
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
		return repository.ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) DeleteByID(ctx context.Context, id int64, accountBookID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target categoryMutationRow
		err := tx.Where("id = ? AND account_book_id = ?", id, accountBookID).Take(&target).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrCategoryNotFound
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
