package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

const currentUserContextKey = "current_user"

func AuthenticateAPIV1(env, appID string, users repository.UserRepository, cache sessioncache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerAppID := strings.TrimSpace(c.GetHeader("X-WX-APP-ID"))
		if headerAppID == "" || headerAppID != appID {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"status": 404,
				"msg":    "invalid appid",
			})
			return
		}

		// .env 的开发环境，默认找 ID 为 1 的用户
		if env == "dev" {
			user, err := users.FindByID(c.Request.Context(), 1)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"status": 404,
					"msg":    "[dev] user=1 not found",
				})
				return
			}
			c.Set(currentUserContextKey, user)
			c.Next()
			return
		}

		thirdSession := strings.TrimSpace(c.GetHeader("X-WX-Skey"))
		if thirdSession == "" {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"status": 301,
				"msg":    "session key overdue",
			})
			return
		}

		user, err := users.FindByThirdSession(c.Request.Context(), thirdSession)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"status": 301,
				"msg":    "session key overdue",
			})
			return
		}

		cachedSession, ok := cache.Get(user.RedisSessionKey())
		if !ok || strings.TrimSpace(cachedSession) == "" || cachedSession != thirdSession {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"status": 301,
				"msg":    "session key overdue",
			})
			return
		}

		c.Set(currentUserContextKey, user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (domain.User, bool) {
	v, ok := c.Get(currentUserContextKey)
	if !ok {
		return domain.User{}, false
	}

	user, ok := v.(domain.User)
	return user, ok
}
