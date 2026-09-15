package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	reportingservice "github.com/yigger/jiezhang-backend/internal/service/reporting"
)

type SuperStatementsHandler struct {
	service reportingservice.SuperStatementService
}

func NewSuperStatementsHandler(service reportingservice.SuperStatementService) SuperStatementsHandler {
	return SuperStatementsHandler{service: service}
}

// Time 月度汇总时间线，Query: `year`, `month`, `asset`, `asset_id`, `category_id`, `order_by`
// @Summary 月度汇总时间线，Query: `year`, `month`, `asset`, `asset_id`, `category_id`, `order_by`
// @ID SuperStatementsHandler_Time
// @Tags SuperStatements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param asset query string false "asset"
// @Param asset_id query string false "asset_id"
// @Param category_id query string false "category_id"
// @Param month query string false "month"
// @Param order_by query string false "order_by"
// @Param year query string false "year"
// @Success 200 {object} types.APIResponse{data=types.SuperTimeResponse} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /super_statements/time [get]
func (h SuperStatementsHandler) Time(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	input, err := buildSuperStatementFilterInput(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid parameters"})
		return
	}
	res, err := h.service.Time(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load super statements"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

// List 流水列表，Query 同上
// @Summary 流水列表，Query 同上
// @ID SuperStatementsHandler_List
// @Tags SuperStatements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param asset query string false "asset"
// @Param asset_id query string false "asset_id"
// @Param category_id query string false "category_id"
// @Param month query string false "month"
// @Param order_by query string false "order_by"
// @Param year query string false "year"
// @Success 200 {object} types.APIResponse{data=[]types.StatementListItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /super_statements/list [get]
func (h SuperStatementsHandler) List(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	input, err := buildSuperStatementFilterInput(c, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid parameters"})
		return
	}
	items, err := h.service.List(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load statements"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func buildSuperStatementFilterInput(c *gin.Context, accountBookID int64) (reportingservice.SuperStatementFilterInput, error) {
	input := reportingservice.SuperStatementFilterInput{
		AccountBookID: accountBookID,
		OrderBy:       strings.TrimSpace(c.Query("order_by")),
	}

	if v := strings.TrimSpace(c.Query("year")); v != "" {
		year, err := strconv.Atoi(v)
		if err != nil || year <= 0 {
			return reportingservice.SuperStatementFilterInput{}, ErrInvalidParam("year")
		}
		input.Year = &year
	}
	if v := strings.TrimSpace(c.Query("month")); v != "" {
		month, err := strconv.Atoi(v)
		if err != nil {
			return reportingservice.SuperStatementFilterInput{}, ErrInvalidParam("month")
		}
		input.Month = &month
	}
	if v := strings.TrimSpace(c.Query("asset")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return reportingservice.SuperStatementFilterInput{}, ErrInvalidParam("asset")
		}
		input.AssetParentID = &id
	}
	if v := strings.TrimSpace(c.Query("asset_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return reportingservice.SuperStatementFilterInput{}, ErrInvalidParam("asset_id")
		}
		input.AssetID = &id
	}
	if v := strings.TrimSpace(c.Query("category_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return reportingservice.SuperStatementFilterInput{}, ErrInvalidParam("category_id")
		}
		input.CategoryID = &id
	}

	return input, nil
}
