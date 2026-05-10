package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type AccountBookHandler struct {
	service service.AccountBookService
}

func NewAccountBookHandler(service service.AccountBookService) AccountBookHandler {
	return AccountBookHandler{
		service: service,
	}
}

func (h AccountBookHandler) List(c *gin.Context) {
	notImplemented(c, "GET /api/account_books")
}

func (h AccountBookHandler) Show(c *gin.Context) {
	notImplemented(c, "GET /api/account_books/:id")
}

func (h AccountBookHandler) Types(c *gin.Context) {
	notImplemented(c, "GET /api/account_books/types")
}

func (h AccountBookHandler) PresetCategories(c *gin.Context) {
	notImplemented(c, "GET /api/account_books/preset_categories")
}

func (h AccountBookHandler) Switch(c *gin.Context) {
	notImplemented(c, "PUT /api/account_books/:id/switch")
}

func (h AccountBookHandler) Create(c *gin.Context) {
	notImplemented(c, "POST /api/account_books")
}

func (h AccountBookHandler) Update(c *gin.Context) {
	notImplemented(c, "PUT /api/account_books/:id")
}

func (h AccountBookHandler) Delete(c *gin.Context) {
	notImplemented(c, "DELETE /api/account_books/:id")
}
