package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/service"
)

type SuperStatementsHandler struct {
	service service.SuperStatementService
}

func NewSuperStatementsHandler(service service.SuperStatementService) SuperStatementsHandler {
	return SuperStatementsHandler{service: service}
}

func (h SuperStatementsHandler) Time(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h SuperStatementsHandler) List(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func buildSuperStatementFilterInput(c *gin.Context, accountBookID int64) (service.SuperStatementFilterInput, error) {
	input := service.SuperStatementFilterInput{
		AccountBookID: accountBookID,
		OrderBy:       strings.TrimSpace(c.Query("order_by")),
	}

	if v := strings.TrimSpace(c.Query("year")); v != "" {
		year, err := strconv.Atoi(v)
		if err != nil || year <= 0 {
			return service.SuperStatementFilterInput{}, errInvalidParam("year")
		}
		input.Year = &year
	}
	if v := strings.TrimSpace(c.Query("month")); v != "" {
		month, err := strconv.Atoi(v)
		if err != nil {
			return service.SuperStatementFilterInput{}, errInvalidParam("month")
		}
		input.Month = &month
	}
	if v := strings.TrimSpace(c.Query("asset")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return service.SuperStatementFilterInput{}, errInvalidParam("asset")
		}
		input.AssetParentID = &id
	}
	if v := strings.TrimSpace(c.Query("asset_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return service.SuperStatementFilterInput{}, errInvalidParam("asset_id")
		}
		input.AssetID = &id
	}
	if v := strings.TrimSpace(c.Query("category_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return service.SuperStatementFilterInput{}, errInvalidParam("category_id")
		}
		input.CategoryID = &id
	}

	return input, nil
}
