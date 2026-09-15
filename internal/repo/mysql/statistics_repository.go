package mysql

import (
	"context"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type StatisticsRepository struct {
	db *gorm.DB
}

func NewStatisticsRepository(db *gorm.DB) *StatisticsRepository { return &StatisticsRepository{db: db} }

func (r *StatisticsRepository) StatisticGroupDate(ctx context.Context, date time.Time, accountBookID int64) ([]repo.CalendarDataItem, error) {
	var rows []repo.CalendarDataItem
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("day as date, SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income, SUM(CASE WHEN type = 'expend' THEN amount ELSE 0 END) as expend").
		Where("account_book_id = ? AND year = ? AND month = ?", accountBookID, date.Year(), int(date.Month())).
		Group("day").
		Order("day ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []repo.CalendarDataItem{}, nil
	}

	return rows, nil
}

func (r *StatisticsRepository) OverviewHeader(ctx context.Context, date time.Time, accountBookID int64) (repo.OverviewHeaderData, error) {
	var row repo.OverviewHeaderData
	err := r.db.WithContext(ctx).
		Table("statements").
		Select([]string{
			"SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income",
			"SUM(CASE WHEN type = 'expend' THEN amount ELSE 0 END) as expend",
			"SUM(CASE WHEN type = 'transfer' THEN amount ELSE 0 END) as transfer",
			"SUM(CASE WHEN type = 'repayment' THEN amount ELSE 0 END) as repayment",
		}).
		Where("account_book_id = ? AND year = ? AND month = ?", accountBookID, date.Year(), int(date.Month())).
		Take(&row).Error

	return row, err
}
