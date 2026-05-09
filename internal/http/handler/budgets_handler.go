package handler

import "github.com/gin-gonic/gin"

type BudgetsHandler struct{}

func NewBudgetsHandler() BudgetsHandler {
	return BudgetsHandler{}
}

func (h BudgetsHandler) Summary(c *gin.Context) {
	notImplemented(c, "GET /api/budgets")
}

func (h BudgetsHandler) ParentList(c *gin.Context) {
	notImplemented(c, "GET /api/budgets/parent")
}

func (h BudgetsHandler) CategoryBudget(c *gin.Context) {
	notImplemented(c, "GET /api/budgets/:categoryId")
}

func (h BudgetsHandler) UpdateAmount(c *gin.Context) {
	notImplemented(c, "PUT /api/budgets/0")
}
