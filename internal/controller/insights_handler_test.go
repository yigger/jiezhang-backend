package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/middleware"
	"github.com/yigger/jiezhang-backend/internal/model"
	"net/http/httptest"
	"testing"
)

func TestInsightsRequiresBookBeforeServiceAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/report", InsightsHandler{}.Report)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/report", nil))
	if w.Code != 200 || w.Body.String() != "{\"msg\":\"session key overdue\",\"status\":301}" {
		t.Fatalf("missing book: %d %s", w.Code, w.Body)
	}
}
func TestInsightsRejectsMalformedYearBeforeServiceAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(middleware.AccountBookContextKey, model.AccountBook{ID: 8}) })
	r.GET("/report", InsightsHandler{}.Report)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/report?year=oops", nil))
	if w.Code != 400 {
		t.Fatalf("malformed year: %d %s", w.Code, w.Body)
	}
}
