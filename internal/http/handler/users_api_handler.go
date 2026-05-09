package handler

import "github.com/gin-gonic/gin"

type UsersAPIHandler struct{}

func NewUsersAPIHandler() UsersAPIHandler {
	return UsersAPIHandler{}
}

func (h UsersAPIHandler) GetSettings(c *gin.Context) {
	notImplemented(c, "GET /api/settings")
}

func (h UsersAPIHandler) GetUserInfo(c *gin.Context) {
	notImplemented(c, "GET /api/users")
}

func (h UsersAPIHandler) UpdateUser(c *gin.Context) {
	notImplemented(c, "PUT /api/users/update_user")
}

func (h UsersAPIHandler) ScanLogin(c *gin.Context) {
	notImplemented(c, "POST /api/users/scan_login")
}
