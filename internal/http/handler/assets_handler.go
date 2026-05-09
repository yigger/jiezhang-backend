package handler

import "github.com/gin-gonic/gin"

type AssetsHandler struct{}

func NewAssetsHandler() AssetsHandler {
	return AssetsHandler{}
}

func (h AssetsHandler) List(c *gin.Context) {
	notImplemented(c, "GET /api/assets")
}

func (h AssetsHandler) Show(c *gin.Context) {
	notImplemented(c, "GET /api/assets/:id")
}

func (h AssetsHandler) Delete(c *gin.Context) {
	notImplemented(c, "DELETE /api/assets/:id")
}

func (h AssetsHandler) Icons(c *gin.Context) {
	notImplemented(c, "GET /api/icons/assets_with_url")
}

func (h AssetsHandler) Update(c *gin.Context) {
	notImplemented(c, "PUT /api/assets/:id")
}

func (h AssetsHandler) Create(c *gin.Context) {
	notImplemented(c, "POST /api/assets")
}

func (h AssetsHandler) UpdateSurplus(c *gin.Context) {
	notImplemented(c, "PUT /api/wallet/surplus")
}
