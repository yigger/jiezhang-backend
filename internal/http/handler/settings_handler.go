package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	httpdto "github.com/yigger/jiezhang-backend/internal/http/dto"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type SettingsHandler struct {
	service service.SettingService
}

func NewSettingsHandler(service service.SettingService) SettingsHandler {
	return SettingsHandler{service: service}
}

func (h SettingsHandler) Feedback(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}

	var req httpdto.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "内容不能为空"})
		return
	}

	err := h.service.SubmitFeedback(c.Request.Context(), service.SettingFeedbackInput{
		UserID:  currentUser.ID,
		Content: req.Content,
		Type:    req.Type,
	})
	if err != nil {
		if errors.Is(err, service.ErrSettingInvalidInput) {
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "内容不能为空"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to submit feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}
