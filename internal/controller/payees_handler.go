package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	payeeservice "github.com/yigger/jiezhang-backend/internal/service/payee"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type PayeesHandler struct {
	service payeeservice.PayeeService
}

func NewPayeesHandler(service payeeservice.PayeeService) PayeesHandler {
	return PayeesHandler{service: service}
}

// List 收款方列表
// @Summary 收款方列表
// @ID PayeesHandler_List
// @Tags Payees
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {array} types.PayeeListItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /payees [get]
func (h PayeesHandler) List(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	items, err := h.service.List(c.Request.Context(), accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to list payees"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Create 创建收款方，Body: `{name}` 或 `{payee: {name}}`
// @Summary 创建收款方，Body: `{name}` 或 `{payee: {name}}`
// @ID PayeesHandler_Create
// @Tags Payees
// @Produce json
// @Accept json
// @Param body body types.PayeeWriteRequest true "名称，可使用 name 或 payee.name"
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} model.Payee "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /payees [post]
func (h PayeesHandler) Create(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	name, err := parsePayeeName(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid payee name"})
		return
	}

	item, err := h.service.Create(c.Request.Context(), currentUser.ID, accountBook.ID, name)
	if err != nil {
		if errors.Is(err, payeeservice.ErrPayeeInvalidInput) {
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": "收款人名称不能为空"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to create payee"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Update 更新收款方
// @Summary 更新收款方
// @ID PayeesHandler_Update
// @Tags Payees
// @Produce json
// @Accept json
// @Param body body types.PayeeWriteRequest true "名称，可使用 name 或 payee.name"
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} model.Payee "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /payees/{id} [put]
func (h PayeesHandler) Update(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	payeeID, err := parsePayeeID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid payee id"})
		return
	}
	name, err := parsePayeeName(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid payee name"})
		return
	}

	item, err := h.service.Update(c.Request.Context(), payeeID, currentUser.ID, name)
	if err != nil {
		switch {
		case errors.Is(err, payeeservice.ErrRepositoryPayeeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "收款人不存在"})
		case errors.Is(err, payeeservice.ErrPayeeInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": "收款人名称不能为空"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to update payee"})
		}
		return
	}
	c.JSON(http.StatusOK, item)
}

// Delete 删除收款方
// @Summary 删除收款方
// @ID PayeesHandler_Delete
// @Tags Payees
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /payees/{id} [delete]
func (h PayeesHandler) Delete(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	payeeID, err := parsePayeeID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid payee id"})
		return
	}

	err = h.service.Delete(c.Request.Context(), payeeID, currentUser.ID)
	if err != nil {
		if errors.Is(err, payeeservice.ErrRepositoryPayeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "收款人不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func parsePayeeName(c *gin.Context) (string, error) {
	var req types.PayeeWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return "", err
	}
	if strings.TrimSpace(req.Payee.Name) != "" {
		return req.Payee.Name, nil
	}
	return req.Name, nil
}

func parsePayeeID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidParam("id")
	}
	return id, nil
}
