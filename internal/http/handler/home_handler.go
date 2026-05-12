package handler

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/service"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

type HomeHandler struct {
	db        *gorm.DB
	statement service.StatementService
}

func NewHomeHandler(db *gorm.DB, statementService service.StatementService) HomeHandler {
	return HomeHandler{db: db, statement: statementService}
}

func (h HomeHandler) Header(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

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

	monthExpend, err := h.sumExpendInRange(c, accountBook.ID, monthStart, monthEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}
	todayExpend, err := h.sumExpendInRange(c, accountBook.ID, todayStart, todayEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}
	yesterdayExpend, err := h.sumExpendInRange(c, accountBook.ID, yesterdayStart, yesterdayEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}
	thisWeekExpend, err := h.sumExpendInRange(c, accountBook.ID, weekStart, weekEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}
	lastWeekExpend, err := h.sumExpendInRange(c, accountBook.ID, lastWeekStart, lastWeekEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}
	thisMonthExpend, err := h.sumExpendInRange(c, accountBook.ID, monthStart, monthEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}
	lastMonthExpend, err := h.sumExpendInRange(c, accountBook.ID, lastMonthStart, lastMonthEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}

	budget, err := h.getAccountBookBudget(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}

	var usePercentage int
	if budget == 0 {
		usePercentage = 0
	} else {
		usePercentage = int((monthExpend / budget) * 100)
	}

	message, err := h.getLatestUnreadMessage(c, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}

	dayRatioValue, dayTrend := calculateRatio(todayExpend, yesterdayExpend)
	weekRatioValue, weekTrend := calculateRatio(thisWeekExpend, lastWeekExpend)
	monthRatioValue, monthTrend := calculateRatio(thisMonthExpend, lastMonthExpend)

	c.JSON(http.StatusOK, gin.H{
		"trends": gin.H{
			"day": gin.H{
				"ratio":  dayRatioValue,
				"trend":  dayTrend,
				"amount": moneyFormat(todayExpend),
			},
			"week": gin.H{
				"ratio":  weekRatioValue,
				"trend":  weekTrend,
				"amount": moneyFormat(thisWeekExpend),
			},
			"month": gin.H{
				"ratio":  monthRatioValue,
				"trend":  monthTrend,
				"amount": moneyFormat(thisMonthExpend),
			},
		},
		"month_expend":   moneyFormat(monthExpend),
		"today_expend":   moneyFormat(todayExpend),
		"month_budget":   moneyFormat(budget),
		"use_pencentage": usePercentage,
		"message":        message,
	})
}

func (h HomeHandler) Index(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	now := time.Now()
	start, end := resolveRange(strings.TrimSpace(c.Query("range")), now)

	items := make([]interface{}, 0)
	offset := 0
	limit := 200
	for {
		list, err := h.statement.GetStatements(c.Request.Context(), statementdto.ListInput{
			UserID:        currentUser.ID,
			AccountBookID: accountBook.ID,
			StartDate:     &start,
			EndDate:       &end,
			OrderBy:       "created_at",
			Limit:         limit,
			Offset:        offset,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get index list"})
			return
		}
		for _, item := range list {
			items = append(items, item)
		}
		if len(list) < limit {
			break
		}
		offset += limit
	}

	c.JSON(http.StatusOK, items)
}

func (h HomeHandler) GetSettings(c *gin.Context) {
	currentUser, _ := requireCurrentUser(c)
	accountBook, _ := requireAccountBook(c)

	var theme domain.Theme
	for _, item := range domain.DefaultThemes {
		if item.ID == currentUser.ThemeID {
			theme = item
			break
		}
	}

	// user.statements.count("distinct year, month, day")
	var persist int64
	err := h.db.WithContext(c.Request.Context()).
		Table("statements").
		Where("user_id = ? AND account_book_id = ?", currentUser.ID, accountBook.ID).
		Select("COUNT(DISTINCT DATE(created_at))").
		Scan(&persist).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"uid":          currentUser.UID,
			"name":         currentUser.Nickname,
			"avatar":       currentUser.AvatarUrl,
			"themes":       domain.DefaultThemes,
			"theme_id":     currentUser.ThemeID,
			"theme":        theme,
			"persist":      persist,
			"show_diamond": currentUser.ID == 2,
			"remind":       false, // TODO: need_remind 后续加上
			"created_at":   currentUser.CreatedAt,
			"account_book": gin.H{
				"id":   accountBook.ID,
				"name": accountBook.Name,
			},
		},
		"version": "1.0.0", // TODO: 版本号后续从配置或者数据库获取
	})

}

func (h HomeHandler) sumExpendInRange(c *gin.Context, accountBookID int64, start time.Time, end time.Time) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	err := h.db.WithContext(c.Request.Context()).
		Table("statements").
		Select("COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ?", accountBookID).
		Where("type IN ?", []string{"expend", "repayment"}).
		Where("created_at BETWEEN ? AND ?", start, end).
		Scan(&row).Error
	return row.Amount, err
}

func (h HomeHandler) getAccountBookBudget(c *gin.Context, accountBookID int64) (float64, error) {
	var row struct {
		Budget float64 `gorm:"column:budget"`
	}
	err := h.db.WithContext(c.Request.Context()).
		Table("account_books").
		Select("COALESCE(budget, 0) AS budget").
		Where("id = ?", accountBookID).
		Take(&row).Error
	return row.Budget, err
}

func (h HomeHandler) getLatestUnreadMessage(c *gin.Context, userID int64) (interface{}, error) {
	var row struct {
		ID       int64  `gorm:"column:id"`
		Title    string `gorm:"column:title"`
		SubTitle string `gorm:"column:sub_title"`
	}
	err := h.db.WithContext(c.Request.Context()).
		Table("messages").
		Select("id, title, sub_title").
		Where("target_id = ? AND already_read = 0", userID).
		Order("created_at DESC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return gin.H{
		"id":        row.ID,
		"title":     row.Title,
		"sub_title": row.SubTitle,
	}, nil
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
