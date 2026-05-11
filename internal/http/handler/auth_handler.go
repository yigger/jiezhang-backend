package handler

import (
	"github.com/gin-gonic/gin"

	authservice "github.com/yigger/jiezhang-backend/internal/service/auth"
)

type AuthHandler struct {
	checkOpenIDService authservice.CheckOpenIDService
}

func NewAuthHandler(checkOpenIDService authservice.CheckOpenIDService) AuthHandler {
	return AuthHandler{checkOpenIDService: checkOpenIDService}
}

func (h AuthHandler) CheckOpenID(c *gin.Context) {
	// log.Printf("CheckOpenID called with headers: %v", c.Request.Header)
	// code := strings.TrimSpace(c.GetHeader("X-WX-Code"))
	// if code == "" {
	// 	c.JSON(200, gin.H{"status": 401, "msg": "登录失败"})
	// 	return
	// }

	// session, err := h.checkOpenIDService.Execute(c.Request.Context(), code)
	// if err != nil {
	// 	if errors.Is(err, authservice.ErrLoginFailed) {
	// 		c.JSON(200, gin.H{"status": 401, "msg": "登录失败"})
	// 		return
	// 	}
	// 	c.JSON(200, gin.H{"status": 500, "msg": "服务异常"})
	// 	return
	// }

	c.JSON(200, gin.H{"status": 200, "session": "mock_session"})
}

func (h AuthHandler) Upload(c *gin.Context) {
	notImplemented(c, "POST /api/v1/upload")
}
