package mysql

import (
	"context"
	"fmt"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
	"time"
)

type InsightsRepository struct{ db *gorm.DB }

func NewInsightsRepository(db *gorm.DB) *InsightsRepository { return &InsightsRepository{db: db} }

func (r *InsightsRepository) ListRows(ctx context.Context, bookID int64, start, end time.Time) ([]repo.InsightStatementRecord, error) {
	rows := make([]repo.InsightStatementRecord, 0)
	err := r.db.WithContext(ctx).Table("statements s").
		Select("s.*, COALESCE(c.name, '') AS category_name, COALESCE(NULLIF(ma.name, ''), p.name, '') AS merchant_name, COALESCE(NULLIF(ac.remark, ''), NULLIF(u.name, ''), NULLIF(u.nickname, ''), CONCAT('成员 ', s.user_id)) AS member_name").
		Joins("LEFT JOIN categories c ON c.id = s.category_id AND (c.account_book_id = s.account_book_id OR c.account_book_id = 0)").
		Joins("LEFT JOIN payees p ON p.id = s.payee_id AND p.account_book_id = s.account_book_id").
		Joins("LEFT JOIN insight_merchant_aliases ma ON ma.payee_id = p.id AND ma.account_book_id = s.account_book_id").
		Joins("LEFT JOIN users u ON u.id = s.user_id").
		Joins("LEFT JOIN account_book_collaborators ac ON ac.user_id = s.user_id AND ac.account_book_id = s.account_book_id").
		Where("s.account_book_id = ? AND s.created_at >= ? AND s.created_at < ?", bookID, start, end).
		Order("s.created_at DESC, s.id DESC").Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("load insight statements: %w", err)
	}
	return rows, nil
}
