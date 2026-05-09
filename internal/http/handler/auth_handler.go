package handler

import "github.com/gin-gonic/gin"

type AuthHandler struct{}

func NewAuthHandler() AuthHandler {
	return AuthHandler{}
}

func (h AuthHandler) CheckOpenID(c *gin.Context) {
	notImplemented(c, "POST /api/check_openid")
}

func (h AuthHandler) Upload(c *gin.Context) {
	notImplemented(c, "POST /api/upload")
}
