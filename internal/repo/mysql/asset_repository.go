package mysql

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

func NewAssetRepository(db *gorm.DB) *AssetRepository { return &AssetRepository{db: db} }

type AssetRepository struct {
	db *gorm.DB
}

func (r *AssetRepository) ListParents(ctx context.Context, accountBookID int64) ([]tablemodel.Asset, error) {
	parents := make([]tablemodel.Asset, 0)
	if err := r.db.WithContext(ctx).
		Table("assets").
		Where("account_book_id = ? AND parent_id = 0", accountBookID).
		Select("id, name, icon_path").
		Order("`order` ASC, id ASC").
		Find(&parents).Error; err != nil {
		return nil, err
	}
	return parents, nil
}

func (r *AssetRepository) ListChildrenByParentIDs(ctx context.Context, accountBookID int64, parentIDs []int64) ([]tablemodel.Asset, error) {
	if len(parentIDs) == 0 {
		return []tablemodel.Asset{}, nil
	}

	rows := make([]tablemodel.Asset, 0)
	if err := r.db.WithContext(ctx).
		Table("assets").
		Where("account_book_id = ? AND parent_id IN ?", accountBookID, parentIDs).
		Select("id, name, icon_path, parent_id").
		Order("`order` ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *AssetRepository) ListFrequentChildren(ctx context.Context, accountBookID int64, limit int) ([]tablemodel.Asset, error) {
	var rows []tablemodel.Asset
	if err := r.db.WithContext(ctx).
		Table("assets a").
		Select("a.*").
		Where("a.account_book_id = ?", accountBookID).
		Where("a.parent_id > 0").
		Where("a.frequent > 5").
		Order("a.frequent DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *AssetRepository) ListGuessedFrequentByStatementTime(ctx context.Context, filter repo.AssetGuessFilter) ([]tablemodel.Asset, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 3
	}

	windowStart := filter.Now.Add(-30 * time.Minute).Format("15:04:05")
	windowEnd := filter.Now.Add(30 * time.Minute).Format("15:04:05")

	var rows []tablemodel.Asset
	err := r.db.WithContext(ctx).
		Table("assets a").
		Joins("INNER JOIN statements s ON s.asset_id = a.id").
		Select("a.*").
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

	return rows, nil
}

func (r *AssetRepository) ListByParent(ctx context.Context, accountBookID int64, parentID int64) ([]tablemodel.Asset, error) {
	rows := make([]tablemodel.Asset, 0)
	err := r.db.WithContext(ctx).
		Table("assets").
		Select("id, name, `order`, icon_path, parent_id, type, amount, COALESCE(remark, '') AS remark").
		Where("account_book_id = ? AND parent_id = ?", accountBookID, parentID).
		Order("type DESC, `order` ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *AssetRepository) FindByID(ctx context.Context, accountBookID int64, id int64) (tablemodel.Asset, error) {
	var row tablemodel.Asset
	err := r.db.WithContext(ctx).
		Table("assets").
		Select("id, name, `order`, icon_path, parent_id, type, amount, COALESCE(remark, '') AS remark").
		Where("id = ? AND account_book_id = ?", id, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.Asset{}, repo.ErrAssetNotFound
		}
		return tablemodel.Asset{}, err
	}
	return row, nil
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

func (r *AssetRepository) Create(ctx context.Context, input tablemodel.Asset) (int64, error) {
	now := time.Now()
	row := input
	row.CreatedAt = now
	row.UpdatedAt = now
	if err := r.db.WithContext(ctx).Omit("Order", "Frequent").Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *AssetRepository) UpdateByID(ctx context.Context, id int64, accountBookID int64, input tablemodel.Asset) error {
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
		return repo.ErrAssetNotFound
	}
	return nil
}

func (r *AssetRepository) DeleteByID(ctx context.Context, id int64, accountBookID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target tablemodel.Asset
		err := tx.Where("id = ? AND account_book_id = ?", id, accountBookID).Take(&target).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repo.ErrAssetNotFound
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
		return repo.ErrAssetNotFound
	}
	return nil
}

func (r *AssetRepository) ListByIDs(ctx context.Context, ids []int64) ([]tablemodel.Asset, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []tablemodel.Asset
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}
