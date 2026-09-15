package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	userservice "github.com/yigger/jiezhang-backend/internal/service/user"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type UserHandler struct {
	service userservice.UserService
}

func NewUserHandler(service userservice.UserService) UserHandler {
	return UserHandler{service: service}
}

// GetUserInfo 获取当前用户信息
// @Summary 获取当前用户信息
// @ID UserHandler_GetUserInfo
// @Tags User
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse{data=types.UserProfile} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /users [get]
func (h UserHandler) GetUserInfo(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), currentUser.ID)
	if err != nil {
		if errors.Is(err, userservice.ErrRepositoryUserNotFound) {
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200, "data": profile})
}

// UpdateUser 更新用户，Body: `{user: {theme_id, country, city, gender, language, province, bg_avatar_id, hidden_asset_money, avatar_url, nickname, bg_avatar}}`，所有字段可选
// @Summary 更新用户，Body: `{user: {theme_id, country, city, gender, language, province, bg_avatar_id, hidden_asset_money, avatar_url, nickname, bg_avatar}}`，所有字段可选
// @ID UserHandler_UpdateUser
// @Tags User
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.UserUpdateRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /users/update_user [put]
func (h UserHandler) UpdateUser(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}

	var req types.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "无法获取相关信息"})
		return
	}

	if isUserUpdatePayloadEmpty(req.User) {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "无法获取相关信息"})
		return
	}

	err := h.service.UpdateProfile(c.Request.Context(), currentUser.ID, userservice.UserProfileUpdateInput{
		ThemeID:          req.User.ThemeID,
		Country:          req.User.Country,
		City:             req.User.City,
		Gender:           req.User.Gender,
		Language:         req.User.Language,
		Province:         req.User.Province,
		BGAvatarID:       req.User.BGAvatarID,
		HiddenAssetMoney: req.User.HiddenAssetMoney,
		AvatarURL:        req.User.AvatarURL,
		Nickname:         req.User.Nickname,
		BGAvatar:         req.User.BGAvatar,
	})
	if err != nil {
		if errors.Is(err, userservice.ErrRepositoryUserNotFound) {
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// ScanLogin PC 扫码登录，Body: `{qr_code}`
// @Summary PC 扫码登录，Body: `{qr_code}`
// @ID UserHandler_ScanLogin
// @Tags User
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.UserScanLoginRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /users/scan_login [post]
func (h UserHandler) ScanLogin(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}

	var req types.UserScanLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "二维码已失效，刷新浏览器界面重新获取二维码..."})
		return
	}

	err := h.service.ScanLogin(c.Request.Context(), currentUser.ID, req.QRCode)
	if err != nil {
		switch {
		case errors.Is(err, userservice.ErrUserInvalidInput), errors.Is(err, userservice.ErrUserQRCodeExpired):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "二维码已失效，刷新浏览器界面重新获取二维码..."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to scan login"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h UserHandler) List(c *gin.Context) {
	users, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h UserHandler) Show(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, userservice.ErrRepositoryUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h UserHandler) Create(c *gin.Context) {
	var req types.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Create(c.Request.Context(), req.Name, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func isUserUpdatePayloadEmpty(p types.UserUpdatePayload) bool {
	return p.ThemeID == nil &&
		p.Country == nil &&
		p.City == nil &&
		p.Gender == nil &&
		p.Language == nil &&
		p.Province == nil &&
		p.BGAvatarID == nil &&
		p.HiddenAssetMoney == nil &&
		p.AvatarURL == nil &&
		p.Nickname == nil &&
		(p.BGAvatar == nil || strings.TrimSpace(*p.BGAvatar) == "")
}
