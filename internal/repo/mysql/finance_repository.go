package mysql

import (
	"context"
	"errors"
	"strings"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type FinanceRepository struct {
	db *gorm.DB
}

func NewFinanceRepository(db *gorm.DB) *FinanceRepository { return &FinanceRepository{db: db} }

func (r *FinanceRepository) ListAssets(ctx context.Context, accountBookID int64) ([]tablemodel.Asset, error) {
	rows := make([]tablemodel.Asset, 0)
	if err := r.db.WithContext(ctx).
		Table("assets").
		Where("account_book_id = ?", accountBookID).
		Order("parent_id ASC, id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *FinanceRepository) FindAssetByID(ctx context.Context, assetID int64, accountBookID int64) (tablemodel.Asset, error) {
	var row tablemodel.Asset
	err := r.db.WithContext(ctx).
		Table("assets").
		Where("id = ? AND account_book_id = ?", assetID, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.Asset{}, repo.ErrFinanceAssetNotFound
		}
		return tablemodel.Asset{}, err
	}
	return row, nil
}

func (r *FinanceRepository) SumStatementAmountByTypes(ctx context.Context, accountBookID int64, statementTypes []string) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND type IN ?", accountBookID, statementTypes).
		Scan(&row).Error
	return row.Amount, err
}

func (r *FinanceRepository) ListSpecialCategoryByTypes(ctx context.Context, statementTypes []string) ([]tablemodel.Category, error) {
	rows := make([]tablemodel.Category, 0)
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("special_type, id").
		Where("special_type IN ?", statementTypes).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *FinanceRepository) ListStatementSumsByTypes(ctx context.Context, accountBookID int64, statementTypes []string) ([]repo.FinanceStatementTypeSumRecord, error) {
	rows := make([]repo.FinanceStatementTypeSumRecord, 0)
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("type AS statement_type, COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND type IN ?", accountBookID, statementTypes).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *FinanceRepository) SumIncomeExpendByAsset(ctx context.Context, accountBookID int64, assetID int64) (repo.FinanceIncomeExpendSumRecord, error) {
	var row repo.FinanceIncomeExpendSumRecord
	err := r.db.WithContext(ctx).
		Table("statements").
		Select(strings.Join([]string{
			"COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income",
			"COALESCE(SUM(CASE WHEN type = 'expend' THEN amount ELSE 0 END), 0) AS expend",
		}, ", ")).
		Where("account_book_id = ? AND asset_id = ?", accountBookID, assetID).
		Scan(&row).Error
	if err != nil {
		return repo.FinanceIncomeExpendSumRecord{}, err
	}
	return row, nil
}

func (r *FinanceRepository) ListAssetTimeline(ctx context.Context, accountBookID int64, assetID int64) ([]repo.FinanceTimelineRecord, error) {
	rows := make([]repo.FinanceTimelineRecord, 0)
	err := r.db.WithContext(ctx).
		Table("statements").
		Select(strings.Join([]string{
			"year",
			"month",
			"COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income_amount",
			"COALESCE(SUM(CASE WHEN type = 'expend' THEN amount ELSE 0 END), 0) AS expend_amount",
		}, ", ")).
		Where("account_book_id = ? AND asset_id = ?", accountBookID, assetID).
		Group("year, month").
		Order("year DESC, month DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}
