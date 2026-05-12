package service

import (
	"context"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type StatisticsService struct {
	statementRepo repository.StatementRepository
}

func NewStatisticsService(statementRepo repository.StatementRepository) StatisticsService {
	return StatisticsService{
		statementRepo: statementRepo,
	}
}

func (s StatisticsService) GetCalendarData(context context.Context, date time.Time, accountBookID int64) ([]repository.CalendarDataItem, error) {
	res, err := s.statementRepo.StatisticGroupDate(context, date, accountBookID)

	return res, err
}
