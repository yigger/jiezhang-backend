package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	jzinfrastructure_signedurl "github.com/yigger/jiezhang-backend/internal/infrastructure/signedurl"
)

func RequireSignedURL(signer *jzinfrastructure_signedurl.Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !jzinfrastructure_signedurl.IsPrivatePath(c.Request.URL.Path) {
			c.Next()
			return
		}

		fullURL := c.Request.URL.String()
		if !signer.Verify(fullURL) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status": 401,
				"msg":    jzinfrastructure_signedurl.FriendlyError(),
			})
			return
		}

		// Strip signature params so static file serving sees a clean path.
		c.Request.URL.RawQuery = ""
		c.Next()
	}
}
