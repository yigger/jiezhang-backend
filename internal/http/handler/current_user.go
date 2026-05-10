package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
)

func fetchCurrentUser(c *gin.Context) (domain.User, domain.AccountBook, bool) {
	user, ok := middleware.CurrentUser(c)
	accountBook, ok2 := middleware.AccountBook(c)
	if ok && ok2 {
		return user, accountBook, true
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"status": 301,
		"msg":    "session key overdue",
	})
	return domain.User{}, domain.AccountBook{}, false
}
