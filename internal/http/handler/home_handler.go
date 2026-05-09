package handler

import "github.com/gin-gonic/gin"

type HomeHandler struct{}

func NewHomeHandler() HomeHandler {
	return HomeHandler{}
}

func (h HomeHandler) Header(c *gin.Context) {
	notImplemented(c, "GET /api/header")
}

func (h HomeHandler) Index(c *gin.Context) {
	notImplemented(c, "GET /api/index")
}
