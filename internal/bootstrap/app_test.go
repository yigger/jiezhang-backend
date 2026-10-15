package bootstrap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestAssembledRoutesKeepAuthentication(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	db, err := gorm.Open(mysqldriver.New(mysqldriver.Config{Conn: database, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	dependencies := assembleDependencies(config.Config{
		Env: "production", MiniProgramAppID: "test-app", MiniProgramSecret: "test-secret",
		SessionTokenSecret: "test-signing-key", PublicBaseURL: "https://example.test",
	}, db, sessioncache.NewMemoryCache())
	if dependencies.runDueFixedCosts == nil {
		t.Fatal("recurring-cost job is not connected")
	}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	setupRoutes(engine, dependencies.authenticate, dependencies.handlers)
	for _, path := range []string{"/api/header", "/api/statements", "/api/calendar/journal", "/api/insights/workspace"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.Header.Set("X-WX-APP-ID", "test-app")
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			var body struct {
				Status int `json:"status"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != http.StatusOK || body.Status != 301 {
				t.Fatalf("route lost authentication: HTTP %d, status %d", response.Code, body.Status)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
