package service

import (
	"context"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type StatisticsService struct {
	statisticsRepo repository.StatisticsRepository
}

func NewStatisticsService(statisticsRepo repository.StatisticsRepository) StatisticsService {
	return StatisticsService{
		statisticsRepo: statisticsRepo,
	}
}

func (s StatisticsService) GetCalendarData(context context.Context, date time.Time, accountBookID int64) ([]repository.CalendarDataItem, error) {
	res, err := s.statisticsRepo.StatisticGroupDate(context, date, accountBookID)

	return res, err
}
