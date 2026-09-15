package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	messageservice "github.com/yigger/jiezhang-backend/internal/service/message"
)

type MessagesHandler struct {
	service       messageservice.MessageService
	publicBaseURL string
}

func NewMessagesHandler(service messageservice.MessageService, publicBaseURL string) MessagesHandler {
	return MessagesHandler{service: service, publicBaseURL: string(strings.TrimSpace(string(publicBaseURL)))}
}

// List 消息列表
// @Summary 消息列表
// @ID MessagesHandler_List
// @Tags Messages
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {array} types.MessageListItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /message [get]
func (h MessagesHandler) List(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}

	items, err := h.service.List(c.Request.Context(), currentUser.ID, string(h.publicBaseURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to list messages"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Show 消息详情
// @Summary 消息详情
// @ID MessagesHandler_Show
// @Tags Messages
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.MessageDetailItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /message/{id} [get]
func (h MessagesHandler) Show(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
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
		if errors.Is(err, messageservice.ErrRepositoryMessageNotFound) {
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
		return 0, ErrInvalidParam("id")
	}
	return id, nil
}
