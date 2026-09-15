package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	statisticsservice "github.com/yigger/jiezhang-backend/internal/service/statistics"
)

type StatisticsHandler struct {
	service statisticsservice.StatisticsService
}

func NewStatisticsHandler(service statisticsservice.StatisticsService) StatisticsHandler {
	return StatisticsHandler{service: service}
}

// CalendarData 日历热力图，Query: `date`(YYYY-MM)
// @Summary 日历热力图，Query: `date`(YYYY-MM)
// @ID StatisticsHandler_CalendarData
// @Tags Statistics
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse{data=[]types.CalendarDataItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /chart/calendar_data [get]
func (h StatisticsHandler) CalendarData(c *gin.Context) {
	accountBook, _ := RequireAccountBook(c)
	date := h.formatDate(c)
	data, err := h.service.GetCalendarData(c, date, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": data})
}

// OverviewHeader 概览头部数据，Query: `date`(YYYY-MM)
// @Summary 概览头部数据，Query: `date`(YYYY-MM)
// @ID StatisticsHandler_OverviewHeader
// @Tags Statistics
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.OverviewHeaderData "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /chart/overview_header [get]
func (h StatisticsHandler) OverviewHeader(c *gin.Context) {
	accountBook, _ := RequireAccountBook(c)
	date := h.formatDate(c)
	data, err := h.service.GetOverviewHeader(c, date, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// OverviewStatements 概览账单列表，Query: `date`, `type`
// @Summary 概览账单列表，Query: `date`, `type`
// @ID StatisticsHandler_OverviewStatements
// @Tags Statistics
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {array} types.StatementListItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /chart/overview_statements [get]
func (h StatisticsHandler) OverviewStatements(c *gin.Context) {
	accountBook, _ := RequireAccountBook(c)
	date := h.formatDate(c)
	data, err := h.service.GetOverviewStatements(c, c.Query("type"), date, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// Rate 收支占比，Query: `date`, `type`
// @Summary 收支占比，Query: `date`, `type`
// @ID StatisticsHandler_Rate
// @Tags Statistics
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {array} types.StatementListItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /chart/rate [get]
func (h StatisticsHandler) Rate(c *gin.Context) {
	accountBook, _ := RequireAccountBook(c)
	date := h.formatDate(c)
	data, err := h.service.GetOverviewRate(c, c.Query("type"), date, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h StatisticsHandler) formatDate(c *gin.Context) time.Time {
	query := c.Query("date")
	dateQuery := fmt.Sprintf("%s-01", query)
	date, err := time.Parse("2006-01-02", dateQuery)
	if err != nil {
		date = time.Now()
	}
	return date
}
