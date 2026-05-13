package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type HomeRepository struct {
	db *gorm.DB
}

func NewHomeRepository(db *gorm.DB) (*HomeRepository, error) {
	return &HomeRepository{db: db}, nil
}

func (r *HomeRepository) SumExpendInRange(ctx context.Context, accountBookID int64, start time.Time, end time.Time) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ?", accountBookID).
		Where("type IN ?", []string{"expend", "repayment"}).
		Where("created_at BETWEEN ? AND ?", start, end).
		Scan(&row).Error
	return row.Amount, err
}

func (r *HomeRepository) GetAccountBookBudget(ctx context.Context, accountBookID int64) (float64, error) {
	var row struct {
		Budget float64 `gorm:"column:budget"`
	}
	err := r.db.WithContext(ctx).
		Table("account_books").
		Select("COALESCE(budget, 0) AS budget").
		Where("id = ?", accountBookID).
		Take(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Budget, nil
}

func (r *HomeRepository) FindLatestUnreadMessage(ctx context.Context, userID int64) (*repository.HomeMessageRecord, error) {
	var row struct {
		ID       int64  `gorm:"column:id"`
		Title    string `gorm:"column:title"`
		SubTitle string `gorm:"column:sub_title"`
	}
	err := r.db.WithContext(ctx).
		Table("messages").
		Select("id, title, sub_title").
		Where("target_id = ? AND already_read = 0", userID).
		Order("created_at DESC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &repository.HomeMessageRecord{ID: row.ID, Title: row.Title, SubTitle: row.SubTitle}, nil
}

func (r *HomeRepository) CountUserPersistDays(ctx context.Context, userID int64, accountBookID int64) (int64, error) {
	var persist int64
	err := r.db.WithContext(ctx).
		Table("statements").
		Where("user_id = ? AND account_book_id = ?", userID, accountBookID).
		Select("COUNT(DISTINCT DATE(created_at))").
		Scan(&persist).Error
	if err != nil {
		return 0, err
	}
	return persist, nil
}
