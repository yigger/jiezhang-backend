package handler

import "github.com/gin-gonic/gin"

type CategoriesHandler struct{}

func NewCategoriesHandler() CategoriesHandler {
	return CategoriesHandler{}
}

func (h CategoriesHandler) List(c *gin.Context) {
	notImplemented(c, "GET /api/categories/category_list")
}

func (h CategoriesHandler) Show(c *gin.Context) {
	notImplemented(c, "GET /api/categories/:id")
}

func (h CategoriesHandler) Delete(c *gin.Context) {
	notImplemented(c, "DELETE /api/categories/:id")
}

func (h CategoriesHandler) Icons(c *gin.Context) {
	notImplemented(c, "GET /api/icons/categories_with_url")
}

func (h CategoriesHandler) Update(c *gin.Context) {
	notImplemented(c, "PUT /api/categories/:id")
}

func (h CategoriesHandler) Create(c *gin.Context) {
	notImplemented(c, "POST /api/categories")
}
