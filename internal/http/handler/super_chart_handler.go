package handler

import "github.com/gin-gonic/gin"

type SuperChartHandler struct{}

func NewSuperChartHandler() SuperChartHandler {
	return SuperChartHandler{}
}

func (h SuperChartHandler) Header(c *gin.Context) {
	notImplemented(c, "GET /api/super_chart/header")
}

func (h SuperChartHandler) PieData(c *gin.Context) {
	notImplemented(c, "GET /api/super_chart/get_pie_data")
}

func (h SuperChartHandler) WeekData(c *gin.Context) {
	notImplemented(c, "GET /api/super_chart/week_data")
}

func (h SuperChartHandler) LineChart(c *gin.Context) {
	notImplemented(c, "GET /api/super_chart/line_chart")
}

func (h SuperChartHandler) CategoriesList(c *gin.Context) {
	notImplemented(c, "GET /api/super_chart/categories_list")
}

func (h SuperChartHandler) TableSummary(c *gin.Context) {
	notImplemented(c, "GET /api/super_chart/table_sumary")
}
