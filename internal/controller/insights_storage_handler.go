package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/service/insights"
	"github.com/yigger/jiezhang-backend/internal/types"
	"net/http"
)

func insightError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "分析数据暂时不可用，请稍后重试"
	if errors.Is(err, insights.ErrInvalidInput) || errors.Is(err, insights.ErrMissingRecord) {
		status = http.StatusBadRequest
		message = err.Error()
	} else if errors.Is(err, insights.ErrForbidden) {
		status = http.StatusForbidden
		message = err.Error()
	}
	c.JSON(status, gin.H{"error": message})
}

// Workspace 汇总项目、固定开销、付款分摊及资产历史。
// @Summary 全历史分析工作区
// @ID InsightsHandler_Workspace
// @Tags Insights
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID"
// @Success 200 {object} types.InsightWorkspace
// @Failure 500 {object} types.APIResponse
// @Router /insights/workspace [get]
func (h InsightsHandler) Workspace(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	user, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	result, err := h.storage.Workspace(c.Request.Context(), book, user.ID)
	if err != nil {
		insightError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// SaveProject 创建或编辑项目。
// @Summary 保存项目
// @ID InsightsHandler_SaveProject
// @Tags Insights
// @Accept json
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID"
// @Param payload body types.InsightProjectInput true "项目资料"
// @Success 200 {object} types.InsightSavedResponse
// @Failure 400 {object} types.APIResponse
// @Failure 403 {object} types.APIResponse
// @Router /insights/projects [post]
func (h InsightsHandler) SaveProject(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	user, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var in types.InsightProjectInput
	if c.ShouldBindJSON(&in) != nil {
		insightError(c, insights.ErrInvalidInput)
		return
	}
	result, err := h.storage.SaveProject(c.Request.Context(), book.ID, user.ID, in)
	if err != nil {
		insightError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.InsightSavedResponse{ID: result.ID})
}

// SaveFixedCost 创建或编辑用户确认的固定开销。
// @Summary 保存固定开销规则
// @ID InsightsHandler_SaveFixedCost
// @Tags Insights
// @Accept json
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID"
// @Param payload body types.InsightFixedCostInput true "规则资料"
// @Success 200 {object} types.InsightSavedResponse
// @Failure 400 {object} types.APIResponse
// @Failure 403 {object} types.APIResponse
// @Router /insights/fixed_costs [post]
func (h InsightsHandler) SaveFixedCost(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	user, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var in types.InsightFixedCostInput
	if c.ShouldBindJSON(&in) != nil {
		insightError(c, insights.ErrInvalidInput)
		return
	}
	result, err := h.storage.SaveFixedCost(c.Request.Context(), book.ID, user.ID, in)
	if err != nil {
		insightError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.InsightSavedResponse{ID: result.ID})
}

// SaveAnnotation 修改账单的分析标注，不影响财务余额。
// @Summary 设置项目、实际付款人和分摊
// @ID InsightsHandler_SaveAnnotation
// @Tags Insights
// @Accept json
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID"
// @Param payload body types.InsightAnnotationInput true "完整标注，null 表示清除"
// @Success 200 {object} types.InsightSavedResponse
// @Failure 400 {object} types.APIResponse
// @Failure 403 {object} types.APIResponse
// @Router /insights/annotation [put]
func (h InsightsHandler) SaveAnnotation(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	user, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var in types.InsightAnnotationInput
	if c.ShouldBindJSON(&in) != nil {
		insightError(c, insights.ErrInvalidInput)
		return
	}
	if err := h.storage.SaveAnnotation(c.Request.Context(), book.ID, user.ID, in); err != nil {
		insightError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.InsightSavedResponse{ID: in.StatementID})
}

// CapturePortfolio 捕获当前资产余额，不回填历史。
// @Summary 保存当前资产组合快照
// @ID InsightsHandler_CapturePortfolio
// @Tags Insights
// @Accept json
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID"
// @Param payload body types.InsightSnapshotInput true "快照备注"
// @Success 200 {object} types.InsightSavedResponse
// @Failure 400 {object} types.APIResponse
// @Failure 403 {object} types.APIResponse
// @Router /insights/portfolio [post]
func (h InsightsHandler) CapturePortfolio(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	user, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var in types.InsightSnapshotInput
	if c.ShouldBindJSON(&in) != nil {
		insightError(c, insights.ErrInvalidInput)
		return
	}
	result, err := h.storage.CapturePortfolio(c.Request.Context(), book.ID, user.ID, in.Note)
	if err != nil {
		insightError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.InsightSavedResponse{ID: result.ID})
}

// SaveMerchantAliases 设置商家统计名称，空名称恢复原名。
// @Summary 合并商家统计
// @ID InsightsHandler_SaveMerchantAliases
// @Tags Insights
// @Accept json
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID"
// @Param payload body types.InsightMerchantAliasInput true "商家 ID 和统一名称"
// @Success 200 {object} types.InsightSavedResponse
// @Failure 400 {object} types.APIResponse
// @Failure 403 {object} types.APIResponse
// @Router /insights/merchants [put]
func (h InsightsHandler) SaveMerchantAliases(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	user, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var in types.InsightMerchantAliasInput
	if c.ShouldBindJSON(&in) != nil {
		insightError(c, insights.ErrInvalidInput)
		return
	}
	if err := h.storage.SaveMerchantAliases(c.Request.Context(), book.ID, user.ID, in); err != nil {
		insightError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.InsightSavedResponse{})
}
