package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type SuperStatementRepository struct {
	db *gorm.DB
}

func NewSuperStatementRepository(db *gorm.DB) *SuperStatementRepository {
	return &SuperStatementRepository{db: db}
}

func (r *SuperStatementRepository) ListSimpleRows(ctx context.Context, filter repo.SuperStatementFilter) ([]tablemodel.Statement, error) {
	query := r.db.WithContext(ctx).Table("statements s").Where("s.account_book_id = ?", filter.AccountBookID)
	query = applySuperStatementFilter(ctx, r.db, query, filter)
	query = query.Order(mapSuperOrderBy(filter.OrderBy))

	rows := make([]tablemodel.Statement, 0)
	err := query.Select(strings.Join([]string{
		"s.id AS id",
		"s.user_id AS user_id",
		"s.type AS type",
		"s.amount AS amount",
		"COALESCE(s.description, '') AS description",
		"s.category_id AS category_id",
		"s.asset_id AS asset_id",
		"COALESCE(s.target_asset_id, 0) AS target_asset_id",
		"COALESCE(s.target_object, '') AS target_object",
		"COALESCE(s.payee_id, 0) AS payee_id",
		"COALESCE(s.mood, '') AS mood",
		"COALESCE(s.residue, 0) AS residue",
		"COALESCE(s.location, '') AS location",
		"COALESCE(s.nation, '') AS nation",
		"COALESCE(s.province, '') AS province",
		"COALESCE(s.city, '') AS city",
		"COALESCE(s.district, '') AS district",
		"COALESCE(s.street, '') AS street",
		"s.created_at AS created_at",
		"s.updated_at AS updated_at",
	}, ", ")).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super list simple rows: %w", err)
	}

	return rows, nil
}

// BatchGetCategories fetches categories by primary keys.
func (r *SuperStatementRepository) BatchGetCategories(ctx context.Context, ids []int64) ([]tablemodel.Category, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows := make([]tablemodel.Category, 0)
	err := r.db.WithContext(ctx).
		Table("categories c").
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path, COALESCE(c.parent_id, 0) AS parent_id").
		Where("c.id IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch get categories: %w", err)
	}
	return rows, nil
}

// BatchGetAssets fetches assets by primary keys.
func (r *SuperStatementRepository) BatchGetAssets(ctx context.Context, ids []int64) ([]tablemodel.Asset, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows := make([]tablemodel.Asset, 0)
	err := r.db.WithContext(ctx).
		Table("assets a").
		Select("a.id AS id, a.name AS name, a.icon_path AS icon_path, COALESCE(a.parent_id, 0) AS parent_id").
		Where("a.id IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch get assets: %w", err)
	}
	return rows, nil
}

// BatchGetPayees fetches payees by primary keys.
func (r *SuperStatementRepository) BatchGetPayees(ctx context.Context, ids []int64) ([]tablemodel.Payee, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows := make([]tablemodel.Payee, 0)
	err := r.db.WithContext(ctx).
		Table("payees").
		Select("id, name").
		Where("id IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch get payees: %w", err)
	}
	items := make([]tablemodel.Payee, 0, len(rows))
	for _, row := range rows {
		items = append(items, tablemodel.Payee{ID: row.ID, Name: row.Name})
	}
	return items, nil
}

// BatchGetCollaboratorRemarks fetches collaborator remarks.
func (r *SuperStatementRepository) BatchGetCollaboratorRemarks(ctx context.Context, accountBookID int64, userIDs []int64) ([]tablemodel.AccountBookCollaborator, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	rows := make([]tablemodel.AccountBookCollaborator, 0)
	err := r.db.WithContext(ctx).
		Table("account_book_collaborators").
		Select("user_id, COALESCE(remark, '') AS remark").
		Where("account_book_id = ? AND user_id IN ?", accountBookID, userIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch get remarks: %w", err)
	}
	items := make([]tablemodel.AccountBookCollaborator, 0, len(rows))
	for _, row := range rows {
		items = append(items, tablemodel.AccountBookCollaborator{UserID: row.UserID, Remark: row.Remark})
	}
	return items, nil
}

// BatchCheckHasPic checks which statement IDs have StatementAvatar records.
func (r *SuperStatementRepository) BatchStatementAvatars(ctx context.Context, statementIDs []int64) ([]tablemodel.UserAsset, error) {
	if len(statementIDs) == 0 {
		return nil, nil
	}
	rows := make([]tablemodel.UserAsset, 0)
	err := r.db.WithContext(ctx).Table("user_assets").Select("imageable_id").
		Where("imageable_type = 'Statement' AND type = 'StatementAvatar' AND imageable_id IN ?", statementIDs).
		Group("imageable_id").Scan(&rows).Error
	return rows, err
}

func (r *SuperStatementRepository) ListMonthSummaries(ctx context.Context, filter repo.SuperStatementFilter) ([]repo.SuperStatementMonthSummaryRecord, error) {
	query := r.db.WithContext(ctx).Table("statements s").Where("s.account_book_id = ?", filter.AccountBookID)
	query = applySuperStatementFilter(ctx, r.db, query, filter)

	rows := make([]repo.SuperStatementMonthSummaryRecord, 0)
	err := query.
		Select([]string{
			"s.year AS year",
			"s.month AS month",
			"SUM(CASE WHEN s.type IN ('expend','repayment') THEN s.amount ELSE 0 END) AS expend_amount",
			"SUM(CASE WHEN s.type = 'income' THEN s.amount ELSE 0 END) AS income_amount",
			"SUM(CASE WHEN s.type = 'income' THEN s.amount ELSE 0 END) - SUM(CASE WHEN s.type IN ('expend','repayment') THEN s.amount ELSE 0 END) AS surplus",
		}).
		Group("s.year, s.month").
		Order("s.year DESC, s.month DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SuperStatementRepository) GetOverview(ctx context.Context, filter repo.SuperStatementFilter) (repo.SuperStatementOverviewRecord, error) {
	query := r.db.WithContext(ctx).Table("statements s").Where("s.account_book_id = ?", filter.AccountBookID)
	query = applySuperStatementFilter(ctx, r.db, query, filter)

	var row struct {
		Expend float64 `gorm:"column:expend"`
		Income float64 `gorm:"column:income"`
		Left   float64 `gorm:"column:left"`
	}
	err := query.
		Select([]string{
			"SUM(CASE WHEN s.type IN ('expend','repayment') THEN s.amount ELSE 0 END) AS expend",
			"SUM(CASE WHEN s.type = 'income' THEN s.amount ELSE 0 END) AS income",
			"SUM(CASE WHEN s.type = 'income' THEN s.amount ELSE 0 END) - SUM(CASE WHEN s.type IN ('expend','repayment') THEN s.amount ELSE 0 END) AS `left`",
		}).
		Scan(&row).Error
	if err != nil {
		return repo.SuperStatementOverviewRecord{}, err
	}
	return repo.SuperStatementOverviewRecord{
		Expend: row.Expend,
		Income: row.Income,
		Left:   row.Left,
	}, nil
}

type SuperChartRepository struct {
	db *gorm.DB
}

func NewSuperChartRepository(db *gorm.DB) *SuperChartRepository { return &SuperChartRepository{db: db} }

func (r *SuperChartRepository) GetMonthSummary(ctx context.Context, accountBookID int64, year int, month int, includeRepayment bool) (repo.SuperChartMonthSummary, error) {
	expendCases := "'expend'"
	if includeRepayment {
		expendCases = "'expend','repayment'"
	}
	var row struct {
		Expend float64 `gorm:"column:expend"`
		Income float64 `gorm:"column:income"`
	}
	err := r.db.WithContext(ctx).
		Table("statements").
		Select([]string{
			"SUM(CASE WHEN type IN (" + expendCases + ") THEN amount ELSE 0 END) AS expend",
			"SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS income",
		}).
		Where("account_book_id = ? AND year = ? AND month = ?", accountBookID, year, month).
		Scan(&row).Error
	if err != nil {
		return repo.SuperChartMonthSummary{}, err
	}
	return repo.SuperChartMonthSummary{Expend: row.Expend, Income: row.Income}, nil
}

func (r *SuperChartRepository) ListDaySummaries(ctx context.Context, accountBookID int64, year int, month int) ([]repo.SuperChartDaySummary, error) {
	rows := make([]repo.SuperChartDaySummary, 0)
	err := r.db.WithContext(ctx).
		Table("statements").
		Select([]string{
			"day AS day",
			"SUM(CASE WHEN type IN ('expend','repayment') THEN amount ELSE 0 END) AS expend",
			"SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS income",
		}).
		Where("account_book_id = ? AND year = ? AND month = ?", accountBookID, year, month).
		Group("day").
		Order("day ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SuperChartRepository) ListPieParents(ctx context.Context, accountBookID int64, year int, month int, statementType string) ([]repo.SuperChartPieParentItem, error) {
	statementType = strings.TrimSpace(statementType)
	if statementType == "" {
		statementType = "expend"
	}

	rows := make([]repo.SuperChartPieParentItem, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN categories c ON c.id = s.category_id").
		Joins("LEFT JOIN categories p ON p.id = c.parent_id").
		Select([]string{
			"c.parent_id AS parent_id",
			"COALESCE(p.name, '') AS parent_name",
			"SUM(s.amount) AS data",
		}).
		Where("s.account_book_id = ? AND s.year = ? AND s.month = ? AND s.type = ?", accountBookID, year, month, statementType).
		Where("c.parent_id > 0").
		Group("c.parent_id, p.name").
		Order("data DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SuperChartRepository) ListCategoryTop(ctx context.Context, accountBookID int64, year int, month int) ([]repo.SuperChartCategoryTopItem, error) {
	rows := make([]repo.SuperChartCategoryTopItem, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN categories c ON c.id = s.category_id").
		Select([]string{
			"c.id AS category_id",
			"COALESCE(c.name, '') AS name",
			"SUM(s.amount) AS data",
		}).
		Where("s.account_book_id = ? AND s.year = ? AND s.month = ?", accountBookID, year, month).
		Where("s.type IN ?", []string{"expend", "repayment"}).
		Group("c.id, c.name").
		Order("data DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SuperChartRepository) ListYearMonths(ctx context.Context, accountBookID int64) ([]repo.SuperChartYearMonth, error) {
	rows := make([]repo.SuperChartYearMonth, 0)
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("year, month").
		Where("account_book_id = ?", accountBookID).
		Group("year, month").
		Order("year ASC, month ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SuperChartRepository) SumExpendBetween(ctx context.Context, accountBookID int64, start time.Time, end time.Time) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND type IN ? AND created_at >= ? AND created_at <= ?", accountBookID, []string{"expend", "repayment"}, start, end).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Amount, nil
}

func (r *SuperStatementRepository) baseListQuery(ctx context.Context, accountBookID int64) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN categories c ON c.id = s.category_id").
		Joins("LEFT JOIN assets a ON a.id = s.asset_id").
		Joins("LEFT JOIN payees p ON p.id = s.payee_id").
		Joins("LEFT JOIN account_book_collaborators abc ON abc.account_book_id = s.account_book_id AND abc.user_id = s.user_id").
		Joins("LEFT JOIN assets ta ON ta.id = s.target_asset_id").
		Where("s.account_book_id = ?", accountBookID)
}

func applySuperStatementFilter(ctx context.Context, db *gorm.DB, query *gorm.DB, filter repo.SuperStatementFilter) *gorm.DB {
	if filter.Year != nil && *filter.Year > 0 {
		query = query.Where("s.year = ?", *filter.Year)
	}
	if filter.Month != nil && *filter.Month >= 0 {
		if *filter.Month != -1 {
			query = query.Where("s.month = ?", *filter.Month)
		}
	}
	if filter.AssetParentID != nil && *filter.AssetParentID > 0 {
		sub := db.Table("assets").Select("id").Where("account_book_id = ? AND parent_id = ?", filter.AccountBookID, *filter.AssetParentID)
		query = query.Where("s.asset_id IN (?)", sub)
	}
	if filter.AssetID != nil && *filter.AssetID > 0 {
		query = query.Where("s.asset_id = ?", *filter.AssetID)
	}
	if filter.CategoryID != nil && *filter.CategoryID > 0 {
		var specialType string
		if err := db.WithContext(ctx).Table("categories").Select("special_type").Where("id = ?", *filter.CategoryID).Take(&specialType).Error; err == nil && specialType != "" {
			query = query.Where("s.category_id = ?", *filter.CategoryID)
		} else {
			sub := db.Table("categories").Select("id").Where("account_book_id = ? AND (id = ? OR parent_id = ?)", filter.AccountBookID, *filter.CategoryID, *filter.CategoryID)
			query = query.Where("s.category_id IN (?)", sub)
		}
	}
	return query
}

func mapSuperOrderBy(orderBy string) string {
	switch strings.TrimSpace(strings.ToLower(orderBy)) {
	case "amount":
		return "s.amount DESC"
	default:
		return "s.created_at DESC"
	}
}

var _ repo.SuperStatementRepository = (*SuperStatementRepository)(nil)
var _ repo.SuperChartRepository = (*SuperChartRepository)(nil)
