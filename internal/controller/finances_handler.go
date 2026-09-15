package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	financeservice "github.com/yigger/jiezhang-backend/internal/service/finance"
)

type FinancesHandler struct {
	service financeservice.FinanceService
}

func NewFinancesHandler(service financeservice.FinanceService) FinancesHandler {
	return FinancesHandler{service: service}
}

// Wallet 钱包总览
// @Summary 钱包总览
// @ID FinancesHandler_Wallet
// @Tags Finances
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.WalletResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /wallet [get]
func (h FinancesHandler) Wallet(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.service.GetWallet(c.Request.Context(), accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load wallet"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// WalletInformation 资产详情，Query: `asset_id`(必填)
// @Summary 资产详情，Query: `asset_id`(必填)
// @ID FinancesHandler_WalletInformation
// @Tags Finances
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.WalletInformation "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /wallet/information [get]
func (h FinancesHandler) WalletInformation(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := ParseInt64Query(c, "asset_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}

	res, err := h.service.GetWalletInformation(c.Request.Context(), accountBook.ID, assetID)
	if err != nil {
		if errors.Is(err, financeservice.ErrRepositoryFinanceAssetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load asset"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// WalletTimeline 资产时间线，Query: `asset_id`(必填)
// @Summary 资产时间线，Query: `asset_id`(必填)
// @ID FinancesHandler_WalletTimeline
// @Tags Finances
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.WalletTimelineResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /wallet/time_line [get]
func (h FinancesHandler) WalletTimeline(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := ParseInt64Query(c, "asset_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}

	res, err := h.service.GetWalletTimeline(c.Request.Context(), accountBook.ID, assetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load timeline"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// WalletStatementList 资产关联账单，Query: `asset_id`, `year`, `month`(均必填)
// @Summary 资产关联账单，Query: `asset_id`, `year`, `month`(均必填)
// @ID FinancesHandler_WalletStatementList
// @Tags Finances
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse{data=[]types.StatementListItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /wallet/statement_list [get]
func (h FinancesHandler) WalletStatementList(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := ParseInt64Query(c, "asset_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}
	year, err := parseIntQuery(c, "year")
	if err != nil || year <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}
	month, err := parseIntQuery(c, "month")
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
		return
	}

	items, err := h.service.GetWalletStatementList(c.Request.Context(), accountBook.ID, assetID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load statement list"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func ParseInt64Query(c *gin.Context, key string) (int64, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, ErrInvalidParam(key)
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidParam(key)
	}
	return id, nil
}

func parseIntQuery(c *gin.Context, key string) (int, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, ErrInvalidParam(key)
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, ErrInvalidParam(key)
	}
	return v, nil
}
