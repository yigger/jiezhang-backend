package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	settingservice "github.com/yigger/jiezhang-backend/internal/service/setting"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type SettingsHandler struct {
	service settingservice.SettingService
}

func NewSettingsHandler(service settingservice.SettingService) SettingsHandler {
	return SettingsHandler{service: service}
}

// Feedback 提交反馈，Body: `{content, type}`
// @Summary 提交反馈，Body: `{content, type}`
// @ID SettingsHandler_Feedback
// @Tags Settings
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.FeedbackRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /settings/feedback [post]
func (h SettingsHandler) Feedback(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}

	var req types.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "内容不能为空"})
		return
	}

	err := h.service.SubmitFeedback(c.Request.Context(), settingservice.SettingFeedbackInput{
		UserID:  currentUser.ID,
		Content: req.Content,
		Type:    req.Type,
	})
	if err != nil {
		if errors.Is(err, settingservice.ErrSettingInvalidInput) {
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "内容不能为空"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to submit feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}
