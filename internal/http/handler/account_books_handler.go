package handler

import "github.com/gin-gonic/gin"

type AccountBooksHandler struct{}

func NewAccountBooksHandler() AccountBooksHandler {
	return AccountBooksHandler{}
}

func (h AccountBooksHandler) List(c *gin.Context) {
	notImplemented(c, "GET /api/account_books")
}

func (h AccountBooksHandler) Show(c *gin.Context) {
	notImplemented(c, "GET /api/account_books/:id")
}

func (h AccountBooksHandler) Types(c *gin.Context) {
	notImplemented(c, "GET /api/account_books/types")
}

func (h AccountBooksHandler) PresetCategories(c *gin.Context) {
	notImplemented(c, "GET /api/account_books/preset_categories")
}

func (h AccountBooksHandler) Switch(c *gin.Context) {
	notImplemented(c, "PUT /api/account_books/:id/switch")
}

func (h AccountBooksHandler) Create(c *gin.Context) {
	notImplemented(c, "POST /api/account_books")
}

func (h AccountBooksHandler) Update(c *gin.Context) {
	notImplemented(c, "PUT /api/account_books/:id")
}

func (h AccountBooksHandler) Delete(c *gin.Context) {
	notImplemented(c, "DELETE /api/account_books/:id")
}
