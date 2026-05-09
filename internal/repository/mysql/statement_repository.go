package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type StatementRepository struct {
	db *gorm.DB
}

type statementListRow struct {
	ID               int64     `gorm:"column:id"`
	Type             string    `gorm:"column:type"`
	Amount           float64   `gorm:"column:amount"`
	Description      string    `gorm:"column:description"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
	CategoryID       int64     `gorm:"column:category_id"`
	CategoryName     string    `gorm:"column:category_name"`
	ParentCategoryID int64     `gorm:"column:parent_category_id"`
	ParentCategory   string    `gorm:"column:parent_category"`
	AssetID          int64     `gorm:"column:asset_id"`
	AssetName        string    `gorm:"column:asset_name"`
	PayeeID          int64     `gorm:"column:payee_id"`
	PayeeName        string    `gorm:"column:payee_name"`
}

func NewStatementRepository(db *gorm.DB) (*StatementRepository, error) {
	return &StatementRepository{db: db}, nil
}

func (r *StatementRepository) ListWithRelations(ctx context.Context, filter repository.StatementListFilter) ([]repository.StatementListItem, error) {
	query := r.db.WithContext(ctx).
		Table("statements s").
		Joins("LEFT JOIN categories c ON c.id = s.category_id").
		Joins("LEFT JOIN categories pc ON pc.id = c.parent_id").
		Joins("LEFT JOIN assets a ON a.id = s.asset_id").
		Joins("LEFT JOIN payees p ON p.id = s.payee_id")

	if filter.AccountBookID > 0 {
		query = query.Where("s.account_book_id = ?", filter.AccountBookID)
	} else {
		query = query.Where("s.user_id = ?", filter.UserID)
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

	if len(filter.ParentCategoryIDs) > 0 {
		query = query.Where("c.parent_id IN ?", filter.ParentCategoryIDs)
	}

	if len(filter.ExceptIDs) > 0 {
		query = query.Where("s.id NOT IN ?", filter.ExceptIDs)
	}

	orderExpr := mapOrderBy(filter.OrderBy)
	query = query.Order(orderExpr)

	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	query = query.Limit(filter.Limit).Offset(filter.Offset)

	var rows []statementListRow
	err := query.Select(strings.Join([]string{
		"s.id AS id",
		"s.type AS type",
		"s.amount AS amount",
		"s.description AS description",
		"s.created_at AS created_at",
		"s.updated_at AS updated_at",
		"s.category_id AS category_id",
		"COALESCE(c.name, '') AS category_name",
		"COALESCE(c.parent_id, 0) AS parent_category_id",
		"COALESCE(pc.name, '') AS parent_category",
		"s.asset_id AS asset_id",
		"COALESCE(a.name, '') AS asset_name",
		"COALESCE(s.payee_id, 0) AS payee_id",
		"COALESCE(p.name, '') AS payee_name",
	}, ", ")).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list statements with relations: %w", err)
	}

	items := make([]repository.StatementListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.StatementListItem{
			ID:               row.ID,
			Type:             row.Type,
			Amount:           row.Amount,
			Description:      row.Description,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
			CategoryID:       row.CategoryID,
			CategoryName:     row.CategoryName,
			ParentCategoryID: row.ParentCategoryID,
			ParentCategory:   row.ParentCategory,
			AssetID:          row.AssetID,
			AssetName:        row.AssetName,
			PayeeID:          row.PayeeID,
			PayeeName:        row.PayeeName,
		})
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
