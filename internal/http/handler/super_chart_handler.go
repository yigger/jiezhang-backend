package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/service"
)

type SuperChartHandler struct {
	service service.SuperChartService
}

func NewSuperChartHandler(service service.SuperChartService) SuperChartHandler {
	return SuperChartHandler{service: service}
}

func (h SuperChartHandler) Header(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h SuperChartHandler) PieData(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h SuperChartHandler) WeekData(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h SuperChartHandler) LineChart(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h SuperChartHandler) CategoriesList(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h SuperChartHandler) TableSummary(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func buildSuperChartInput(c *gin.Context, accountBookID int64) (service.SuperChartYearMonthInput, error) {
	year, month, err := service.ParseSuperYearMonth(c.Query("year"), c.Query("month"))
	if err != nil {
		return service.SuperChartYearMonthInput{}, err
	}
	return service.SuperChartYearMonthInput{
		AccountBookID: accountBookID,
		Year:          year,
		Month:         month,
	}, nil
}
