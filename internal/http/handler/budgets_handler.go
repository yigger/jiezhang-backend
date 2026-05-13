package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	httpdto "github.com/yigger/jiezhang-backend/internal/http/dto"
	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type BudgetsHandler struct {
	service service.BudgetService
}

func NewBudgetsHandler(service service.BudgetService) BudgetsHandler {
	return BudgetsHandler{service: service}
}

func (h BudgetsHandler) Summary(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	year, month := service.ResolveBudgetYearMonth(c.Query("year"), c.Query("month"))

	res, err := h.service.Summary(c.Request.Context(), accountBook.ID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load budget summary"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h BudgetsHandler) ParentList(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	year, month := service.ResolveBudgetYearMonth(c.Query("year"), c.Query("month"))

	res, err := h.service.ParentList(c.Request.Context(), accountBook.ID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load parent budgets"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h BudgetsHandler) CategoryBudget(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	categoryID, err := parseCategoryID(c.Param("categoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid category id"})
		return
	}
	year, month := service.ResolveBudgetYearMonth(c.Query("year"), c.Query("month"))

	res, err := h.service.CategoryDetail(c.Request.Context(), accountBook.ID, categoryID, year, month)
	if err != nil {
		if errors.Is(err, repository.ErrBudgetCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的分类"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load category budget"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h BudgetsHandler) UpdateAmount(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	var req httpdto.BudgetUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}

	err := h.service.UpdateAmount(c.Request.Context(), accountBook.ID, service.BudgetUpdateInput{
		Type:       req.Type,
		Amount:     req.Amount,
		CategoryID: req.CategoryID,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBudgetInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "金额无效"})
		case errors.Is(err, repository.ErrBudgetCategoryNotFound):
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
