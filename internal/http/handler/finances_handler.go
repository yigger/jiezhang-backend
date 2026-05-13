package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type FinancesHandler struct {
	service service.FinanceService
}

func NewFinancesHandler(service service.FinanceService) FinancesHandler {
	return FinancesHandler{service: service}
}

func (h FinancesHandler) Wallet(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h FinancesHandler) WalletInformation(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := parseInt64Query(c, "asset_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}

	res, err := h.service.GetWalletInformation(c.Request.Context(), accountBook.ID, assetID)
	if err != nil {
		if errors.Is(err, repository.ErrFinanceAssetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load asset"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h FinancesHandler) WalletTimeline(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := parseInt64Query(c, "asset_id")
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

func (h FinancesHandler) WalletStatementList(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := parseInt64Query(c, "asset_id")
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

func parseInt64Query(c *gin.Context, key string) (int64, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, errInvalidParam(key)
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidParam(key)
	}
	return id, nil
}

func parseIntQuery(c *gin.Context, key string) (int, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, errInvalidParam(key)
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, errInvalidParam(key)
	}
	return v, nil
}
