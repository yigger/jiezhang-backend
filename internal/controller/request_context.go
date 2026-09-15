package controller

import (
	"github.com/yigger/jiezhang-backend/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
	jzmiddleware "github.com/yigger/jiezhang-backend/internal/middleware"
	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

func RequireCurrentUser(c *gin.Context) (types.UserContext, bool) {
	user, ok := c.Get(jzmiddleware.CurrentUserContextKey)
	if ok {
		return user.(types.UserContext), true
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"status": 301,
		"msg":    "session key overdue",
	})
	return types.UserContext{}, false
}

func RequireAccountBook(c *gin.Context) (tablemodel.AccountBook, bool) {
	accountBook, ok := c.Get(jzmiddleware.AccountBookContextKey)
	if ok {
		return accountBook.(tablemodel.AccountBook), true
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"status": 301,
		"msg":    "session key overdue",
	})
	return tablemodel.AccountBook{}, false
}
