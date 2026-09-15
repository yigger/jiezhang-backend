package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	homeservice "github.com/yigger/jiezhang-backend/internal/service/home"
)

type HomeHandler struct {
	service homeservice.HomeService
}

func NewHomeHandler(homeService homeservice.HomeService) HomeHandler {
	return HomeHandler{service: homeService}
}

// Header 首页头部数据（收支趋势、预算等）
// @Summary 首页头部数据（收支趋势、预算等）
// @ID HomeHandler_Header
// @Tags Home
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.HomeHeaderResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /header [get]
func (h HomeHandler) Header(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.service.GetHeader(c.Request.Context(), currentUser.ID, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Index 首页账单列表，Query: `range`（today/yesterday/week/month/year）
// @Summary 首页账单列表，Query: `range`（today/yesterday/week/month/year）
// @ID HomeHandler_Index
// @Tags Home
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param range query string false "range"
// @Success 200 {array} types.StatementListItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /index [get]
func (h HomeHandler) Index(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	items, err := h.service.GetIndex(c.Request.Context(), currentUser.ID, accountBook.ID, c.Query("range"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get index list"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetSettings 用户设置，返回 `{user, version}`，`user.avatar_url` 已按 Rails avatar_path 逻辑处理
// @Summary 用户设置，返回 `{user, version}`，`user.avatar_url` 已按 Rails avatar_path 逻辑处理
// @ID HomeHandler_GetSettings
// @Tags Home
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.HomeSettingsResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /settings [get]
func (h HomeHandler) GetSettings(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.service.GetSettings(c.Request.Context(), currentUser, accountBook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get settings"})
		return
	}
	c.JSON(http.StatusOK, res)
}
