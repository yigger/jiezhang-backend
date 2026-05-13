package repository

import (
	"context"
	"time"
)

type StatisticsRepository interface {
	StatisticGroupDate(ctx context.Context, date time.Time, accountBookID int64) ([]CalendarDataItem, error)
}

type CalendarDataItem struct {
	Day    int     `json:"day"`
	Income float64 `json:"income"`
	Expend float64 `json:"expend"`
}
