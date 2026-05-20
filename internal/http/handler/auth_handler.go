package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service"
	authservice "github.com/yigger/jiezhang-backend/internal/service/auth"
)

type AuthHandler struct {
	checkOpenIDService authservice.CheckOpenIDService
	uploadService      service.UploadService
}

func NewAuthHandler(checkOpenIDService authservice.CheckOpenIDService, uploadService service.UploadService) AuthHandler {
	return AuthHandler{
		checkOpenIDService: checkOpenIDService,
		uploadService:      uploadService,
	}
}

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

func (h AuthHandler) Upload(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "上传文件不能为空"})
		return
	}

	statementID := int64(0)
	if v := strings.TrimSpace(c.PostForm("statement_id")); v != "" {
		if id, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil && id > 0 {
			statementID = id
		}
	}

	res, err := h.uploadService.Upload(c.Request.Context(), service.UploadInput{
		Type:          c.PostForm("type"),
		UserID:        currentUser.ID,
		AccountBookID: accountBook.ID,
		StatementID:   statementID,
		File:          file,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUploadUnknownType):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		case errors.Is(err, service.ErrUploadInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		case errors.Is(err, repository.ErrStatementNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "statement not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "上传失败"})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}
