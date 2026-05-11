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

const CurrentUserContextKey = "current_user"
const AccountBookContextKey = "account_book"

func AuthenticateAPIV1(env, appID string, users repository.UserRepository, accountBooks repository.AccountBookRepository, cache sessioncache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		if env == "dev" {
			user, err := users.FindByID(c.Request.Context(), 1)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"status": 404,
					"msg":    "[dev] user=1 not found",
				})
				return
			}
			c.Set(CurrentUserContextKey, user)
			if !attachAccountBook(c, accountBooks, user, true) {
				return
			}
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

		c.Set(CurrentUserContextKey, user)
		if !attachAccountBook(c, accountBooks, user, false) {
			return
		}
		c.Next()
	}
}

func attachAccountBook(c *gin.Context, accountBooks repository.AccountBookRepository, user domain.User, isDev bool) bool {
	accountBookID, err := resolveAccountBookID(c, user)
	if err != nil {
		msg := "invalid account book id"
		if isDev {
			msg = "[dev] invalid account book id"
		}
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status": 400,
			"msg":    msg,
		})
		return false
	}

	accountBook, err := accountBooks.FindByID(c.Request.Context(), accountBookID, user.ID)
	if err != nil {
		msg := "account book not found"
		if isDev {
			msg = "[dev] account book not found"
		}
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status": 404,
			"msg":    msg,
		})
		return false
	}

	c.Set(AccountBookContextKey, accountBook)
	return true
}

func resolveAccountBookID(c *gin.Context, user domain.User) (int64, error) {
	if v := strings.TrimSpace(c.Query("account_book_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return 0, strconv.ErrSyntax
		}
		return id, nil
	}
	if user.AccountBookId <= 0 {
		return 0, strconv.ErrSyntax
	}
	return user.AccountBookId, nil
}
