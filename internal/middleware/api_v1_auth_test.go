package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/service/accountbook"
	"github.com/yigger/jiezhang-backend/internal/service/auth"
)

type sessionUsers struct{}

func (sessionUsers) FindByID(context.Context, int64) (tablemodel.User, error) {
	return tablemodel.User{ID: 7, AccountBookID: 3}, nil
}
func (s sessionUsers) FindByThirdSession(ctx context.Context, _ string) (tablemodel.User, error) {
	return s.FindByID(ctx, 7)
}

type sessionCache struct{}

func (sessionCache) Get(string) (string, bool) { return "session", true }

type bookAccess struct{}

func (bookAccess) FindByID(_ context.Context, book, user int64) (tablemodel.AccountBook, error) {
	if book != 3 || user != 7 {
		return tablemodel.AccountBook{}, errors.New("not a member")
	}
	return tablemodel.AccountBook{ID: 3}, nil
}

func TestAuthenticationHTTPContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, app, key, query string
		status                int
	}{
		{"valid", "app", "session", "", 200},
		{"bad app", "bad", "session", "", 404},
		{"expired", "app", "old-session", "", 301},
		{"invalid book", "app", "session", "?account_book_id=x", 400},
		{"other book", "app", "session", "?account_book_id=4", 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := gin.New()
			e.Use(AuthenticateAPIV1(false, auth.NewSessionService(sessionUsers{}, sessionCache{}, "app", false), accountbook.NewAccessService(bookAccess{})))
			reached := false
			e.GET("/test", func(c *gin.Context) {
				reached = true
				u, err := auth.CurrentUser(c.Request.Context())
				if err != nil || u.ID != 7 {
					t.Error("missing principal")
				}
				book, ok := c.Get(AccountBookContextKey)
				if !ok || book.(tablemodel.AccountBook).ID != 3 {
					t.Error("missing book")
				}
				c.JSON(200, gin.H{"status": 200})
			})
			req := httptest.NewRequest("GET", "/test"+tc.query, nil)
			req.Header.Set("X-WX-APP-ID", tc.app)
			req.Header.Set("X-WX-Skey", tc.key)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			var body struct {
				Status int `json:"status"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if rec.Code != 200 || body.Status != tc.status || reached != (tc.status == 200) {
				t.Fatalf("HTTP=%d body=%s reached=%v", rec.Code, rec.Body.String(), reached)
			}
		})
	}
}
