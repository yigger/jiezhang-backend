package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type StatisticsHandler struct {
	service service.StatisticsService
}

func NewStatisticsHandler(service service.StatisticsService) StatisticsHandler {
	return StatisticsHandler{service: service}
}

func (h StatisticsHandler) CalendarData(c *gin.Context) {
	accountBook, _ := requireAccountBook(c)
	query := c.Query("date")
	dateQuery := fmt.Sprintf("%s-01", query)
	date, err := time.Parse("2006-01-02", dateQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid date params"})
		return
	}

	data, err := h.service.GetCalendarData(c, date, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, data)
}

func (h StatisticsHandler) OverviewHeader(c *gin.Context) {
	notImplemented(c, "GET /api/chart/overview_header")
}

func (h StatisticsHandler) OverviewStatements(c *gin.Context) {
	notImplemented(c, "GET /api/chart/overview_statements")
}

func (h StatisticsHandler) Rate(c *gin.Context) {
	notImplemented(c, "GET /api/chart/rate")
}
