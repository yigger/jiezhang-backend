package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
)

func fetchCurrentUser(c *gin.Context) (domain.User, bool) {
	user, ok := middleware.CurrentUser(c)
	if ok {
		return user, true
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"status": 301,
		"msg":    "session key overdue",
	})
	return domain.User{}, false
}
