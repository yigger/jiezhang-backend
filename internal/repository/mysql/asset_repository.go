package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"gorm.io/gorm"
)

func NewAssetRepository(db *gorm.DB) (*AssetRepository, error) {
	return &AssetRepository{db: db}, nil
}

type AssetRepository struct {
	db *gorm.DB
}

type assetParentRow struct {
	ID       int64  `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	IconPath string `gorm:"column:icon_path"`
}

func (r *AssetRepository) ListParents(ctx context.Context, filter repository.AssetFilter) ([]repository.AssetParentRecord, error) {
	var parents []assetParentRow
	if err := r.db.WithContext(ctx).
		Table("assets").
		Where("account_book_id = ? AND parent_id = 0", filter.AccountBookID).
		Select("id, name, icon_path").
		Order("`order` ASC, id ASC").
		Find(&parents).Error; err != nil {
		return nil, err
	}
	assetRecords := make([]repository.AssetParentRecord, 0, len(parents))
	for _, parent := range parents {
		assetRecords = append(assetRecords, repository.AssetParentRecord{
			ID:       parent.ID,
			Name:     parent.Name,
			IconPath: parent.IconPath,
		})
	}
	return assetRecords, nil
}

type assetChildRow struct {
	ID       int64  `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	IconPath string `gorm:"column:icon_path"`
	ParentID int64  `gorm:"column:parent_id"`
}

func (r *AssetRepository) ListChildrenByParentIDs(ctx context.Context, filter repository.AssetFilter, parentIDs []int64) ([]repository.AssetChildRecord, error) {
	if len(parentIDs) == 0 {
		return []repository.AssetChildRecord{}, nil
	}

	var rows []assetChildRow
	if err := r.db.WithContext(ctx).
		Table("assets").
		Where("account_book_id = ? AND parent_id IN ?", filter.AccountBookID, parentIDs).
		Select("id, name, icon_path, parent_id").
		Order("`order` ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	assetRecords := make([]repository.AssetChildRecord, 0, len(rows))
	for _, row := range rows {
		assetRecords = append(assetRecords, repository.AssetChildRecord{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: row.IconPath,
			ParentID: row.ParentID,
		})
	}
	return assetRecords, nil
}

type assetFrequentRow struct {
	ID         int64          `gorm:"column:id"`
	Name       string         `gorm:"column:name"`
	IconPath   string         `gorm:"column:icon_path"`
	ParentID   sql.NullInt64  `gorm:"column:parent_id"`
	ParentName sql.NullString `gorm:"column:parent_name"`
}

func (r *AssetRepository) ListFrequentChildren(ctx context.Context, filter repository.AssetFilter, limit int) ([]repository.AssetFrequentRecord, error) {
	var rows []assetFrequentRow
	if err := r.db.WithContext(ctx).
		Table("assets a").
		Joins("LEFT JOIN assets parent ON parent.id = a.parent_id").
		Select("a.id AS id, a.name AS name, a.icon_path AS icon_path, a.parent_id AS parent_id, COALESCE(parent.name, '') AS parent_name").
		Where("a.account_book_id = ?", filter.AccountBookID).
		Where("a.parent_id > 0").
		Where("a.frequent > 5").
		Order("a.frequent DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	records := make([]repository.AssetFrequentRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, repository.AssetFrequentRecord{
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

func (r *AssetRepository) ListGuessedFrequentByStatementTime(ctx context.Context, filter repository.AssetGuessFilter) ([]repository.AssetFrequentRecord, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 3
	}

	windowStart := filter.Now.Add(-30 * time.Minute).Format("15:04:05")
	windowEnd := filter.Now.Add(30 * time.Minute).Format("15:04:05")

	var rows []assetFrequentRow
	err := r.db.WithContext(ctx).
		Table("assets a").
		Joins("INNER JOIN statements s ON s.asset_id = a.id").
		Joins("LEFT JOIN assets parent ON parent.id = a.parent_id").
		Select("a.id AS id, a.name AS name, a.icon_path AS icon_path, parent.id AS parent_id, parent.name AS parent_name").
		Where("a.account_book_id = ?", filter.AccountBookID).
		Where("a.parent_id > 0").
		Where("a.frequent >= 5").
		Where("TIME(s.created_at) <= ? AND TIME(s.created_at) >= ?", windowEnd, windowStart).
		Group("a.id").
		Order("a.frequent DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	records := make([]repository.AssetFrequentRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, repository.AssetFrequentRecord{
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

type assetManageRow struct {
	ID       int64   `gorm:"column:id"`
	Name     string  `gorm:"column:name"`
	Order    int     `gorm:"column:order"`
	IconPath string  `gorm:"column:icon_path"`
	ParentID int64   `gorm:"column:parent_id"`
	Type     string  `gorm:"column:type"`
	Amount   float64 `gorm:"column:amount"`
	Remark   string  `gorm:"column:remark"`
}

type assetMutationRow struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	AccountBookID int64     `gorm:"column:account_book_id"`
	Name          string    `gorm:"column:name"`
	Amount        float64   `gorm:"column:amount"`
	ParentID      int64     `gorm:"column:parent_id"`
	Type          string    `gorm:"column:type"`
	IconPath      string    `gorm:"column:icon_path"`
	Remark        string    `gorm:"column:remark"`
	CreatorID     int64     `gorm:"column:creator_id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (assetMutationRow) TableName() string {
	return "assets"
}

func (r *AssetRepository) ListByParent(ctx context.Context, accountBookID int64, parentID int64) ([]repository.AssetManageRecord, error) {
	rows := make([]assetManageRow, 0)
	err := r.db.WithContext(ctx).
		Table("assets").
		Select("id, name, `order`, icon_path, parent_id, type, amount, COALESCE(remark, '') AS remark").
		Where("account_book_id = ? AND parent_id = ?", accountBookID, parentID).
		Order("type DESC, `order` ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.AssetManageRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.AssetManageRecord{
			ID:       row.ID,
			Name:     row.Name,
			Order:    row.Order,
			IconPath: row.IconPath,
			ParentID: row.ParentID,
			Type:     row.Type,
			Amount:   row.Amount,
			Remark:   row.Remark,
		})
	}
	return items, nil
}

func (r *AssetRepository) FindByID(ctx context.Context, accountBookID int64, id int64) (repository.AssetManageRecord, error) {
	var row assetManageRow
	err := r.db.WithContext(ctx).
		Table("assets").
		Select("id, name, `order`, icon_path, parent_id, type, amount, COALESCE(remark, '') AS remark").
		Where("id = ? AND account_book_id = ?", id, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.AssetManageRecord{}, repository.ErrAssetNotFound
		}
		return repository.AssetManageRecord{}, err
	}
	return repository.AssetManageRecord{
		ID:       row.ID,
		Name:     row.Name,
		Order:    row.Order,
		IconPath: row.IconPath,
		ParentID: row.ParentID,
		Type:     row.Type,
		Amount:   row.Amount,
		Remark:   row.Remark,
	}, nil
}

func (r *AssetRepository) CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error) {
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

func (r *AssetRepository) Create(ctx context.Context, input repository.AssetWriteRecord) (int64, error) {
	now := time.Now()
	row := assetMutationRow{
		AccountBookID: input.AccountBookID,
		Name:          input.Name,
		Amount:        input.Amount,
		ParentID:      input.ParentID,
		Type:          input.Type,
		IconPath:      input.IconPath,
		Remark:        input.Remark,
		CreatorID:     input.CreatorID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *AssetRepository) UpdateByID(ctx context.Context, id int64, accountBookID int64, input repository.AssetWriteRecord) error {
	res := r.db.WithContext(ctx).
		Table("assets").
		Where("id = ? AND account_book_id = ?", id, accountBookID).
		Updates(map[string]interface{}{
			"name":       input.Name,
			"amount":     input.Amount,
			"parent_id":  input.ParentID,
			"icon_path":  input.IconPath,
			"remark":     input.Remark,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrAssetNotFound
	}
	return nil
}

func (r *AssetRepository) DeleteByID(ctx context.Context, id int64, accountBookID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target assetMutationRow
		err := tx.Where("id = ? AND account_book_id = ?", id, accountBookID).Take(&target).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrAssetNotFound
			}
			return err
		}

		assetIDs := []int64{target.ID}
		if target.ParentID == 0 {
			childIDs := make([]int64, 0)
			if err := tx.Table("assets").Where("account_book_id = ? AND parent_id = ?", accountBookID, target.ID).Pluck("id", &childIDs).Error; err != nil {
				return err
			}
			assetIDs = append(assetIDs, childIDs...)
		}

		if err := tx.Table("statements").Where("account_book_id = ? AND asset_id IN ?", accountBookID, assetIDs).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		if err := tx.Table("assets").Where("account_book_id = ? AND id IN ?", accountBookID, assetIDs).Delete(&struct{}{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *AssetRepository) UpdateAmountByID(ctx context.Context, id int64, accountBookID int64, amount float64) error {
	res := r.db.WithContext(ctx).
		Table("assets").
		Where("id = ? AND account_book_id = ?", id, accountBookID).
		Updates(map[string]interface{}{
			"amount":     amount,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrAssetNotFound
	}
	return nil
}
