package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

type HomeService struct {
	repo      repository.HomeRepository
	statement StatementService
}

func NewHomeService(repo repository.HomeRepository, statement StatementService) HomeService {
	return HomeService{repo: repo, statement: statement}
}

type HeaderResponse struct {
	Trends        HeaderTrends   `json:"trends"`
	MonthExpend   string         `json:"month_expend"`
	TodayExpend   string         `json:"today_expend"`
	MonthBudget   string         `json:"month_budget"`
	UsePencentage int            `json:"use_pencentage"`
	Message       *HeaderMessage `json:"message"`
}

type HeaderTrends struct {
	Day   HeaderTrendItem `json:"day"`
	Week  HeaderTrendItem `json:"week"`
	Month HeaderTrendItem `json:"month"`
}

type HeaderTrendItem struct {
	Ratio  float64 `json:"ratio"`
	Trend  string  `json:"trend"`
	Amount string  `json:"amount"`
}

type HeaderMessage struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
}

type SettingsResponse struct {
	User    SettingsUser `json:"user"`
	Version string       `json:"version"`
}

type SettingsUser struct {
	UID         int64               `json:"uid"`
	Name        string              `json:"name"`
	Avatar      string              `json:"avatar"`
	Themes      []domain.Theme      `json:"themes"`
	ThemeID     int64               `json:"theme_id"`
	Theme       domain.Theme        `json:"theme"`
	Persist     int64               `json:"persist"`
	ShowDiamond bool                `json:"show_diamond"`
	Remind      bool                `json:"remind"`
	CreatedAt   time.Time           `json:"created_at"`
	AccountBook SettingsAccountBook `json:"account_book"`
}

type SettingsAccountBook struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (s HomeService) GetHeader(ctx context.Context, userID int64, accountBookID int64) (HeaderResponse, error) {
	now := time.Now()
	monthStart := startOfMonth(now)
	monthEnd := endOfMonth(now)
	todayStart := startOfDay(now)
	todayEnd := endOfDay(now)
	yesterdayStart := startOfDay(now.AddDate(0, 0, -1))
	yesterdayEnd := endOfDay(now.AddDate(0, 0, -1))

	weekStart := startOfWeek(now)
	weekEnd := endOfWeek(now)
	lastWeekStart := startOfWeek(now.AddDate(0, 0, -7))
	lastWeekEnd := endOfWeek(now.AddDate(0, 0, -7))

	lastMonthStart := startOfMonth(now.AddDate(0, -1, 0))
	lastMonthEnd := endOfMonth(now.AddDate(0, -1, 0))

	monthExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, monthStart, monthEnd)
	if err != nil {
		return HeaderResponse{}, err
	}
	todayExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, todayStart, todayEnd)
	if err != nil {
		return HeaderResponse{}, err
	}
	yesterdayExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, yesterdayStart, yesterdayEnd)
	if err != nil {
		return HeaderResponse{}, err
	}
	thisWeekExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, weekStart, weekEnd)
	if err != nil {
		return HeaderResponse{}, err
	}
	lastWeekExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, lastWeekStart, lastWeekEnd)
	if err != nil {
		return HeaderResponse{}, err
	}
	lastMonthExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, lastMonthStart, lastMonthEnd)
	if err != nil {
		return HeaderResponse{}, err
	}

	budget, err := s.repo.GetAccountBookBudget(ctx, accountBookID)
	if err != nil {
		return HeaderResponse{}, err
	}

	usePercentage := 0
	if budget != 0 {
		usePercentage = int((monthExpend / budget) * 100)
	}

	messageRow, err := s.repo.FindLatestUnreadMessage(ctx, userID)
	if err != nil {
		return HeaderResponse{}, err
	}

	var message *HeaderMessage
	if messageRow != nil {
		message = &HeaderMessage{ID: messageRow.ID, Title: messageRow.Title, SubTitle: messageRow.SubTitle}
	}

	dayRatioValue, dayTrend := calculateRatio(todayExpend, yesterdayExpend)
	weekRatioValue, weekTrend := calculateRatio(thisWeekExpend, lastWeekExpend)
	monthRatioValue, monthTrend := calculateRatio(monthExpend, lastMonthExpend)

	return HeaderResponse{
		Trends: HeaderTrends{
			Day:   HeaderTrendItem{Ratio: dayRatioValue, Trend: dayTrend, Amount: moneyFormat(todayExpend)},
			Week:  HeaderTrendItem{Ratio: weekRatioValue, Trend: weekTrend, Amount: moneyFormat(thisWeekExpend)},
			Month: HeaderTrendItem{Ratio: monthRatioValue, Trend: monthTrend, Amount: moneyFormat(monthExpend)},
		},
		MonthExpend:   moneyFormat(monthExpend),
		TodayExpend:   moneyFormat(todayExpend),
		MonthBudget:   moneyFormat(budget),
		UsePencentage: usePercentage,
		Message:       message,
	}, nil
}

func (s HomeService) GetIndex(ctx context.Context, userID int64, accountBookID int64, rangeKey string) ([]statementdto.ListItem, error) {
	start, end := resolveRange(strings.TrimSpace(rangeKey), time.Now())

	items := make([]statementdto.ListItem, 0)
	offset := 0
	limit := 200
	for {
		list, err := s.statement.GetStatements(ctx, statementdto.ListInput{
			UserID:        userID,
			AccountBookID: accountBookID,
			StartDate:     &start,
			EndDate:       &end,
			OrderBy:       "created_at",
			Limit:         limit,
			Offset:        offset,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, list...)
		if len(list) < limit {
			break
		}
		offset += limit
	}

	return items, nil
}

func (s HomeService) GetSettings(ctx context.Context, currentUser domain.User, accountBook domain.AccountBook) (SettingsResponse, error) {
	theme := findThemeByID(currentUser.ThemeID)
	persist, err := s.repo.CountUserPersistDays(ctx, currentUser.ID, accountBook.ID)
	if err != nil {
		return SettingsResponse{}, err
	}

	return SettingsResponse{
		User: SettingsUser{
			UID:         currentUser.UID,
			Name:        currentUser.Nickname,
			Avatar:      currentUser.AvatarUrl,
			Themes:      domain.DefaultThemes,
			ThemeID:     currentUser.ThemeID,
			Theme:       theme,
			Persist:     persist,
			ShowDiamond: currentUser.ID == 2,
			Remind:      false,
			CreatedAt:   currentUser.CreatedAt,
			AccountBook: SettingsAccountBook{ID: accountBook.ID, Name: accountBook.Name},
		},
		Version: "1.0.0",
	}, nil
}

func findThemeByID(themeID int64) domain.Theme {
	for _, item := range domain.DefaultThemes {
		if item.ID == themeID {
			return item
		}
	}
	return domain.Theme{}
}

func resolveRange(r string, now time.Time) (time.Time, time.Time) {
	switch r {
	case "yesterday":
		d := now.AddDate(0, 0, -1)
		return startOfDay(d), endOfDay(d)
	case "week":
		return startOfWeek(now), endOfWeek(now)
	case "month":
		return startOfMonth(now), endOfMonth(now)
	case "year":
		return startOfYear(now), endOfYear(now)
	default:
		return startOfDay(now), endOfDay(now)
	}
}

func calculateRatio(current, previous float64) (float64, string) {
	if previous == 0 {
		return 0, "down"
	}
	ratio := ((current - previous) / previous) * 100
	ratio = math.Round(math.Abs(ratio)*100) / 100
	if current-previous > 0 {
		return ratio, "up"
	}
	return ratio, "down"
}

func moneyFormat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

func startOfWeek(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	d := t.AddDate(0, 0, -(wd - 1))
	return startOfDay(d)
}

func endOfWeek(t time.Time) time.Time {
	return endOfDay(startOfWeek(t).AddDate(0, 0, 6))
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func endOfMonth(t time.Time) time.Time {
	return endOfDay(startOfMonth(t).AddDate(0, 1, -1))
}

func startOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func endOfYear(t time.Time) time.Time {
	return endOfDay(time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location()))
}
