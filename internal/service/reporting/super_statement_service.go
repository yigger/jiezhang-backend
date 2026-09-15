package reporting

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	statementservice "github.com/yigger/jiezhang-backend/internal/service/statement"
	"github.com/yigger/jiezhang-backend/internal/types"
)

var ErrSuperStatementInvalidInput = errors.New("super statement invalid input")

type SuperStatementService struct {
	repo      repo.SuperStatementRepository
	rowMapper statementservice.RowMapper
}

func NewSuperStatementService(repo repo.SuperStatementRepository, rowMapper statementservice.RowMapper) SuperStatementService {
	return SuperStatementService{repo: repo, rowMapper: rowMapper}
}

type SuperStatementFilterInput struct {
	AccountBookID int64
	Year          *int
	Month         *int
	AssetParentID *int64
	AssetID       *int64
	CategoryID    *int64
	OrderBy       string
}

func (s SuperStatementService) Time(ctx context.Context, input SuperStatementFilterInput) (types.SuperTimeResponse, error) {
	filter := toSuperStatementFilter(input)
	summaries, err := s.repo.ListMonthSummaries(ctx, filter)
	if err != nil {
		return types.SuperTimeResponse{}, err
	}
	overview, err := s.repo.GetOverview(ctx, filter)
	if err != nil {
		return types.SuperTimeResponse{}, err
	}

	months := make([]types.SuperMonthItem, 0, len(summaries))
	for _, row := range summaries {
		months = append(months, types.SuperMonthItem{
			ExpendAmount: row.ExpendAmount,
			IncomeAmount: row.IncomeAmount,
			Surplus:      row.Surplus,
			Year:         row.Year,
			Month:        row.Month,
			Hidden:       1,
		})
	}
	return types.SuperTimeResponse{
		Statements: months,
		Header: types.SuperHeader{
			Expend: superMoney(overview.Expend),
			Income: superMoney(overview.Income),
			Left:   superMoney(overview.Left),
		},
	}, nil
}

func (s SuperStatementService) List(ctx context.Context, input SuperStatementFilterInput) ([]types.StatementListItem, error) {
	filter := toSuperStatementFilter(input)
	rows, err := helperservice.AssembleSuperListRows(ctx, s.repo, filter)
	if err != nil {
		return nil, err
	}

	items := make([]types.StatementListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}
	return items, nil
}

type SuperChartService struct {
	repo repo.SuperChartRepository
}

func NewSuperChartService(repo repo.SuperChartRepository) SuperChartService {
	return SuperChartService{repo: repo}
}

type SuperChartYearMonthInput struct {
	AccountBookID int64
	Year          int
	Month         int
	StatementType string
}

func (s SuperChartService) Header(ctx context.Context, input SuperChartYearMonthInput) (types.SuperChartHeaderData, error) {
	current, err := s.repo.GetMonthSummary(ctx, input.AccountBookID, input.Year, input.Month, true)
	if err != nil {
		return types.SuperChartHeaderData{}, err
	}
	lastMonth := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.Local).AddDate(0, -1, 0)
	prev, err := s.repo.GetMonthSummary(ctx, input.AccountBookID, lastMonth.Year(), int(lastMonth.Month()), true)
	if err != nil {
		return types.SuperChartHeaderData{}, err
	}

	curSurplus := current.Income - current.Expend
	prevSurplus := prev.Income - prev.Expend
	expendPercent, expendRise := superGetPercent(current.Expend, prev.Expend)
	incomePercent, incomeRise := superGetPercent(current.Income, prev.Income)
	surplusPercent, surplusRise := superGetPercent(curSurplus, prevSurplus)

	return types.SuperChartHeaderData{
		ExpendCount:    current.Expend,
		IncomeCount:    current.Income,
		Surplus:        curSurplus,
		ExpendPercent:  expendPercent,
		ExpendRise:     expendRise,
		IncomePercent:  incomePercent,
		IncomeRise:     incomeRise,
		SurplusPercent: surplusPercent,
		SurplusRise:    surplusRise,
	}, nil
}

func (s SuperChartService) TableSummary(ctx context.Context, input SuperChartYearMonthInput) ([]types.SuperTableSummaryItem, error) {
	days, err := s.repo.ListDaySummaries(ctx, input.AccountBookID, input.Year, input.Month)
	if err != nil {
		return nil, err
	}

	result := make([]types.SuperTableSummaryItem, 0, len(days))
	totalExpend := 0.0
	totalIncome := 0.0
	for _, d := range days {
		totalExpend += d.Expend
		totalIncome += d.Income
		totalSurplus := totalIncome - totalExpend
		result = append(result, types.SuperTableSummaryItem{
			Date:         fmt.Sprintf("%d/%d/%d", input.Year, input.Month, d.Day),
			Expend:       superMoney(d.Expend),
			Income:       superMoney(d.Income),
			TotalIncome:  superMoney(totalIncome),
			TotalExpend:  superMoney(totalExpend),
			TotalSurplus: superMoney(totalSurplus),
		})
	}
	return result, nil
}

func (s SuperChartService) PieData(ctx context.Context, input SuperChartYearMonthInput) ([]types.SuperPieItem, error) {
	rows, err := s.repo.ListPieParents(ctx, input.AccountBookID, input.Year, input.Month, input.StatementType)
	if err != nil {
		return nil, err
	}
	items := make([]types.SuperPieItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, types.SuperPieItem{
			Name: row.ParentName,
			Data: int64(row.Data),
		})
	}
	return items, nil
}

func (s SuperChartService) CategoriesTop(ctx context.Context, input SuperChartYearMonthInput) ([]types.SuperCategoryTopItem, error) {
	rows, err := s.repo.ListCategoryTop(ctx, input.AccountBookID, input.Year, input.Month)
	if err != nil {
		return nil, err
	}
	total := 0.0
	for _, row := range rows {
		total += row.Data
	}
	items := make([]types.SuperCategoryTopItem, 0, len(rows))
	for _, row := range rows {
		percent := 0.0
		if total > 0 {
			percent = (row.Data / total) * 100
		}
		items = append(items, types.SuperCategoryTopItem{
			Name:         row.Name,
			Data:         row.Data,
			FormatAmount: superMoney(row.Data),
			Percent:      fmt.Sprintf("%.2f", percent),
			CategoryID:   row.CategoryID,
		})
	}
	return items, nil
}

func (s SuperChartService) LineChart(ctx context.Context, input SuperChartYearMonthInput) (types.SuperLineChartData, error) {
	yearMonths, err := s.repo.ListYearMonths(ctx, input.AccountBookID)
	if err != nil {
		return types.SuperLineChartData{}, err
	}
	limitMonth := input.Month
	if limitMonth <= 0 {
		limitMonth = int(time.Now().Month())
	}

	filtered := make([]repo.SuperChartYearMonth, 0, len(yearMonths))
	for _, ym := range yearMonths {
		if ym.Year < input.Year || (ym.Year == input.Year && ym.Month <= limitMonth) {
			filtered = append(filtered, ym)
		}
	}
	if len(filtered) > 6 {
		filtered = filtered[len(filtered)-6:]
	}

	res := types.SuperLineChartData{
		Months:  make([]int, 0, len(filtered)),
		Expends: make([]float64, 0, len(filtered)),
		Incomes: make([]float64, 0, len(filtered)),
		Surplus: make([]float64, 0, len(filtered)),
	}
	for _, ym := range filtered {
		sum, sumErr := s.repo.GetMonthSummary(ctx, input.AccountBookID, ym.Year, ym.Month, false)
		if sumErr != nil {
			return types.SuperLineChartData{}, sumErr
		}
		res.Months = append(res.Months, ym.Month)
		res.Expends = append(res.Expends, sum.Expend)
		res.Incomes = append(res.Incomes, sum.Income)
		res.Surplus = append(res.Surplus, sum.Income-sum.Expend)
	}
	return res, nil
}

func (s SuperChartService) WeekData(ctx context.Context, input SuperChartYearMonthInput) (types.SuperWeekData, error) {
	loc := time.Local
	cur := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, loc)
	endDay := cur.AddDate(0, 1, -1).Day()

	weeks := make([]string, 0)
	data := make([]float64, 0)
	dayCursor := cur
	index := 1
	for {
		weekStart, weekEnd := weekRange(dayCursor)
		amount, err := s.repo.SumExpendBetween(ctx, input.AccountBookID, weekStart, weekEnd)
		if err != nil {
			return types.SuperWeekData{}, err
		}
		weeks = append(weeks, fmt.Sprintf("第%d周", index))
		data = append(data, amount)

		dayCursor = dayCursor.AddDate(0, 0, index)
		if dayCursor.Day() > endDay || dayCursor.Month() != time.Month(input.Month) {
			break
		}
		index++
	}
	return types.SuperWeekData{Weeks: weeks, Data: data}, nil
}

func toSuperStatementFilter(input SuperStatementFilterInput) repo.SuperStatementFilter {
	return repo.SuperStatementFilter{
		AccountBookID: input.AccountBookID,
		Year:          input.Year,
		Month:         input.Month,
		AssetParentID: input.AssetParentID,
		AssetID:       input.AssetID,
		CategoryID:    input.CategoryID,
		OrderBy:       strings.TrimSpace(input.OrderBy),
	}
}

func weekRange(t time.Time) (time.Time, time.Time) {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, -(wd - 1))
	end := start.AddDate(0, 0, 6)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), end.Location())
	return start, end
}

func superMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func superGetPercent(curAmount float64, prevAmount float64) (float64, string) {
	allAmount := curAmount + prevAmount
	if allAmount == 0 {
		return 0, "expend"
	}
	if curAmount > prevAmount {
		return ((curAmount - prevAmount) * 100) / allAmount, "income"
	}
	return ((prevAmount - curAmount) * 100) / allAmount, "expend"
}

func ParseSuperYearMonth(yearRaw string, monthRaw string) (int, int, error) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if strings.TrimSpace(yearRaw) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(yearRaw))
		if err != nil || v <= 0 {
			return 0, 0, ErrSuperStatementInvalidInput
		}
		year = v
	}
	if strings.TrimSpace(monthRaw) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(monthRaw))
		if err != nil || v < 1 || v > 12 {
			return 0, 0, ErrSuperStatementInvalidInput
		}
		month = v
	}
	return year, month, nil
}
