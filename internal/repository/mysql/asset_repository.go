package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"gorm.io/gorm"
)

func NewAssetRepository(db *gorm.DB) (repository.AssetRepository, error) {
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
		Find(&parents).Error; err != nil {
		return nil, err
	}
	var assetRecords []repository.AssetParentRecord
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
		Find(&rows).Error; err != nil {
		return nil, err
	}

	var assetRecords []repository.AssetChildRecord
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
