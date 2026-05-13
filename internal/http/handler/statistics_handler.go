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
	date := h.formatDate(c)
	data, err := h.service.GetCalendarData(c, date, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": data})
}

func (h StatisticsHandler) OverviewHeader(c *gin.Context) {
	accountBook, _ := requireAccountBook(c)
	date := h.formatDate(c)
	data, err := h.service.GetOverviewHeader(c, date, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h StatisticsHandler) OverviewStatements(c *gin.Context) {
	accountBook, _ := requireAccountBook(c)
	date := h.formatDate(c)
	data, err := h.service.GetOverviewStatements(c, c.Query("type"), date, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h StatisticsHandler) Rate(c *gin.Context) {
	accountBook, _ := requireAccountBook(c)
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
