package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	reportingservice "github.com/yigger/jiezhang-backend/internal/service/reporting"
)

type SuperChartHandler struct {
	service reportingservice.SuperChartService
}

func NewSuperChartHandler(service reportingservice.SuperChartService) SuperChartHandler {
	return SuperChartHandler{service: service}
}

// Header 图表头部数据，Query: `year`, `month`
// @Summary 图表头部数据，Query: `year`, `month`
// @ID SuperChartHandler_Header
// @Tags SuperChart
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param month query string false "month"
// @Param year query string false "year"
// @Success 200 {object} types.APIResponse{data=types.SuperChartHeaderData} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /super_chart/header [get]
func (h SuperChartHandler) Header(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	input, err := buildSuperChartInput(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid year/month"})
		return
	}

	data, err := h.service.Header(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load super chart header"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// PieData 饼图数据，Query: `year`, `month`, `statement_type`
// @Summary 饼图数据，Query: `year`, `month`, `statement_type`
// @ID SuperChartHandler_PieData
// @Tags SuperChart
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param month query string false "month"
// @Param statement_type query string false "statement_type"
// @Param year query string false "year"
// @Success 200 {object} types.APIResponse{data=[]types.SuperPieItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /super_chart/get_pie_data [get]
func (h SuperChartHandler) PieData(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	input, err := buildSuperChartInput(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid year/month"})
		return
	}
	input.StatementType = strings.TrimSpace(c.Query("statement_type"))

	items, err := h.service.PieData(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load pie data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// WeekData 周数据，Query: `year`, `month`
// @Summary 周数据，Query: `year`, `month`
// @ID SuperChartHandler_WeekData
// @Tags SuperChart
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param month query string false "month"
// @Param year query string false "year"
// @Success 200 {object} types.SuperWeekData "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /super_chart/week_data [get]
func (h SuperChartHandler) WeekData(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	input, err := buildSuperChartInput(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid year/month"})
		return
	}
	data, err := h.service.WeekData(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load week data"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// LineChart 折线图数据，Query: `year`, `month`
// @Summary 折线图数据，Query: `year`, `month`
// @ID SuperChartHandler_LineChart
// @Tags SuperChart
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param month query string false "month"
// @Param year query string false "year"
// @Success 200 {object} types.SuperLineChartData "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /super_chart/line_chart [get]
func (h SuperChartHandler) LineChart(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	input, err := buildSuperChartInput(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid year/month"})
		return
	}
	data, err := h.service.LineChart(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load line chart"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// CategoriesList 分类 Top 排行，Query: `year`, `month`
// @Summary 分类 Top 排行，Query: `year`, `month`
// @ID SuperChartHandler_CategoriesList
// @Tags SuperChart
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param month query string false "month"
// @Param year query string false "year"
// @Success 200 {object} types.APIResponse{data=[]types.SuperCategoryTopItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /super_chart/categories_list [get]
func (h SuperChartHandler) CategoriesList(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	input, err := buildSuperChartInput(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid year/month"})
		return
	}
	items, err := h.service.CategoriesTop(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load categories top"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// TableSummary 日汇总表，Query: `year`, `month`
// @Summary 日汇总表，Query: `year`, `month`
// @ID SuperChartHandler_TableSummary
// @Tags SuperChart
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param month query string false "month"
// @Param year query string false "year"
// @Success 200 {object} types.APIResponse{data=[]types.SuperTableSummaryItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /super_chart/table_sumary [get]
func (h SuperChartHandler) TableSummary(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	input, err := buildSuperChartInput(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid year/month"})
		return
	}
	items, err := h.service.TableSummary(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load table summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": items})
}

func buildSuperChartInput(c *gin.Context, accountBookID int64) (reportingservice.SuperChartYearMonthInput, error) {
	year, month, err := reportingservice.ParseSuperYearMonth(c.Query("year"), c.Query("month"))
	if err != nil {
		return reportingservice.SuperChartYearMonthInput{}, err
	}
	return reportingservice.SuperChartYearMonthInput{
		AccountBookID: accountBookID,
		Year:          year,
		Month:         month,
	}, nil
}
