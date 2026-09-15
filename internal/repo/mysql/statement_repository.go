package mysql

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatementRepository struct {
	db *gorm.DB
}

func NewStatementRepository(db *gorm.DB) *StatementRepository { return &StatementRepository{db: db} }

func (r *statementMutation) Create(ctx context.Context, input tablemodel.Statement, effect repo.BalanceEffect) (int64, error) {
	var statementID int64
	err := func() error {
		tx := r.db.WithContext(ctx)
		assetAmount, err := r.getAssetAmountForUpdate(tx, input.AssetID, input.AccountBookID)
		if err != nil {
			return err
		}

		model := input
		model.UpdatedAt = time.Now()
		model.Residue = assetAmount + effect.Source
		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		if err := r.applyStatementEffect(tx, input.AccountBookID, effect, input.AssetID, input.TargetAssetID, 1); err != nil {
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
	}()
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
			return 0, repo.ErrStatementNotFound
		}
		return 0, err
	}
	return row.UserID, nil
}

func (r *statementMutation) UpdateByID(ctx context.Context, statementID int64, accountBookID int64, input tablemodel.Statement, oldEffect, effect repo.BalanceEffect) error {
	return func() error {
		tx := r.db.WithContext(ctx)
		oldStatement, err := r.getStatementForUpdate(tx, statementID, accountBookID)
		if err != nil {
			return err
		}

		if err := r.applyStatementEffect(tx, accountBookID, oldEffect, oldStatement.AssetID, oldStatement.TargetAssetID, -1); err != nil {
			return err
		}
		if err := r.applyStatementEffect(tx, accountBookID, effect, input.AssetID, input.TargetAssetID, 1); err != nil {
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
			"year":            input.Year,
			"month":           input.Month,
			"day":             input.Day,
			"time":            input.TimeText,
			"created_at":      input.CreatedAt,
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
			return repo.ErrStatementNotFound
		}
		return nil
	}()
}

func (r *statementMutation) DeleteByID(ctx context.Context, statementID int64, accountBookID int64, effect repo.BalanceEffect) error {
	return func() error {
		tx := r.db.WithContext(ctx)
		oldStatement, err := r.getStatementForUpdate(tx, statementID, accountBookID)
		if err != nil {
			return err
		}

		if err := r.applyStatementEffect(tx, accountBookID, effect, oldStatement.AssetID, oldStatement.TargetAssetID, -1); err != nil {
			return err
		}

		res := tx.Where("id = ? AND account_book_id = ?", statementID, accountBookID).Delete(&tablemodel.Statement{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return repo.ErrStatementNotFound
		}
		return nil
	}()
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
		return repo.ErrStatementNotFound
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
		return repo.ErrStatementAvatarNotFound
	}
	return nil
}

func (r *StatementRepository) getStatementForUpdate(tx *gorm.DB, statementID int64, accountBookID int64) (tablemodel.Statement, error) {
	var model tablemodel.Statement
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND account_book_id = ?", statementID, accountBookID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.Statement{}, repo.ErrStatementNotFound
		}
		return tablemodel.Statement{}, err
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

func (r *StatementRepository) applyStatementEffect(tx *gorm.DB, accountBookID int64, effect repo.BalanceEffect, sourceAssetID int64, targetAssetID *int64, direction int) error {
	sourceDelta := effect.Source * float64(direction)
	if err := r.updateAssetAmountByDelta(tx, sourceAssetID, accountBookID, sourceDelta); err != nil {
		return err
	}

	if targetAssetID != nil && *targetAssetID > 0 {
		targetDelta, ok := effect.Target, effect.HasTarget
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

func (r *StatementRepository) GetLatestCategoryAssetByType(ctx context.Context, accountBookID int64, statementType string) (*tablemodel.Statement, error) {
	var row tablemodel.Statement
	err := r.db.WithContext(ctx).
		Table("statements s").
		Select("s.*").
		Where("s.account_book_id = ? AND s.type = ?", accountBookID, statementType).
		Order("s.created_at DESC").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &row, nil
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

func (r *StatementRepository) ListAvatarRows(ctx context.Context, accountBookID int64) ([]tablemodel.UserAsset, error) {
	rows := make([]tablemodel.UserAsset, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN user_assets ua ON ua.imageable_type = 'Statement' AND ua.type = 'StatementAvatar' AND ua.imageable_id = s.id").
		Where("s.account_book_id = ?", accountBookID).
		Order("s.year DESC, s.month DESC, s.day DESC, s.created_at DESC, ua.id DESC").
		Select("ua.*").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *StatementRepository) ListAvatarsByStatementID(ctx context.Context, statementID int64) ([]tablemodel.UserAsset, error) {
	rows := make([]tablemodel.UserAsset, 0)
	err := r.db.WithContext(ctx).
		Table("user_assets ua").
		Joins("INNER JOIN statements s ON s.id = ua.imageable_id").
		Where("ua.imageable_type = 'Statement' AND ua.type = 'StatementAvatar' AND ua.imageable_id = ?", statementID).
		Order("ua.id ASC").
		Select("ua.*").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *StatementRepository) ListExportRows(ctx context.Context, filter repo.StatementExportFilter) ([]tablemodel.Statement, error) {
	if filter.Limit <= 0 || filter.Limit > 3000 {
		filter.Limit = 3000
	}

	rows := make([]tablemodel.Statement, 0)
	err := r.db.WithContext(ctx).
		Table("statements s").
		Where("s.account_book_id = ? AND s.created_at BETWEEN ? AND ?", filter.AccountBookID, filter.StartDate, filter.EndDate).
		Order("s.created_at ASC").
		Limit(filter.Limit).
		Select("s.*").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// Simple single-table row struct for ListSimpleRows.

// Batch lookup row structs.

// ListSimpleRows returns statement rows from a single-table query (no JOINs).
func (r *StatementRepository) ListSimpleRows(ctx context.Context, filter repo.StatementListFilter) ([]tablemodel.Statement, error) {
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
		return nil, fmt.Errorf("list simple statement rows: %w", err)
	}

	return rows, nil
}

// GetSimpleRowByID fetches a single statement row without JOINs.
func (r *StatementRepository) GetSimpleRowByID(ctx context.Context, statementID int64, accountBookID int64) (tablemodel.Statement, error) {
	var row tablemodel.Statement
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
			return tablemodel.Statement{}, repo.ErrStatementNotFound
		}
		return tablemodel.Statement{}, fmt.Errorf("get simple row by id: %w", err)
	}
	return row, nil
}

// BatchGetCategories fetches categories by primary keys, including parent info.
func (r *StatementRepository) BatchGetCategories(ctx context.Context, ids []int64) ([]tablemodel.Category, error) {
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
		return nil, fmt.Errorf("batch get categories: %w", err)
	}
	return rows, nil
}

// BatchGetAssets fetches assets by primary keys, including parent info.
func (r *StatementRepository) BatchGetAssets(ctx context.Context, ids []int64) ([]tablemodel.Asset, error) {
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
		return nil, fmt.Errorf("batch get assets: %w", err)
	}
	return rows, nil
}

// BatchGetPayees fetches payees by primary keys.
func (r *StatementRepository) BatchGetPayees(ctx context.Context, ids []int64) ([]tablemodel.Payee, error) {
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
		return nil, fmt.Errorf("batch get payees: %w", err)
	}
	items := make([]tablemodel.Payee, 0, len(rows))
	for _, row := range rows {
		items = append(items, tablemodel.Payee{ID: row.ID, Name: row.Name})
	}
	return items, nil
}

// BatchGetCollaboratorRemarks fetches collaborator remarks for given users in an account book.
func (r *StatementRepository) BatchGetCollaboratorRemarks(ctx context.Context, accountBookID int64, userIDs []int64) ([]tablemodel.AccountBookCollaborator, error) {
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
		return nil, fmt.Errorf("batch get collaborator remarks: %w", err)
	}
	items := make([]tablemodel.AccountBookCollaborator, 0, len(rows))
	for _, row := range rows {
		items = append(items, tablemodel.AccountBookCollaborator{UserID: row.UserID, Remark: row.Remark})
	}
	return items, nil
}

// BatchCheckHasPic checks which statement IDs have StatementAvatar records.
func (r *StatementRepository) BatchStatementAvatars(ctx context.Context, statementIDs []int64) ([]tablemodel.UserAsset, error) {
	if len(statementIDs) == 0 {
		return nil, nil
	}
	rows := make([]tablemodel.UserAsset, 0)
	err := r.db.WithContext(ctx).Table("user_assets").Select("imageable_id").
		Where("imageable_type = 'Statement' AND type = 'StatementAvatar' AND imageable_id IN ?", statementIDs).
		Group("imageable_id").Scan(&rows).Error
	return rows, err
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

func (r *StatementRepository) BatchGetStatements(ctx context.Context, ids []int64) ([]tablemodel.Statement, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []tablemodel.Statement
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}
