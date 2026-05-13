package repository

import (
	"context"
	"time"
)

type StatisticsRepository interface {
	StatisticGroupDate(ctx context.Context, date time.Time, accountBookID int64) ([]CalendarDataItem, error)
	OverviewHeader(ctx context.Context, date time.Time, accountBookID int64) (OverviewHeaderData, error)
}

type CalendarDataItem struct {
	Day    int     `json:"day"`
	Income float64 `json:"income"`
	Expend float64 `json:"expend"`
}

type OverviewHeaderData struct {
	TotalBalance float64 `json:"total"`
	Repay        float64 `json:"repay"`
	Transfer     float64 `json:"transfer"`
	Income       float64 `json:"income"`
	Expend       float64 `json:"expend"`
}
