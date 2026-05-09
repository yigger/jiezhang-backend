package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type HelloHandler struct{}

func NewHelloHandler() HelloHandler {
	return HelloHandler{}
}

func (h HelloHandler) Hello(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		name = "world"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "hello, " + name,
	})
}
