package mysql

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type StatementRepository struct {
	db *gorm.DB
}

type statementBase struct {
	ID            int64     `gorm:"column:id"`
	Type          string    `gorm:"column:type"`
	Amount        float64   `gorm:"column:amount"`
	CategoryID    int64     `gorm:"column:category_id"`
	AssetID       int64     `gorm:"column:asset_id"`
	TargetAssetID int64     `gorm:"column:target_asset_id"`
	TargetObject  string    `gorm:"column:target_object"`
	Description   string    `gorm:"column:description"`
	Remark        string    `gorm:"column:remark"`
	Mood          string    `gorm:"column:mood"`
	IconPath      string    `gorm:"column:icon_path"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
	CategoryName  string    `gorm:"column:category_name"`
	AssetName     string    `gorm:"column:asset_name"`
	Location      string    `gorm:"column:location"`
	Nation        string    `gorm:"column:nation"`
	Province      string    `gorm:"column:province"`
	City          string    `gorm:"column:city"`
	District      string    `gorm:"column:district"`
	Street        string    `gorm:"column:street"`
}

type statementListRow struct {
	ID                 int64     `gorm:"column:id"`
	UserID             int64     `gorm:"column:user_id"`
	Type               string    `gorm:"column:type"`
	Amount             float64   `gorm:"column:amount"`
	CategoryID         int64     `gorm:"column:category_id"`
	AssetID            int64     `gorm:"column:asset_id"`
	TargetAssetID      int64     `gorm:"column:target_asset_id"`
	TargetObject       string    `gorm:"column:target_object"`
	Description        string    `gorm:"column:description"`
	Remark             string    `gorm:"column:remark"`
	Mood               string    `gorm:"column:mood"`
	IconPath           string    `gorm:"column:icon_path"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
	CategoryName       string    `gorm:"column:category_name"`
	AssetName          string    `gorm:"column:asset_name"`
	Location           string    `gorm:"column:location"`
	Nation             string    `gorm:"column:nation"`
	Province           string    `gorm:"column:province"`
	City               string    `gorm:"column:city"`
	District           string    `gorm:"column:district"`
	Street             string    `gorm:"column:street"`
	HasPic             bool      `gorm:"column:has_pic"`
	Residue            float64   `gorm:"column:residue"`
	PayeeID            int64     `gorm:"column:payee_id"`
	PayeeName          string    `gorm:"column:payee_name"`
	TargetAssetName    string    `gorm:"column:target_asset_name"`
	CategoryParentName string    `gorm:"column:category_parent_name"`
	AssetParentName    string    `gorm:"column:asset_parent_name"`
}

type defaultCategoryAssetRow struct {
	CategoryID   int64  `gorm:"column:category_id"`
	AssetID      int64  `gorm:"column:asset_id"`
	CategoryName string `gorm:"column:category_name"`
	AssetName    string `gorm:"column:asset_name"`
}

type statementAvatarRow struct {
	StatementID int64  `gorm:"column:statement_id"`
	UserID      int64  `gorm:"column:user_id"`
	Year        int    `gorm:"column:year"`
	Month       int    `gorm:"column:month"`
	AvatarID    int64  `gorm:"column:avatar_id"`
	AvatarPath  string `gorm:"column:avatar_path"`
}

type statementExportRow struct {
	CategoryName       string    `gorm:"column:category_name"`
	ParentCategoryName string    `gorm:"column:parent_category_name"`
	Type               string    `gorm:"column:type"`
	AssetName          string    `gorm:"column:asset_name"`
	Description        string    `gorm:"column:description"`
	Amount             float64   `gorm:"column:amount"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

type statementMutationModel struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID        int64     `gorm:"column:user_id"`
	AccountBookID int64     `gorm:"column:account_book_id"`
	CategoryID    int64     `gorm:"column:category_id"`
	AssetID       int64     `gorm:"column:asset_id"`
	TargetAssetID *int64    `gorm:"column:target_asset_id"`
	PayeeID       *int64    `gorm:"column:payee_id"`
	Type          string    `gorm:"column:type"`
	Amount        float64   `gorm:"column:amount"`
	Refund        float64   `gorm:"column:refund"`
	Residue       float64   `gorm:"column:residue"`
	Mood          string    `gorm:"column:mood"`
	Description   string    `gorm:"column:description"`
	TargetObject  string    `gorm:"column:target_object"`
	Year          int       `gorm:"column:year"`
	Month         int       `gorm:"column:month"`
	Day           int       `gorm:"column:day"`
	TimeText      string    `gorm:"column:time"`
	Location      string    `gorm:"column:location"`
	Nation        string    `gorm:"column:nation"`
	Province      string    `gorm:"column:province"`
	City          string    `gorm:"column:city"`
	District      string    `gorm:"column:district"`
	Street        string    `gorm:"column:street"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (statementMutationModel) TableName() string {
	return "statements"
}

func NewStatementRepository(db *gorm.DB) (*StatementRepository, error) {
	return &StatementRepository{db: db}, nil
}

func (r *StatementRepository) Create(ctx context.Context, input repository.StatementWriteRecord) (int64, error) {
	var statementID int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		assetAmount, err := r.getAssetAmountForUpdate(tx, input.AssetID, input.AccountBookID)
		if err != nil {
			return err
		}

		model := statementMutationModel{
			UserID:        input.UserID,
			AccountBookID: input.AccountBookID,
			CategoryID:    input.CategoryID,
			AssetID:       input.AssetID,
			TargetAssetID: input.TargetAssetID,
			PayeeID:       input.PayeeID,
			Type:          input.Type,
			Amount:        input.Amount,
			Mood:          input.Mood,
			Description:   input.Description,
			TargetObject:  input.TargetObject,
			Year:          input.OccurredAt.Year(),
			Month:         int(input.OccurredAt.Month()),
			Day:           input.OccurredAt.Day(),
			TimeText:      input.OccurredAt.Format("15:04"),
			Location:      input.Location,
			Nation:        input.Nation,
			Province:      input.Province,
			City:          input.City,
			District:      input.District,
			Street:        input.Street,
			CreatedAt:     input.OccurredAt,
			UpdatedAt:     time.Now(),
			Residue:       calculateResidue(assetAmount, input.Type, input.Amount),
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		if err := r.applyStatementEffect(tx, input.AccountBookID, input.Type, input.Amount, input.AssetID, input.TargetAssetID, 1); err != nil {
			return err
		}

		if input.Type == "income" || input.Type == "expend" {
			if err := tx.Table("categories").Where("id = ?", input.CategoryID).UpdateColumn("frequent", gorm.Expr("frequent + ?", 1)).Error; err != nil {
				return err
			}
			if err := tx.Table("assets").Where("id = ?", input.AssetID).UpdateColumn("frequent", gorm.Expr("frequent + ?", 1)).Error; err != nil {
				return err
			}
		}

		statementID = model.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return statementID, nil
}

func (r *StatementRepository) GetOwnerID(ctx context.Context, statementID int64, accountBookID int64) (int64, error) {
	var row struct {
		UserID int64 `gorm:"column:user_id"`
	}
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("user_id").
		Where("id = ? AND account_book_id = ?", statementID, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, repository.ErrStatementNotFound
		}
		return 0, err
	}
	return row.UserID, nil
}

func (r *StatementRepository) UpdateByID(ctx context.Context, statementID int64, accountBookID int64, input repository.StatementWriteRecord) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		oldStatement, err := r.getStatementForUpdate(tx, statementID, accountBookID)
		if err != nil {
			return err
		}

		if err := r.applyStatementEffect(tx, accountBookID, oldStatement.Type, oldStatement.Amount, oldStatement.AssetID, oldStatement.TargetAssetID, -1); err != nil {
			return err
		}
		if err := r.applyStatementEffect(tx, accountBookID, input.Type, input.Amount, input.AssetID, input.TargetAssetID, 1); err != nil {
			return err
		}

		updates := map[string]interface{}{
			"category_id":     input.CategoryID,
			"asset_id":        input.AssetID,
			"target_asset_id": input.TargetAssetID,
			"payee_id":        input.PayeeID,
			"type":            input.Type,
			"amount":          input.Amount,
			"mood":            input.Mood,
			"description":     input.Description,
			"target_object":   input.TargetObject,
			"year":            input.OccurredAt.Year(),
			"month":           int(input.OccurredAt.Month()),
			"day":             input.OccurredAt.Day(),
			"time":            input.OccurredAt.Format("15:04"),
			"created_at":      input.OccurredAt,
			"updated_at":      time.Now(),
			"location":        input.Location,
			"nation":          input.Nation,
			"province":        input.Province,
			"city":            input.City,
			"district":        input.District,
			"street":          input.Street,
		}
		res := tx.Table("statements").Where("id = ? AND account_book_id = ?", statementID, accountBookID).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return repository.ErrStatementNotFound
		}
		return nil
	})
}

func (r *StatementRepository) DeleteByID(ctx context.Context, statementID int64, accountBookID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		oldStatement, err := r.getStatementForUpdate(tx, statementID, accountBookID)
		if err != nil {
			return err
		}

		if err := r.applyStatementEffect(tx, accountBookID, oldStatement.Type, oldStatement.Amount, oldStatement.AssetID, oldStatement.TargetAssetID, -1); err != nil {
			return err
		}

		res := tx.Where("id = ? AND account_book_id = ?", statementID, accountBookID).Delete(&statementMutationModel{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return repository.ErrStatementNotFound
		}
		return nil
	})
}

func (r *StatementRepository) DeleteAvatarByID(ctx context.Context, accountBookID int64, statementID int64, avatarID int64) error {
	var count int64
	if err := r.db.WithContext(ctx).
		Table("statements").
		Where("id = ? AND account_book_id = ?", statementID, accountBookID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return repository.ErrStatementNotFound
	}

	res := r.db.WithContext(ctx).
		Table("user_assets ua").
		Joins("INNER JOIN statements s ON s.id = ua.imageable_id AND ua.imageable_type = 'Statement' AND ua.type = 'StatementAvatar'").
		Where("ua.id = ? AND ua.imageable_id = ? AND s.account_book_id = ?", avatarID, statementID, accountBookID).
		Delete(&struct{}{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrStatementAvatarNotFound
	}
	return nil
}

func (r *StatementRepository) getStatementForUpdate(tx *gorm.DB, statementID int64, accountBookID int64) (statementMutationModel, error) {
	var model statementMutationModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND account_book_id = ?", statementID, accountBookID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return statementMutationModel{}, repository.ErrStatementNotFound
		}
		return statementMutationModel{}, err
	}
	return model, nil
}

func (r *StatementRepository) getAssetAmountForUpdate(tx *gorm.DB, assetID int64, accountBookID int64) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	err := tx.Table("assets").Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("amount").
		Where("id = ? AND account_book_id = ?", assetID, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("asset %d not found", assetID)
		}
		return 0, err
	}
	return row.Amount, nil
}

func (r *StatementRepository) applyStatementEffect(tx *gorm.DB, accountBookID int64, statementType string, amount float64, sourceAssetID int64, targetAssetID *int64, direction int) error {
	sourceDelta := calcSourceAssetDelta(statementType, amount) * float64(direction)
	if err := r.updateAssetAmountByDelta(tx, sourceAssetID, accountBookID, sourceDelta); err != nil {
		return err
	}

	if targetAssetID != nil && *targetAssetID > 0 {
		targetDelta, ok := calcTargetAssetDelta(statementType, amount)
		if ok {
			if err := r.updateAssetAmountByDelta(tx, *targetAssetID, accountBookID, targetDelta*float64(direction)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *StatementRepository) updateAssetAmountByDelta(tx *gorm.DB, assetID int64, accountBookID int64, delta float64) error {
	if delta == 0 {
		return nil
	}
	res := tx.Table("assets").
		Where("id = ? AND account_book_id = ?", assetID, accountBookID).
		UpdateColumn("amount", gorm.Expr("amount + ?", delta))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("asset %d not found", assetID)
	}
	return nil
}

func calcSourceAssetDelta(statementType string, amount float64) float64 {
	switch statementType {
	case "income", "loan_in":
		return amount
	case "expend", "transfer", "repayment", "loan_out", "reimburse", "payment_proxy":
		return -amount
	default:
		return 0
	}
}

func calcTargetAssetDelta(statementType string, amount float64) (float64, bool) {
	switch statementType {
	case "transfer":
		return amount, true
	case "repayment":
		return -amount, true
	default:
		return 0, false
	}
}

func calculateResidue(assetAmount float64, statementType string, amount float64) float64 {
	switch statementType {
	case "income", "loan_in":
		return assetAmount + amount
	case "expend", "transfer", "repayment", "loan_out", "reimburse", "payment_proxy":
		return assetAmount - amount
	default:
		return assetAmount
	}
}

func (r *StatementRepository) ListRowsWithRelations(ctx context.Context, filter repository.StatementListFilter) ([]repository.StatementListRowRecord, error) {
	query := r.baseStatementListQuery(ctx)
	if filter.AccountBookID > 0 {
		query = query.Where("s.account_book_id = ?", filter.AccountBookID)
	} else {
		query = query.Where("s.user_id = ?", filter.UserID)
	}

	if filter.Type != "" {
		query = query.Where("s.type = ?", filter.Type)
	}
	if filter.AssetID > 0 {
		query = query.Where("s.asset_id = ?", filter.AssetID)
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		endOfDay := time.Date(
			filter.EndDate.Year(),
			filter.EndDate.Month(),
			filter.EndDate.Day(),
			23, 59, 59, int(time.Second-time.Nanosecond),
			filter.EndDate.Location(),
		)
		query = query.Where("s.created_at BETWEEN ? AND ?", *filter.StartDate, endOfDay)
	}

	if filter.Keyword != "" {
		amount, err := strconv.ParseFloat(filter.Keyword, 64)
		if err == nil {
			query = query.Where("s.amount = ?", amount)
		} else {
			likePattern := "%" + strings.TrimSpace(filter.Keyword) + "%"
			query = query.Where("s.description LIKE ?", likePattern)
		}
	}

	if len(filter.ParentCategoryIDs) > 0 {
		query = query.Where("c.parent_id IN ?", filter.ParentCategoryIDs)
	}
	if len(filter.ExceptIDs) > 0 {
		query = query.Where("s.id NOT IN ?", filter.ExceptIDs)
	}
	query = query.Order(mapOrderBy(filter.OrderBy))

	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 1000
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	query = query.Limit(filter.Limit).Offset(filter.Offset)

	rows, err := r.scanStatementRows(query)
	if err != nil {
		return nil, fmt.Errorf("list statements with relations: %w", err)
	}
	return rows, nil
}

func (r *StatementRepository) GetRowByIDWithRelations(ctx context.Context, statementID int64, accountBookID int64) (repository.StatementListRowRecord, error) {
	query := r.baseStatementListQuery(ctx).
		Where("s.id = ? AND s.account_book_id = ?", statementID, accountBookID).
		Limit(1)
	rows, err := r.scanStatementRows(query)
	if err != nil {
		return repository.StatementListRowRecord{}, err
	}
	if len(rows) == 0 {
		return repository.StatementListRowRecord{}, repository.ErrStatementNotFound
	}
	return rows[0], nil
}

func (r *StatementRepository) GetLatestCategoryAssetByType(ctx context.Context, accountBookID int64, statementType string) (*repository.StatementDefaultCategoryAssetRecord, error) {
	var row defaultCategoryAssetRow
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("LEFT JOIN categories c ON c.id = s.category_id").
		Joins("LEFT JOIN categories cp ON cp.id = c.parent_id").
		Joins("LEFT JOIN assets a ON a.id = s.asset_id").
		Joins("LEFT JOIN assets ap ON ap.id = a.parent_id").
		Select(strings.Join([]string{
			"s.category_id AS category_id",
			"s.asset_id AS asset_id",
			"CASE WHEN c.parent_id > 0 THEN CONCAT(COALESCE(cp.name, ''), ' -> ', COALESCE(c.name, '')) ELSE COALESCE(c.name, '') END AS category_name",
			"CASE WHEN a.parent_id > 0 THEN CONCAT(COALESCE(ap.name, ''), ' -> ', COALESCE(a.name, '')) ELSE COALESCE(a.name, '') END AS asset_name",
		}, ", ")).
		Where("s.account_book_id = ? AND s.type = ?", accountBookID, statementType).
		Order("s.created_at DESC").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &repository.StatementDefaultCategoryAssetRecord{
		CategoryID:   row.CategoryID,
		AssetID:      row.AssetID,
		CategoryName: row.CategoryName,
		AssetName:    row.AssetName,
	}, nil
}

func (r *StatementRepository) ListDistinctTargetObjectsByType(ctx context.Context, accountBookID int64, statementType string) ([]string, error) {
	if strings.TrimSpace(statementType) == "" {
		return []string{}, nil
	}

	rows := make([]string, 0)
	err := r.db.WithContext(ctx).
		Table("statements").
		Where("account_book_id = ? AND type = ?", accountBookID, statementType).
		Distinct("target_object").
		Pluck("target_object", &rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *StatementRepository) ListAvatarRows(ctx context.Context, accountBookID int64) ([]repository.StatementAvatarRowRecord, error) {
	rows := make([]statementAvatarRow, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN user_assets ua ON ua.imageable_type = 'Statement' AND ua.type = 'StatementAvatar' AND ua.imageable_id = s.id").
		Where("s.account_book_id = ?", accountBookID).
		Order("s.year DESC, s.month DESC, s.day DESC, s.created_at DESC, ua.id DESC").
		Select(strings.Join([]string{
			"s.id AS statement_id",
				"s.user_id AS user_id",
			"s.year AS year",
			"s.month AS month",
			"ua.id AS avatar_id",
			"ua.path AS avatar_path",
		}, ", ")).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.StatementAvatarRowRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementAvatarRowRecord{
			StatementID: row.StatementID,
			UserID:      row.UserID,
			Year:        row.Year,
			Month:       row.Month,
			AvatarID:    row.AvatarID,
			AvatarPath:  row.AvatarPath,
		})
	}
	return items, nil
}

func (r *StatementRepository) ListAvatarsByStatementID(ctx context.Context, statementID int64) ([]repository.StatementAvatarRowRecord, error) {
	rows := make([]statementAvatarRow, 0)
	err := r.db.WithContext(ctx).
		Table("user_assets ua").
		Joins("INNER JOIN statements s ON s.id = ua.imageable_id").
		Where("ua.imageable_type = 'Statement' AND ua.type = 'StatementAvatar' AND ua.imageable_id = ?", statementID).
		Order("ua.id ASC").
		Select("ua.id AS avatar_id, ua.path AS avatar_path, s.user_id AS user_id, s.id AS statement_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.StatementAvatarRowRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementAvatarRowRecord{
			StatementID: row.StatementID,
			UserID:      row.UserID,
			AvatarID:    row.AvatarID,
			AvatarPath:  row.AvatarPath,
		})
	}
	return items, nil
}

func (r *StatementRepository) ListExportRows(ctx context.Context, filter repository.StatementExportFilter) ([]repository.StatementExportRowRecord, error) {
	if filter.Limit <= 0 || filter.Limit > 3000 {
		filter.Limit = 3000
	}

	rows := make([]statementExportRow, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("LEFT JOIN categories c ON c.id = s.category_id").
		Joins("LEFT JOIN categories cp ON cp.id = c.parent_id").
		Joins("LEFT JOIN assets a ON a.id = s.asset_id").
		Where("s.account_book_id = ? AND s.created_at BETWEEN ? AND ?", filter.AccountBookID, filter.StartDate, filter.EndDate).
		Order("s.created_at ASC").
		Limit(filter.Limit).
		Select(strings.Join([]string{
			"COALESCE(c.name, '') AS category_name",
			"COALESCE(cp.name, '') AS parent_category_name",
			"s.type AS type",
			"COALESCE(a.name, '') AS asset_name",
			"COALESCE(s.description, '') AS description",
			"s.amount AS amount",
			"s.created_at AS created_at",
			"s.updated_at AS updated_at",
		}, ", ")).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.StatementExportRowRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementExportRowRecord{
			CategoryName:       row.CategoryName,
			ParentCategoryName: row.ParentCategoryName,
			Type:               row.Type,
			AssetName:          row.AssetName,
			Description:        row.Description,
			Amount:             row.Amount,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *StatementRepository) baseStatementListQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN categories c ON c.id = s.category_id").
		Joins("LEFT JOIN categories cp ON cp.id = c.parent_id").
		Joins("LEFT JOIN assets a ON a.id = s.asset_id").
		Joins("LEFT JOIN assets ap ON ap.id = a.parent_id").
		Joins("LEFT JOIN payees p ON p.id = s.payee_id").
		Joins("LEFT JOIN account_book_collaborators abc ON abc.account_book_id = s.account_book_id AND abc.user_id = s.user_id").
		Joins("LEFT JOIN assets ta ON ta.id = s.target_asset_id")
}

func (r *StatementRepository) scanStatementRows(query *gorm.DB) ([]repository.StatementListRowRecord, error) {
	var rows []statementListRow
	err := query.Select(strings.Join([]string{
		"s.id AS id",
		"s.user_id AS user_id",
		"s.type AS type",
		"s.amount AS amount",
		"s.mood AS mood",
		"s.category_id AS category_id",
		"s.asset_id AS asset_id",
		"s.description AS description",
		"COALESCE(abc.remark, '') AS remark",
		"s.created_at AS created_at",
		"s.updated_at AS updated_at",
		"s.location AS location",
		"s.nation AS nation",
		"s.province AS province",
		"s.city AS city",
		"s.district AS district",
		"s.street AS street",
		"COALESCE(s.target_asset_id, 0) AS target_asset_id",
		"COALESCE(s.target_object, '') AS target_object",
		"EXISTS (SELECT 1 FROM user_assets ua WHERE ua.imageable_type = 'Statement' AND ua.type = 'StatementAvatar' AND ua.imageable_id = s.id) AS has_pic",
		"COALESCE(s.residue, 0) AS residue",
		"c.icon_path AS icon_path",
		"ta.name AS target_asset_name",
		"COALESCE(c.name, '') AS category_name",
		"COALESCE(a.name, '') AS asset_name",
		"COALESCE(cp.name, '') AS category_parent_name",
		"COALESCE(ap.name, '') AS asset_parent_name",
		"COALESCE(s.payee_id, 0) AS payee_id",
		"COALESCE(p.name, '') AS payee_name",
	}, ", ")).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.StatementListRowRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementListRowRecord{
			ID:                 row.ID,
			UserID:             row.UserID,
			Type:               row.Type,
			Amount:             row.Amount,
			Description:        row.Description,
			Remark:             row.Remark,
			Mood:               row.Mood,
			IconPath:           row.IconPath,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
			CategoryName:       row.CategoryName,
			AssetName:          row.AssetName,
			Location:           row.Location,
			Nation:             row.Nation,
			Province:           row.Province,
			City:               row.City,
			District:           row.District,
			Street:             row.Street,
			HasPic:             row.HasPic,
			Residue:            row.Residue,
			PayeeID:            row.PayeeID,
			PayeeName:          row.PayeeName,
			TargetAssetID:      row.TargetAssetID,
			TargetAssetName:    row.TargetAssetName,
			TargetObject:       row.TargetObject,
			CategoryID:         row.CategoryID,
			AssetID:            row.AssetID,
			CategoryParentName: row.CategoryParentName,
			AssetParentName:    row.AssetParentName,
		})
	}
	return items, nil
}

// Simple single-table row struct for ListSimpleRows.
type statementSimpleRow struct {
	ID            int64     `gorm:"column:id"`
	UserID        int64     `gorm:"column:user_id"`
	Type          string    `gorm:"column:type"`
	Amount        float64   `gorm:"column:amount"`
	Description   string    `gorm:"column:description"`
	CategoryID    int64     `gorm:"column:category_id"`
	AssetID       int64     `gorm:"column:asset_id"`
	TargetAssetID int64     `gorm:"column:target_asset_id"`
	TargetObject  string    `gorm:"column:target_object"`
	PayeeID       int64     `gorm:"column:payee_id"`
	Mood          string    `gorm:"column:mood"`
	Residue       float64   `gorm:"column:residue"`
	Location      string    `gorm:"column:location"`
	Nation        string    `gorm:"column:nation"`
	Province      string    `gorm:"column:province"`
	City          string    `gorm:"column:city"`
	District      string    `gorm:"column:district"`
	Street        string    `gorm:"column:street"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// Batch lookup row structs.
type categoryBatchRow struct {
	ID         int64  `gorm:"column:id"`
	Name       string `gorm:"column:name"`
	IconPath   string `gorm:"column:icon_path"`
	ParentID   int64  `gorm:"column:parent_id"`
	ParentName string `gorm:"column:parent_name"`
}

type assetBatchRow struct {
	ID         int64  `gorm:"column:id"`
	Name       string `gorm:"column:name"`
	IconPath   string `gorm:"column:icon_path"`
	ParentID   int64  `gorm:"column:parent_id"`
	ParentName string `gorm:"column:parent_name"`
}

type payeeBatchRow struct {
	ID   int64  `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

type collaboratorRemarkRow struct {
	UserID int64  `gorm:"column:user_id"`
	Remark string `gorm:"column:remark"`
}

type hasPicRow struct {
	StatementID int64 `gorm:"column:statement_id"`
	HasPic      bool  `gorm:"column:has_pic"`
}

// ListSimpleRows returns statement rows from a single-table query (no JOINs).
func (r *StatementRepository) ListSimpleRows(ctx context.Context, filter repository.StatementListFilter) ([]repository.StatementSimpleRowRecord, error) {
	query := r.db.WithContext(ctx).Table("statements s")

	if filter.AccountBookID > 0 {
		query = query.Where("s.account_book_id = ?", filter.AccountBookID)
	} else {
		query = query.Where("s.user_id = ?", filter.UserID)
	}

	if filter.Type != "" {
		query = query.Where("s.type = ?", filter.Type)
	}
	if filter.AssetID > 0 {
		query = query.Where("s.asset_id = ?", filter.AssetID)
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		endOfDay := time.Date(
			filter.EndDate.Year(),
			filter.EndDate.Month(),
			filter.EndDate.Day(),
			23, 59, 59, int(time.Second-time.Nanosecond),
			filter.EndDate.Location(),
		)
		query = query.Where("s.created_at BETWEEN ? AND ?", *filter.StartDate, endOfDay)
	}

	if filter.Keyword != "" {
		amount, err := strconv.ParseFloat(filter.Keyword, 64)
		if err == nil {
			query = query.Where("s.amount = ?", amount)
		} else {
			likePattern := "%" + strings.TrimSpace(filter.Keyword) + "%"
			query = query.Where("s.description LIKE ?", likePattern)
		}
	}

	if len(filter.ParentCategoryIDs) > 0 {
		query = query.Where("s.category_id IN (SELECT id FROM categories WHERE parent_id IN ?)", filter.ParentCategoryIDs)
	}
	if len(filter.ExceptIDs) > 0 {
		query = query.Where("s.id NOT IN ?", filter.ExceptIDs)
	}
	query = query.Order(mapOrderBy(filter.OrderBy))

	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 1000
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	query = query.Limit(filter.Limit).Offset(filter.Offset)

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
		return nil, fmt.Errorf("list simple statement rows: %w", err)
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

// GetSimpleRowByID fetches a single statement row without JOINs.
func (r *StatementRepository) GetSimpleRowByID(ctx context.Context, statementID int64, accountBookID int64) (repository.StatementSimpleRowRecord, error) {
	var row statementSimpleRow
	err := r.db.WithContext(ctx).
		Table("statements").
		Select(strings.Join([]string{
			"id AS id",
			"user_id AS user_id",
			"type AS type",
			"amount AS amount",
			"COALESCE(description, '') AS description",
			"category_id AS category_id",
			"asset_id AS asset_id",
			"COALESCE(target_asset_id, 0) AS target_asset_id",
			"COALESCE(target_object, '') AS target_object",
			"COALESCE(payee_id, 0) AS payee_id",
			"COALESCE(mood, '') AS mood",
			"COALESCE(residue, 0) AS residue",
			"COALESCE(location, '') AS location",
			"COALESCE(nation, '') AS nation",
			"COALESCE(province, '') AS province",
			"COALESCE(city, '') AS city",
			"COALESCE(district, '') AS district",
			"COALESCE(street, '') AS street",
			"created_at AS created_at",
			"updated_at AS updated_at",
		}, ", ")).
		Where("id = ? AND account_book_id = ?", statementID, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.StatementSimpleRowRecord{}, repository.ErrStatementNotFound
		}
		return repository.StatementSimpleRowRecord{}, fmt.Errorf("get simple row by id: %w", err)
	}
	return repository.StatementSimpleRowRecord{
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
	}, nil
}

// BatchGetCategories fetches categories by primary keys, including parent info.
func (r *StatementRepository) BatchGetCategories(ctx context.Context, ids []int64) ([]repository.CategoryBatchRecord, error) {
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
		return nil, fmt.Errorf("batch get categories: %w", err)
	}
	items := make([]repository.CategoryBatchRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.CategoryBatchRecord{
			ID:         row.ID,
			Name:       row.Name,
			IconPath:   row.IconPath,
			ParentID:   row.ParentID,
			ParentName: row.ParentName,
		})
	}
	return items, nil
}

// BatchGetAssets fetches assets by primary keys, including parent info.
func (r *StatementRepository) BatchGetAssets(ctx context.Context, ids []int64) ([]repository.AssetBatchRecord, error) {
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
		return nil, fmt.Errorf("batch get assets: %w", err)
	}
	items := make([]repository.AssetBatchRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.AssetBatchRecord{
			ID:         row.ID,
			Name:       row.Name,
			IconPath:   row.IconPath,
			ParentID:   row.ParentID,
			ParentName: row.ParentName,
		})
	}
	return items, nil
}

// BatchGetPayees fetches payees by primary keys.
func (r *StatementRepository) BatchGetPayees(ctx context.Context, ids []int64) ([]repository.PayeeBatchRecord, error) {
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
		return nil, fmt.Errorf("batch get payees: %w", err)
	}
	items := make([]repository.PayeeBatchRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.PayeeBatchRecord{ID: row.ID, Name: row.Name})
	}
	return items, nil
}

// BatchGetCollaboratorRemarks fetches collaborator remarks for given users in an account book.
func (r *StatementRepository) BatchGetCollaboratorRemarks(ctx context.Context, accountBookID int64, userIDs []int64) ([]repository.CollaboratorRemarkRecord, error) {
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
		return nil, fmt.Errorf("batch get collaborator remarks: %w", err)
	}
	items := make([]repository.CollaboratorRemarkRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.CollaboratorRemarkRecord{UserID: row.UserID, Remark: row.Remark})
	}
	return items, nil
}

// BatchCheckHasPic checks which statement IDs have StatementAvatar records.
func (r *StatementRepository) BatchCheckHasPic(ctx context.Context, statementIDs []int64) ([]repository.StatementHasPicRecord, error) {
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
		return nil, fmt.Errorf("batch check has pic: %w", err)
	}
	items := make([]repository.StatementHasPicRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementHasPicRecord{StatementID: row.StatementID, HasPic: row.HasPic})
	}
	return items, nil
}

func mapOrderBy(orderBy string) string {
	switch strings.TrimSpace(strings.ToLower(orderBy)) {
	case "created_at":
		return "s.created_at DESC"
	case "updated_at":
		return "s.updated_at DESC"
	case "amount":
		return "s.amount DESC"
	default:
		return "s.created_at DESC"
	}
}
