package mysql

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type FinanceRepository struct {
	db *gorm.DB
}

type financeAssetRow struct {
	ID       int64   `gorm:"column:id"`
	Name     string  `gorm:"column:name"`
	Amount   float64 `gorm:"column:amount"`
	ParentID int64   `gorm:"column:parent_id"`
	IconPath string  `gorm:"column:icon_path"`
	Type     string  `gorm:"column:type"`
}

type financeSpecialCategoryRow struct {
	SpecialType string `gorm:"column:special_type"`
	ID          int64  `gorm:"column:id"`
}

type financeStatementTypeSumRow struct {
	StatementType string  `gorm:"column:statement_type"`
	Amount        float64 `gorm:"column:amount"`
}

type financeIncomeExpendSumRow struct {
	Income float64 `gorm:"column:income"`
	Expend float64 `gorm:"column:expend"`
}

type financeTimelineRow struct {
	Year         int     `gorm:"column:year"`
	Month        int     `gorm:"column:month"`
	IncomeAmount float64 `gorm:"column:income_amount"`
	ExpendAmount float64 `gorm:"column:expend_amount"`
}

func NewFinanceRepository(db *gorm.DB) (*FinanceRepository, error) {
	return &FinanceRepository{db: db}, nil
}

func (r *FinanceRepository) ListAssets(ctx context.Context, accountBookID int64) ([]repository.FinanceAssetRecord, error) {
	rows := make([]financeAssetRow, 0)
	if err := r.db.WithContext(ctx).
		Table("assets").
		Where("account_book_id = ?", accountBookID).
		Order("parent_id ASC, id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]repository.FinanceAssetRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.FinanceAssetRecord{
			ID:       row.ID,
			Name:     row.Name,
			Amount:   row.Amount,
			ParentID: row.ParentID,
			IconPath: row.IconPath,
			Type:     row.Type,
		})
	}
	return items, nil
}

func (r *FinanceRepository) FindAssetByID(ctx context.Context, assetID int64, accountBookID int64) (repository.FinanceAssetRecord, error) {
	var row financeAssetRow
	err := r.db.WithContext(ctx).
		Table("assets").
		Where("id = ? AND account_book_id = ?", assetID, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.FinanceAssetRecord{}, repository.ErrFinanceAssetNotFound
		}
		return repository.FinanceAssetRecord{}, err
	}
	return repository.FinanceAssetRecord{
		ID:       row.ID,
		Name:     row.Name,
		Amount:   row.Amount,
		ParentID: row.ParentID,
		IconPath: row.IconPath,
		Type:     row.Type,
	}, nil
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

func (r *FinanceRepository) ListSpecialCategoryByTypes(ctx context.Context, accountBookID int64, statementTypes []string) ([]repository.FinanceSpecialCategoryRecord, error) {
	rows := make([]financeSpecialCategoryRow, 0)
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("special_type, id").
		Where("account_book_id = ? AND special_type IN ?", accountBookID, statementTypes).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.FinanceSpecialCategoryRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.FinanceSpecialCategoryRecord{
			SpecialType: row.SpecialType,
			ID:          row.ID,
		})
	}
	return items, nil
}

func (r *FinanceRepository) ListStatementSumsByTypes(ctx context.Context, accountBookID int64, statementTypes []string) ([]repository.FinanceStatementTypeSumRecord, error) {
	rows := make([]financeStatementTypeSumRow, 0)
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("type AS statement_type, COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND type IN ?", accountBookID, statementTypes).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.FinanceStatementTypeSumRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.FinanceStatementTypeSumRecord{
			StatementType: row.StatementType,
			Amount:        row.Amount,
		})
	}
	return items, nil
}

func (r *FinanceRepository) SumIncomeExpendByAsset(ctx context.Context, accountBookID int64, assetID int64) (repository.FinanceIncomeExpendSumRecord, error) {
	var row financeIncomeExpendSumRow
	err := r.db.WithContext(ctx).
		Table("statements").
		Select(strings.Join([]string{
			"COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income",
			"COALESCE(SUM(CASE WHEN type = 'expend' THEN amount ELSE 0 END), 0) AS expend",
		}, ", ")).
		Where("account_book_id = ? AND asset_id = ?", accountBookID, assetID).
		Scan(&row).Error
	if err != nil {
		return repository.FinanceIncomeExpendSumRecord{}, err
	}
	return repository.FinanceIncomeExpendSumRecord{Income: row.Income, Expend: row.Expend}, nil
}

func (r *FinanceRepository) ListAssetTimeline(ctx context.Context, accountBookID int64, assetID int64) ([]repository.FinanceTimelineRecord, error) {
	rows := make([]financeTimelineRow, 0)
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

	items := make([]repository.FinanceTimelineRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.FinanceTimelineRecord{
			Year:         row.Year,
			Month:        row.Month,
			IncomeAmount: row.IncomeAmount,
			ExpendAmount: row.ExpendAmount,
		})
	}
	return items, nil
}
