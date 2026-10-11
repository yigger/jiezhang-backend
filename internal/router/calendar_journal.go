package router

import (
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/controller"
)

func RegisterCalendarJournal(engine *gin.Engine, auth gin.HandlerFunc, h controller.CalendarJournalHandler) {
	api := engine.Group("/api/calendar", auth)
	api.GET("/journal", h.Month)
	api.PUT("/journal", h.Save)
}
