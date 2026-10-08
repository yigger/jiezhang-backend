package router

import (
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/controller"
)

// RegisterInsights adds opt-in analytics endpoints without altering existing routes.
func RegisterInsights(engine *gin.Engine, auth gin.HandlerFunc, handler controller.InsightsHandler) {
	api := engine.Group("/api/insights", auth)
	api.PUT("/merchants", handler.SaveMerchantAliases)
	api.GET("/report", handler.Report)
	api.GET("/workspace", handler.Workspace)
	api.POST("/projects", handler.SaveProject)
	api.POST("/fixed_costs", handler.SaveFixedCost)
	api.PUT("/annotation", handler.SaveAnnotation)
	api.POST("/portfolio", handler.CapturePortfolio)
}
