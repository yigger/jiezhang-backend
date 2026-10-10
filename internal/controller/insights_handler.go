package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/service/insights"
	"net/http"
	"strconv"
	"time"
)

type InsightsHandler struct {
	service *insights.Service
	storage *insights.StorageService
}

func NewInsightsHandler(service *insights.Service, storage *insights.StorageService) InsightsHandler {
	return InsightsHandler{service: service, storage: storage}
}

// Report 年度收支、商家、记录成员和固定开销候选汇总。
// @Summary 年度收支与多维汇总（整数分）
// @ID InsightsHandler_Report
// @Tags Insights
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID，沿用账簿访问权限"
// @Param year query integer false "年份，默认今年"
// @Success 200 {object} types.InsightReport
// @Failure 400 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /insights/report [get]
func (h InsightsHandler) Report(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	year := time.Now().Year()
	if value := c.Query("year"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "年份格式无效"})
			return
		}
		year = parsed
	}
	report, err := h.service.Report(c.Request.Context(), book.ID, year)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, insights.ErrInvalidYear) {
			status = http.StatusBadRequest
		}
		message := "分析数据暂时不可用，请稍后重试"
		if status == http.StatusBadRequest {
			message = err.Error()
		}
		c.JSON(status, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusOK, report)
}
