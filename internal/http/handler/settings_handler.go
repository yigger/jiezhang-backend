package handler

import "github.com/gin-gonic/gin"

type SettingsHandler struct{}

func NewSettingsHandler() SettingsHandler {
	return SettingsHandler{}
}

func (h SettingsHandler) Feedback(c *gin.Context) {
	notImplemented(c, "POST /api/settings/feedback")
}
