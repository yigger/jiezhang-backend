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
	ID              int64     `gorm:"column:id"`
	Type            string    `gorm:"column:type"`
	Amount          float64   `gorm:"column:amount"`
	Description     string    `gorm:"column:description"`
	Remark          string    `gorm:"column:remark"`
	Mood            string    `gorm:"column:mood"`
	IconPath        string    `gorm:"column:icon_path"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
	CategoryName    string    `gorm:"column:category_name"`
	AssetName       string    `gorm:"column:asset_name"`
	Location        string    `gorm:"column:location"`
	Province        string    `gorm:"column:province"`
	City            string    `gorm:"column:city"`
	Street          string    `gorm:"column:street"`
	HasPic          bool      `gorm:"column:has_pic"`
	PayeeID         int64     `gorm:"column:payee_id"`
	PayeeName       string    `gorm:"column:payee_name"`
	TargetAssetName string    `gorm:"column:target_asset_name"`
}

func NewStatementRepository(db *gorm.DB) (*StatementRepository, error) {
	return &StatementRepository{db: db}, nil
}

func (r *StatementRepository) ListRowsWithRelations(ctx context.Context, filter repository.StatementListFilter) ([]repository.StatementListRowRecord, error) {
	query := r.db.WithContext(ctx).
		Table("statements s").
		Joins("INNER JOIN categories c ON c.id = s.category_id").
		Joins("LEFT JOIN assets a ON a.id = s.asset_id").
		Joins("LEFT JOIN payees p ON p.id = s.payee_id").
		Joins("LEFT JOIN account_book_collaborators abc ON abc.account_book_id = s.account_book_id AND abc.user_id = s.user_id").
		Joins("LEFT JOIN assets ta ON ta.id = s.target_asset_id")

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
		"s.mood AS mood",
		"s.description AS description",
		"COALESCE(abc.remark, '') AS remark",
		"s.created_at AS created_at",
		"s.updated_at AS updated_at",
		"s.location AS location",
		"s.province AS province",
		"s.city AS city",
		"s.street AS street",
		"EXISTS (SELECT 1 FROM user_assets ua WHERE ua.imageable_type = 'Statement' AND ua.type = 'StatementAvatar' AND ua.imageable_id = s.id) AS has_pic",
		"c.icon_path AS icon_path",
		"ta.name AS target_asset_name",
		"COALESCE(c.name, '') AS category_name",
		"COALESCE(a.name, '') AS asset_name",
		"COALESCE(s.payee_id, 0) AS payee_id",
		"COALESCE(p.name, '') AS payee_name",
	}, ", ")).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list statements with relations: %w", err)
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
			IconPath:        row.IconPath,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			CategoryName:    row.CategoryName,
			AssetName:       row.AssetName,
			Location:        row.Location,
			Province:        row.Province,
			City:            row.City,
			Street:          row.Street,
			HasPic:          row.HasPic,
			PayeeID:         row.PayeeID,
			PayeeName:       row.PayeeName,
			TargetAssetName: row.TargetAssetName,
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
