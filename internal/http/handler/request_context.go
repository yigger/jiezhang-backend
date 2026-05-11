package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
)

func requireCurrentUser(c *gin.Context) (domain.User, bool) {
	user, ok := c.Get(middleware.CurrentUserContextKey)
	if ok {
		return user.(domain.User), true
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"status": 301,
		"msg":    "session key overdue",
	})
	return domain.User{}, false
}

func requireAccountBook(c *gin.Context) (domain.AccountBook, bool) {
	accountBook, ok := c.Get(middleware.AccountBookContextKey)
	if ok {
		return accountBook.(domain.AccountBook), true
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"status": 301,
		"msg":    "session key overdue",
	})
	return domain.AccountBook{}, false
}
