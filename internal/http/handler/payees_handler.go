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

type PayeesHandler struct {
	service service.PayeeService
}

func NewPayeesHandler(service service.PayeeService) PayeesHandler {
	return PayeesHandler{service: service}
}

func (h PayeesHandler) List(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h PayeesHandler) Create(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
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
		if errors.Is(err, service.ErrPayeeInvalidInput) {
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": "收款人名称不能为空"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to create payee"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h PayeesHandler) Update(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
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
		case errors.Is(err, repository.ErrPayeeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "收款人不存在"})
		case errors.Is(err, service.ErrPayeeInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": "收款人名称不能为空"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to update payee"})
		}
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h PayeesHandler) Delete(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
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
		if errors.Is(err, repository.ErrPayeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "收款人不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

type payeeRequest struct {
	Payee struct {
		Name string `json:"name"`
	} `json:"payee"`
	Name string `json:"name"`
}

func parsePayeeName(c *gin.Context) (string, error) {
	var req payeeRequest
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
		return 0, errInvalidParam("id")
	}
	return id, nil
}
