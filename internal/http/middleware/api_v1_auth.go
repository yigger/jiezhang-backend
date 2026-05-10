package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

const currentUserContextKey = "current_user"
const accountBookContextKey = "account_book"

func AuthenticateAPIV1(env, appID string, users repository.UserRepository, accountBooks repository.AccountBookRepository, cache sessioncache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			// 设置当前会话默认账本，查看请求中是否存在 account_book_id 参数，若存在则使用该账本，否则使用默认账本
			accountBookId := c.Query("account_book_id")
			var accountBookIdInt int64
			if accountBookId == "" {
				accountBookIdInt = user.AccountBookId
			} else {
				accountBookIdInt, err = strconv.ParseInt(accountBookId, 10, 64)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusOK, gin.H{
						"status": 400,
						"msg":    "[dev] invalid account book id",
					})
					return
				}
			}

			accountBook, err := accountBooks.FindByID(c.Request.Context(), accountBookIdInt, user.ID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"status": 404,
					"msg":    "[dev] account book not found",
				})
				return
			}
			c.Set(accountBookContextKey, accountBook)

			c.Next()
			return
		}

		headerAppID := strings.TrimSpace(c.GetHeader("X-WX-APP-ID"))
		if headerAppID == "" || headerAppID != appID {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"status": 404,
				"msg":    "invalid appid",
			})
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

func AccountBook(c *gin.Context) (domain.AccountBook, bool) {
	v, ok := c.Get(accountBookContextKey)
	if !ok {
		return domain.AccountBook{}, false
	}

	accountBook, ok := v.(domain.AccountBook)
	return accountBook, ok
}
