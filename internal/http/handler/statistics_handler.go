package handler

import "github.com/gin-gonic/gin"

type StatisticsHandler struct{}

func NewStatisticsHandler() StatisticsHandler {
	return StatisticsHandler{}
}

func (h StatisticsHandler) CalendarData(c *gin.Context) {
	notImplemented(c, "GET /api/chart/calendar_data")
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
