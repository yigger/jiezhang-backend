package handler

import "github.com/gin-gonic/gin"

type MessagesHandler struct{}

func NewMessagesHandler() MessagesHandler {
	return MessagesHandler{}
}

func (h MessagesHandler) List(c *gin.Context) {
	notImplemented(c, "GET /api/message")
}

func (h MessagesHandler) Show(c *gin.Context) {
	notImplemented(c, "GET /api/message/:id")
}
