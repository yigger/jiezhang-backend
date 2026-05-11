package mysql

import (
	"context"

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
	ID         int64  `gorm:"column:id"`
	Name       string `gorm:"column:name"`
	IconPath   string `gorm:"column:icon_path"`
	ParentID   int64  `gorm:"column:parent_id"`
	ParentName string `gorm:"column:parent_name"`
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
			ParentID:   row.ParentID,
			ParentName: row.ParentName,
		})
	}
	return records, nil
}
