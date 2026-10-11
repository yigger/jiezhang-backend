package controller

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestCalendarJournalRequiresAuthenticationBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, method := range []string{"GET", "PUT"} {
		r := gin.New()
		h := CalendarJournalHandler{}
		r.GET("/journal", h.Month)
		r.PUT("/journal", h.Save)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, "/journal", nil))
		if w.Code != 200 || w.Body.String() != "{\"msg\":\"session key overdue\",\"status\":301}" {
			t.Fatalf("%s: %d %s", method, w.Code, w.Body)
		}
	}
}
