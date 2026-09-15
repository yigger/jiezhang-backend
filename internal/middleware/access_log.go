package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLog prints a compact line for each request.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()

		c.Next()

		latency := time.Since(startedAt)
		log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
	}
}
