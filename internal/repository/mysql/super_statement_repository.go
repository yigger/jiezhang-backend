package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type SuperStatementRepository struct {
	db *gorm.DB
}

func NewSuperStatementRepository(db *gorm.DB) (*SuperStatementRepository, error) {
	return &SuperStatementRepository{db: db}, nil
}

func (r *SuperStatementRepository) ListRowsWithRelations(ctx context.Context, filter repository.SuperStatementFilter) ([]repository.StatementListRowRecord, error) {
	query := r.baseListQuery(ctx, filter.AccountBookID)
	query = applySuperStatementFilter(ctx, r.db, query, filter)
	query = query.Order(mapSuperOrderBy(filter.OrderBy))

	rows := make([]statementListRow, 0)
	err := query.Select(strings.Join([]string{
		"s.id AS id",
		"s.type AS type",
		"s.amount AS amount",
		"COALESCE(s.description, '') AS description",
		"COALESCE(abc.remark, '') AS remark",
		"COALESCE(s.mood, '') AS mood",
		"COALESCE(s.target_object, '') AS target_object",
		"s.category_id AS category_id",
		"s.asset_id AS asset_id",
		"COALESCE(s.target_asset_id, 0) AS target_asset_id",
		"COALESCE(c.icon_path, '') AS icon_path",
		"s.created_at AS created_at",
		"s.updated_at AS updated_at",
		"COALESCE(c.name, '') AS category_name",
		"COALESCE(a.name, '') AS asset_name",
		"COALESCE(s.location, '') AS location",
		"COALESCE(s.nation, '') AS nation",
		"COALESCE(s.province, '') AS province",
		"COALESCE(s.city, '') AS city",
		"COALESCE(s.district, '') AS district",
		"COALESCE(s.street, '') AS street",
		"EXISTS (SELECT 1 FROM user_assets ua WHERE ua.imageable_type = 'Statement' AND ua.type = 'StatementAvatar' AND ua.imageable_id = s.id) AS has_pic",
		"COALESCE(s.payee_id, 0) AS payee_id",
		"COALESCE(p.name, '') AS payee_name",
		"COALESCE(ta.name, '') AS target_asset_name",
	}, ", ")).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.StatementListRowRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementListRowRecord{
			ID:              row.ID,
			Type:            row.Type,
			Amount:          row.Amount,
			Description:     row.Description,
			Remark:          row.Remark,
			Mood:            row.Mood,
			TargetObject:    row.TargetObject,
			CategoryID:      row.CategoryID,
			AssetID:         row.AssetID,
			TargetAssetID:   row.TargetAssetID,
			IconPath:        row.IconPath,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			CategoryName:    row.CategoryName,
			AssetName:       row.AssetName,
			Location:        row.Location,
			Nation:          row.Nation,
			Province:        row.Province,
			City:            row.City,
			District:        row.District,
			Street:          row.Street,
			HasPic:          row.HasPic,
			PayeeID:         row.PayeeID,
			PayeeName:       row.PayeeName,
			TargetAssetName: row.TargetAssetName,
		})
	}
	return items, nil
}

// ListSimpleRows returns statement rows from a single-table query (no JOINs).
func (r *SuperStatementRepository) ListSimpleRows(ctx context.Context, filter repository.SuperStatementFilter) ([]repository.StatementSimpleRowRecord, error) {
	query := r.db.WithContext(ctx).Table("statements s").Where("s.account_book_id = ?", filter.AccountBookID)
	query = applySuperStatementFilter(ctx, r.db, query, filter)
	query = query.Order(mapSuperOrderBy(filter.OrderBy))

	var rows []statementSimpleRow
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

	items := make([]repository.StatementSimpleRowRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementSimpleRowRecord{
			ID:            row.ID,
			UserID:        row.UserID,
			Type:          row.Type,
			Amount:        row.Amount,
			Description:   row.Description,
			CategoryID:    row.CategoryID,
			AssetID:       row.AssetID,
			TargetAssetID: row.TargetAssetID,
			TargetObject:  row.TargetObject,
			PayeeID:       row.PayeeID,
			Mood:          row.Mood,
			Residue:       row.Residue,
			Location:      row.Location,
			Nation:        row.Nation,
			Province:      row.Province,
			City:          row.City,
			District:      row.District,
			Street:        row.Street,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}
	return items, nil
}

// BatchGetCategories fetches categories by primary keys.
func (r *SuperStatementRepository) BatchGetCategories(ctx context.Context, ids []int64) ([]repository.CategoryBatchRecord, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []categoryBatchRow
	err := r.db.WithContext(ctx).
		Table("categories c").
		Joins("LEFT JOIN categories cp ON cp.id = c.parent_id").
		Select("c.id AS id, c.name AS name, c.icon_path AS icon_path, COALESCE(c.parent_id, 0) AS parent_id, COALESCE(cp.name, '') AS parent_name").
		Where("c.id IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch get categories: %w", err)
	}
	items := make([]repository.CategoryBatchRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.CategoryBatchRecord{
			ID: row.ID, Name: row.Name, IconPath: row.IconPath,
			ParentID: row.ParentID, ParentName: row.ParentName,
		})
	}
	return items, nil
}

// BatchGetAssets fetches assets by primary keys.
func (r *SuperStatementRepository) BatchGetAssets(ctx context.Context, ids []int64) ([]repository.AssetBatchRecord, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []assetBatchRow
	err := r.db.WithContext(ctx).
		Table("assets a").
		Joins("LEFT JOIN assets ap ON ap.id = a.parent_id").
		Select("a.id AS id, a.name AS name, a.icon_path AS icon_path, COALESCE(a.parent_id, 0) AS parent_id, COALESCE(ap.name, '') AS parent_name").
		Where("a.id IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch get assets: %w", err)
	}
	items := make([]repository.AssetBatchRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.AssetBatchRecord{
			ID: row.ID, Name: row.Name, IconPath: row.IconPath,
			ParentID: row.ParentID, ParentName: row.ParentName,
		})
	}
	return items, nil
}

// BatchGetPayees fetches payees by primary keys.
func (r *SuperStatementRepository) BatchGetPayees(ctx context.Context, ids []int64) ([]repository.PayeeBatchRecord, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []payeeBatchRow
	err := r.db.WithContext(ctx).
		Table("payees").
		Select("id, name").
		Where("id IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch get payees: %w", err)
	}
	items := make([]repository.PayeeBatchRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.PayeeBatchRecord{ID: row.ID, Name: row.Name})
	}
	return items, nil
}

// BatchGetCollaboratorRemarks fetches collaborator remarks.
func (r *SuperStatementRepository) BatchGetCollaboratorRemarks(ctx context.Context, accountBookID int64, userIDs []int64) ([]repository.CollaboratorRemarkRecord, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var rows []collaboratorRemarkRow
	err := r.db.WithContext(ctx).
		Table("account_book_collaborators").
		Select("user_id, COALESCE(remark, '') AS remark").
		Where("account_book_id = ? AND user_id IN ?", accountBookID, userIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch get remarks: %w", err)
	}
	items := make([]repository.CollaboratorRemarkRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.CollaboratorRemarkRecord{UserID: row.UserID, Remark: row.Remark})
	}
	return items, nil
}

// BatchCheckHasPic checks which statement IDs have StatementAvatar records.
func (r *SuperStatementRepository) BatchCheckHasPic(ctx context.Context, statementIDs []int64) ([]repository.StatementHasPicRecord, error) {
	if len(statementIDs) == 0 {
		return nil, nil
	}
	var rows []hasPicRow
	err := r.db.WithContext(ctx).
		Table("user_assets").
		Select("imageable_id AS statement_id, 1 AS has_pic").
		Where("imageable_type = 'Statement' AND type = 'StatementAvatar' AND imageable_id IN ?", statementIDs).
		Group("imageable_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("super batch check has pic: %w", err)
	}
	items := make([]repository.StatementHasPicRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementHasPicRecord{StatementID: row.StatementID, HasPic: row.HasPic})
	}
	return items, nil
}

func (r *SuperStatementRepository) ListMonthSummaries(ctx context.Context, filter repository.SuperStatementFilter) ([]repository.SuperStatementMonthSummaryRecord, error) {
	query := r.db.WithContext(ctx).Table("statements s").Where("s.account_book_id = ?", filter.AccountBookID)
	query = applySuperStatementFilter(ctx, r.db, query, filter)

	rows := make([]repository.SuperStatementMonthSummaryRecord, 0)
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

func (r *SuperStatementRepository) GetOverview(ctx context.Context, filter repository.SuperStatementFilter) (repository.SuperStatementOverviewRecord, error) {
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
		return repository.SuperStatementOverviewRecord{}, err
	}
	return repository.SuperStatementOverviewRecord{
		Expend: row.Expend,
		Income: row.Income,
		Left:   row.Left,
	}, nil
}

type SuperChartRepository struct {
	db *gorm.DB
}

func NewSuperChartRepository(db *gorm.DB) (*SuperChartRepository, error) {
	return &SuperChartRepository{db: db}, nil
}

func (r *SuperChartRepository) GetMonthSummary(ctx context.Context, accountBookID int64, year int, month int, includeRepayment bool) (repository.SuperChartMonthSummary, error) {
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
		return repository.SuperChartMonthSummary{}, err
	}
	return repository.SuperChartMonthSummary{Expend: row.Expend, Income: row.Income}, nil
}

func (r *SuperChartRepository) ListDaySummaries(ctx context.Context, accountBookID int64, year int, month int) ([]repository.SuperChartDaySummary, error) {
	rows := make([]repository.SuperChartDaySummary, 0)
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

func (r *SuperChartRepository) ListPieParents(ctx context.Context, accountBookID int64, year int, month int, statementType string) ([]repository.SuperChartPieParentItem, error) {
	statementType = strings.TrimSpace(statementType)
	if statementType == "" {
		statementType = "expend"
	}

	rows := make([]repository.SuperChartPieParentItem, 0)
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

func (r *SuperChartRepository) ListCategoryTop(ctx context.Context, accountBookID int64, year int, month int) ([]repository.SuperChartCategoryTopItem, error) {
	rows := make([]repository.SuperChartCategoryTopItem, 0)
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

func (r *SuperChartRepository) ListYearMonths(ctx context.Context, accountBookID int64) ([]repository.SuperChartYearMonth, error) {
	rows := make([]repository.SuperChartYearMonth, 0)
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

func applySuperStatementFilter(ctx context.Context, db *gorm.DB, query *gorm.DB, filter repository.SuperStatementFilter) *gorm.DB {
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

var _ repository.SuperStatementRepository = (*SuperStatementRepository)(nil)
var _ repository.SuperChartRepository = (*SuperChartRepository)(nil)
