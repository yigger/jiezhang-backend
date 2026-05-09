package handler

import "github.com/gin-gonic/gin"

type SuperStatementsHandler struct{}

func NewSuperStatementsHandler() SuperStatementsHandler {
	return SuperStatementsHandler{}
}

func (h SuperStatementsHandler) Time(c *gin.Context) {
	notImplemented(c, "GET /api/super_statements/time")
}

func (h SuperStatementsHandler) List(c *gin.Context) {
	notImplemented(c, "GET /api/super_statements/list")
}
