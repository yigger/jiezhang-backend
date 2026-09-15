package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	budgetservice "github.com/yigger/jiezhang-backend/internal/service/budget"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type BudgetsHandler struct {
	service budgetservice.BudgetService
}

func NewBudgetsHandler(service budgetservice.BudgetService) BudgetsHandler {
	return BudgetsHandler{service: service}
}

// Summary 预算总览，Query: `year`, `month`
// @Summary 预算总览，Query: `year`, `month`
// @ID BudgetsHandler_Summary
// @Tags Budgets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param month query string false "month"
// @Param year query string false "year"
// @Success 200 {object} types.BudgetSummary "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /budgets [get]
func (h BudgetsHandler) Summary(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	year, month := budgetservice.ResolveBudgetYearMonth(c.Query("year"), c.Query("month"))

	res, err := h.service.Summary(c.Request.Context(), accountBook.ID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load budget summary"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// ParentList 父级分类预算列表，Query: `year`, `month`
// @Summary 父级分类预算列表，Query: `year`, `month`
// @ID BudgetsHandler_ParentList
// @Tags Budgets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param month query string false "month"
// @Param year query string false "year"
// @Success 200 {array} types.BudgetParentItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /budgets/parent [get]
func (h BudgetsHandler) ParentList(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	year, month := budgetservice.ResolveBudgetYearMonth(c.Query("year"), c.Query("month"))

	res, err := h.service.ParentList(c.Request.Context(), accountBook.ID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load parent budgets"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// CategoryBudget 分类预算详情，Query: `year`, `month`
// @Summary 分类预算详情，Query: `year`, `month`
// @ID BudgetsHandler_CategoryBudget
// @Tags Budgets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param categoryId path integer true "categoryId"
// @Param month query string false "month"
// @Param year query string false "year"
// @Success 200 {object} types.BudgetCategoryDetail "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /budgets/{categoryId} [get]
func (h BudgetsHandler) CategoryBudget(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	categoryID, err := ParseCategoryID(c.Param("categoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid category id"})
		return
	}
	year, month := budgetservice.ResolveBudgetYearMonth(c.Query("year"), c.Query("month"))

	res, err := h.service.CategoryDetail(c.Request.Context(), accountBook.ID, categoryID, year, month)
	if err != nil {
		if errors.Is(err, budgetservice.ErrRepositoryBudgetCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的分类"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load category budget"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// UpdateAmount 更新预算，Body: `{type, amount, category_id}`，type='user' 为总预算，'category' 为分类预算，amount 支持数字或字符串，category_id 可选
// @Summary 更新预算，Body: `{type, amount, category_id}`，type='user' 为总预算，'category' 为分类预算，amount 支持数字或字符串，category_id 可选
// @ID BudgetsHandler_UpdateAmount
// @Tags Budgets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.BudgetUpdateRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /budgets/0 [put]
func (h BudgetsHandler) UpdateAmount(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	var req types.BudgetUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}

	err := h.service.UpdateAmount(c.Request.Context(), accountBook.ID, budgetservice.BudgetUpdateInput{
		Type:       req.Type,
		Amount:     string(req.Amount),
		CategoryID: req.CategoryID,
	})
	if err != nil {
		switch {
		case errors.Is(err, budgetservice.ErrBudgetInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "金额无效"})
		case errors.Is(err, budgetservice.ErrRepositoryBudgetCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的分类"})
		default:
			if err.Error() == "总预算必须大于分类预算的总和" || err.Error() == "一级分类预算不能少于二级分类的总和" {
				c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update budget"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}
