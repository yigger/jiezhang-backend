package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func notImplemented(c *gin.Context, endpoint string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"status": 501,
		"msg":    "not implemented",
		"data": gin.H{
			"endpoint": endpoint,
		},
	})
}
