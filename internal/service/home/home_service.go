package home

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	statementservice "github.com/yigger/jiezhang-backend/internal/service/statement"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type StatementReader interface {
	GetStatements(context.Context, statementservice.ListInput) ([]types.StatementListItem, error)
}

type HomeService struct {
	repo          repo.HomeRepository
	statement     StatementReader
	publicBaseURL string
}

func NewHomeService(repo repo.HomeRepository, statement StatementReader, publicBaseURL string) HomeService {
	return HomeService{repo: repo, statement: statement, publicBaseURL: publicBaseURL}
}

func (s HomeService) GetHeader(ctx context.Context, userID int64, accountBookID int64) (types.HomeHeaderResponse, error) {
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
		return types.HomeHeaderResponse{}, err
	}
	todayExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, todayStart, todayEnd)
	if err != nil {
		return types.HomeHeaderResponse{}, err
	}
	yesterdayExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, yesterdayStart, yesterdayEnd)
	if err != nil {
		return types.HomeHeaderResponse{}, err
	}
	thisWeekExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, weekStart, weekEnd)
	if err != nil {
		return types.HomeHeaderResponse{}, err
	}
	lastWeekExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, lastWeekStart, lastWeekEnd)
	if err != nil {
		return types.HomeHeaderResponse{}, err
	}
	lastMonthExpend, err := s.repo.SumExpendInRange(ctx, accountBookID, lastMonthStart, lastMonthEnd)
	if err != nil {
		return types.HomeHeaderResponse{}, err
	}

	budget, err := s.repo.GetAccountBookBudget(ctx, accountBookID)
	if err != nil {
		return types.HomeHeaderResponse{}, err
	}

	usePercentage := 0
	if budget != 0 {
		usePercentage = int((monthExpend / budget) * 100)
	}

	messageRow, err := s.repo.FindLatestUnreadMessage(ctx, userID)
	if err != nil {
		return types.HomeHeaderResponse{}, err
	}

	var message *types.HomeHeaderMessage
	if messageRow != nil {
		message = &types.HomeHeaderMessage{ID: messageRow.ID, Title: messageRow.Title, SubTitle: messageRow.SubTitle}
	}

	dayRatioValue, dayTrend := calculateRatio(todayExpend, yesterdayExpend)
	weekRatioValue, weekTrend := calculateRatio(thisWeekExpend, lastWeekExpend)
	monthRatioValue, monthTrend := calculateRatio(monthExpend, lastMonthExpend)

	return types.HomeHeaderResponse{
		Trends: types.HomeHeaderTrends{
			Day:   types.HomeHeaderTrendItem{Ratio: dayRatioValue, Trend: dayTrend, Amount: moneyFormat(todayExpend)},
			Week:  types.HomeHeaderTrendItem{Ratio: weekRatioValue, Trend: weekTrend, Amount: moneyFormat(thisWeekExpend)},
			Month: types.HomeHeaderTrendItem{Ratio: monthRatioValue, Trend: monthTrend, Amount: moneyFormat(monthExpend)},
		},
		MonthExpend:   moneyFormat(monthExpend),
		TodayExpend:   moneyFormat(todayExpend),
		MonthBudget:   moneyFormat(budget),
		UsePencentage: usePercentage,
		Message:       message,
	}, nil
}

func (s HomeService) GetIndex(ctx context.Context, userID int64, accountBookID int64, rangeKey string) ([]types.StatementListItem, error) {
	start, end := resolveRange(strings.TrimSpace(rangeKey), time.Now())

	items := make([]types.StatementListItem, 0)
	offset := 0
	limit := 200
	for {
		list, err := s.statement.GetStatements(ctx, statementservice.ListInput{
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

func (s HomeService) GetSettings(ctx context.Context, currentUser types.UserContext, accountBook tablemodel.AccountBook) (types.HomeSettingsResponse, error) {
	theme := findThemeByID(currentUser.ThemeID)
	persist, err := s.repo.CountUserPersistDays(ctx, currentUser.ID, accountBook.ID)
	if err != nil {
		return types.HomeSettingsResponse{}, err
	}

	return types.HomeSettingsResponse{
		User: types.HomeSettingsUser{
			UID:         currentUser.UID,
			Name:        currentUser.Nickname,
			AvatarUrl:   s.buildAvatarURL(currentUser.AvatarUrl, currentUser.ID),
			Themes:      defaultThemes,
			ThemeID:     currentUser.ThemeID,
			Theme:       theme,
			Persist:     persist,
			ShowDiamond: currentUser.ID == 2,
			Remind:      false,
			CreatedAt:   currentUser.CreatedAt,
			AccountBook: types.HomeSettingsAccountBook{ID: accountBook.ID, Name: accountBook.Name},
		},
		Version: "1.0.0",
	}, nil
}

func (s HomeService) buildAvatarURL(avatarURL string, userID int64) string {
	// 微信头像等已经是完整 URL 的，直接返回
	if strings.HasPrefix(avatarURL, "http") {
		return avatarURL
	}
	baseURL := string(s.publicBaseURL)
	// avatarURL 为空时兜底为默认头像
	if avatarURL == "" {
		return baseURL + "/public/common-avatar.png"
	}
	// 本地头像拼接完整路径: baseURL/private/{userID}/user/xxx.jpg
	return fmt.Sprintf("%s/private/%d/user/%s", baseURL, userID, avatarURL)
}

func findThemeByID(themeID int64) types.Theme {
	for _, item := range defaultThemes {
		if item.ID == themeID {
			return item
		}
	}
	return types.Theme{}
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
