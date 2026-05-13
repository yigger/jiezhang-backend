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

type MessagesHandler struct {
	service       service.MessageService
	publicBaseURL string
}

func NewMessagesHandler(service service.MessageService, publicBaseURL string) MessagesHandler {
	return MessagesHandler{service: service, publicBaseURL: strings.TrimSpace(publicBaseURL)}
}

func (h MessagesHandler) List(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}

	items, err := h.service.List(c.Request.Context(), currentUser.ID, h.publicBaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to list messages"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h MessagesHandler) Show(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseMessageID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid message id"})
		return
	}

	item, err := h.service.Show(c.Request.Context(), currentUser.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrMessageNotFound) {
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load message"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func parseMessageID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidParam("id")
	}
	return id, nil
}
