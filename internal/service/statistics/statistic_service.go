package statistics

import (
	"context"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	statementservice "github.com/yigger/jiezhang-backend/internal/service/statement"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type StatisticsService struct {
	statisticsRepo repo.StatisticsRepository
	statementsRepo repo.StatementQueryRepository
	rowMapper      statementservice.RowMapper
}

func NewStatisticsService(statisticsRepo repo.StatisticsRepository, statementsRepo repo.StatementQueryRepository, rowMapper statementservice.RowMapper) StatisticsService {
	return StatisticsService{
		statisticsRepo: statisticsRepo,
		statementsRepo: statementsRepo,
		rowMapper:      rowMapper,
	}
}

func (s StatisticsService) GetCalendarData(context context.Context, date time.Time, accountBookID int64) ([]types.CalendarDataItem, error) {
	res, err := s.statisticsRepo.StatisticGroupDate(context, date, accountBookID)

	items := make([]types.CalendarDataItem, len(res))
	if res == nil {
		return nil, err
	}
	for i, row := range res {
		items[i] = types.CalendarDataItem{Date: row.Date, Income: row.Income, Expend: row.Expend}
	}
	return items, err
}

func (s StatisticsService) GetOverviewHeader(context context.Context, date time.Time, accountBookID int64) (types.OverviewHeaderData, error) {
	res, err := s.statisticsRepo.OverviewHeader(context, date, accountBookID)
	res.TotalBalance = res.Income - res.Expend - res.Repay
	return types.OverviewHeaderData{TotalBalance: res.TotalBalance, Repay: res.Repay, Transfer: res.Transfer, Income: res.Income, Expend: res.Expend}, err
}

func (s StatisticsService) GetOverviewRate(context context.Context, statementType string, date time.Time, accountBookID int64) ([]types.StatementListItem, error) {
	startDate, endDate := s.monthRange(date)
	filter := repo.StatementListFilter{
		AccountBookID: accountBookID,
		Type:          statementType,
		StartDate:     &startDate,
		EndDate:       &endDate,
		OrderBy:       "amount desc",
		Limit:         20,
		Offset:        0,
	}
	rows, err := helperservice.AssembleListRows(context, s.statementsRepo, filter)
	if err != nil {
		return nil, err
	}

	items := make([]types.StatementListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s StatisticsService) GetOverviewStatements(context context.Context, statementType string, date time.Time, accountBookID int64) ([]types.StatementListItem, error) {
	startDate, endDate := s.monthRange(date)
	filter := repo.StatementListFilter{
		AccountBookID: accountBookID,
		Type:          statementType,
		StartDate:     &startDate,
		EndDate:       &endDate,
		OrderBy:       "created_at desc",
	}
	rows, err := helperservice.AssembleListRows(context, s.statementsRepo, filter)
	if err != nil {
		return nil, err
	}

	items := make([]types.StatementListItem, 0, len(rows))
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
