package service

import (
	"context"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

type StatisticsService struct {
	statisticsRepo repository.StatisticsRepository
	statementsRepo repository.StatementQueryRepository
	rowMapper      statementdto.RowMapper
}

func NewStatisticsService(statisticsRepo repository.StatisticsRepository, statementsRepo repository.StatementQueryRepository, rowMapper statementdto.RowMapper) StatisticsService {
	return StatisticsService{
		statisticsRepo: statisticsRepo,
		statementsRepo: statementsRepo,
		rowMapper:      rowMapper,
	}
}

func (s StatisticsService) GetCalendarData(context context.Context, date time.Time, accountBookID int64) ([]repository.CalendarDataItem, error) {
	res, err := s.statisticsRepo.StatisticGroupDate(context, date, accountBookID)

	return res, err
}

func (s StatisticsService) GetOverviewHeader(context context.Context, date time.Time, accountBookID int64) (repository.OverviewHeaderData, error) {
	res, err := s.statisticsRepo.OverviewHeader(context, date, accountBookID)
	res.TotalBalance = res.Income - res.Expend - res.Repay
	return res, err
}

func (s StatisticsService) GetOverviewRate(context context.Context, statementType string, date time.Time, accountBookID int64) ([]statementdto.ListItem, error) {
	startDate, endDate := s.monthRange(date)
	filter := repository.StatementListFilter{
		AccountBookID: accountBookID,
		Type:          statementType,
		StartDate:     &startDate,
		EndDate:       &endDate,
		OrderBy:       "amount desc",
		Limit:         20,
		Offset:        0,
	}
	rows, err := s.statementsRepo.ListRowsWithRelations(context, filter)
	if err != nil {
		return nil, err
	}

	items := make([]statementdto.ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s StatisticsService) GetOverviewStatements(context context.Context, statementType string, date time.Time, accountBookID int64) ([]statementdto.ListItem, error) {
	startDate, endDate := s.monthRange(date)
	filter := repository.StatementListFilter{
		AccountBookID: accountBookID,
		Type:          statementType,
		StartDate:     &startDate,
		EndDate:       &endDate,
		OrderBy:       "created_at desc",
	}
	rows, err := s.statementsRepo.ListRowsWithRelations(context, filter)
	if err != nil {
		return nil, err
	}

	items := make([]statementdto.ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s StatisticsService) monthRange(date time.Time) (time.Time, time.Time) {
	location := date.Location()
	if location == nil {
		location = time.Local
	}

	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, location)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}
