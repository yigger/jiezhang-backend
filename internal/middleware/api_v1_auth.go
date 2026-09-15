package middleware

import (
	"errors"
	"github.com/yigger/jiezhang-backend/internal/types"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	accountbookservice "github.com/yigger/jiezhang-backend/internal/service/accountbook"
	authservice "github.com/yigger/jiezhang-backend/internal/service/auth"
)

const CurrentUserContextKey = "current_user"
const AccountBookContextKey = "account_book"

func AuthenticateAPIV1(development bool, sessions *authservice.SessionService, access *accountbookservice.AccessService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := sessions.Authenticate(c.Request.Context(), c.GetHeader("X-WX-APP-ID"), c.GetHeader("X-WX-Skey"))
		if err != nil {
			status := 301
			if errors.Is(err, authservice.ErrInvalidAppID) || errors.Is(err, authservice.ErrDevUserMissing) {
				status = 404
			}
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": status, "msg": err.Error()})
			return
		}
		c.Set(CurrentUserContextKey, user)
		c.Request = c.Request.WithContext(authservice.WithUser(c.Request.Context(), user))
		bookID, err := resolveAccountBookID(c, user)
		if err != nil {
			msg := "invalid account book id"
			if development {
				msg = "[dev] invalid account book id"
			}
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": 400, "msg": msg})
			return
		}
		book, err := access.Authorize(c.Request.Context(), bookID, user.ID)
		if err != nil {
			msg := "account book not found"
			if development {
				msg = "[dev] account book not found"
			}
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": 404, "msg": msg})
			return
		}
		c.Set(AccountBookContextKey, book)
		c.Next()
	}
}

func resolveAccountBookID(c *gin.Context, user types.UserContext) (int64, error) {
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
