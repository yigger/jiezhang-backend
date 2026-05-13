package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	httpdto "github.com/yigger/jiezhang-backend/internal/http/dto"
	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) UserHandler {
	return UserHandler{service: service}
}

func (h UserHandler) GetUserInfo(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), currentUser.ID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200, "data": profile})
}

func (h UserHandler) UpdateUser(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}

	var req httpdto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "无法获取相关信息"})
		return
	}

	if isUserUpdatePayloadEmpty(req.User) {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "无法获取相关信息"})
		return
	}

	err := h.service.UpdateProfile(c.Request.Context(), currentUser.ID, service.UserProfileUpdateInput{
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
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h UserHandler) ScanLogin(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}

	var req httpdto.UserScanLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "二维码已失效，刷新浏览器界面重新获取二维码..."})
		return
	}

	err := h.service.ScanLogin(c.Request.Context(), currentUser.ID, req.QRCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserInvalidInput), errors.Is(err, service.ErrUserQRCodeExpired):
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
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

type createUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Email string `json:"email" binding:"required,email"`
}

func (h UserHandler) Create(c *gin.Context) {
	var req createUserRequest
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

func isUserUpdatePayloadEmpty(p httpdto.UserUpdatePayload) bool {
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
