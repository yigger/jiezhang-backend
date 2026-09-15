package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	authservice "github.com/yigger/jiezhang-backend/internal/service/auth"
	uploadservice "github.com/yigger/jiezhang-backend/internal/service/upload"
)

type AuthHandler struct {
	checkOpenIDService authservice.CheckOpenIDService
	uploadService      uploadservice.UploadService
}

func NewAuthHandler(checkOpenIDService authservice.CheckOpenIDService, uploadService uploadservice.UploadService) AuthHandler {
	return AuthHandler{
		checkOpenIDService: checkOpenIDService,
		uploadService:      uploadService,
	}
}

// CheckOpenID 微信登录，Header: `X-WX-Code`，返回 `{'status': 200, 'session': '...'}`
// @Summary 微信登录，Header: `X-WX-Code`，返回 `{'status': 200, 'session': '...'}`
// @ID AuthHandler_CheckOpenID
// @Tags Auth
// @Produce json
// @Param X-WX-Code header string true "微信登录 code"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /check_openid [post]
func (h AuthHandler) CheckOpenID(c *gin.Context) {
	code := strings.TrimSpace(c.GetHeader("X-WX-Code"))
	if code == "" {
		c.JSON(200, gin.H{"status": 401, "msg": "登录失败"})
		return
	}

	session, err := h.checkOpenIDService.Execute(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, authservice.ErrLoginFailed) {
			c.JSON(200, gin.H{"status": 401, "msg": "登录失败"})
			return
		}
		c.JSON(200, gin.H{"status": 500, "msg": "服务异常"})
		return
	}

	c.JSON(200, gin.H{"status": 200, "session": session})
}

// Upload 上传文件（FormData），字段：`file`(文件)、`type`、`statement_id`(选填)
// @Summary 上传文件（FormData），字段：`file`(文件)、`type`、`statement_id`(选填)
// @ID AuthHandler_Upload
// @Tags Auth
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept multipart/form-data
// @Param file formData file true "上传文件"
// @Param statement_id formData string false "statement_id"
// @Param type formData string false "type"
// @Success 200 {object} types.UploadResult "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /upload [post]
func (h AuthHandler) Upload(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "上传文件不能为空"})
		return
	}

	src, openErr := file.Open()
	if openErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "上传失败"})
		return
	}
	defer src.Close()
	statementID := int64(0)
	if v := strings.TrimSpace(c.PostForm("statement_id")); v != "" {
		if id, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil && id > 0 {
			statementID = id
		}
	}

	res, err := h.uploadService.Upload(c.Request.Context(), uploadservice.UploadInput{
		Type:          c.PostForm("type"),
		UserID:        currentUser.ID,
		AccountBookID: accountBook.ID,
		StatementID:   statementID,
		File:          src, Filename: file.Filename,
	})
	if err != nil {
		switch {
		case errors.Is(err, uploadservice.ErrUploadUnknownType):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		case errors.Is(err, uploadservice.ErrUploadInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		case errors.Is(err, authservice.ErrRepositoryStatementNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "statement not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "上传失败"})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}
