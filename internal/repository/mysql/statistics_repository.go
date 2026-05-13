package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"gorm.io/gorm"
)

type StatisticsRepository struct {
	db *gorm.DB
}

func NewStatisticsRepository(db *gorm.DB) (repository.StatisticsRepository, error) {
	return &StatisticsRepository{db: db}, nil
}

func (r *StatisticsRepository) StatisticGroupDate(ctx context.Context, date time.Time, accountBookID int64) ([]repository.CalendarDataItem, error) {
	var rows []repository.CalendarDataItem
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("day, SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income, SUM(CASE WHEN type = 'expend' THEN amount ELSE 0 END) as expend").
		Where("account_book_id = ? AND year = ? AND month = ?", accountBookID, date.Year(), int(date.Month())).
		Group("day").
		Order("day ASC").
		Scan(&rows).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return rows, nil
}
