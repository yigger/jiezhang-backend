package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/yigger/jiezhang-backend/docs/swagger"
)

func RegisterDocumentation(engine *gin.Engine) {
	handler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	engine.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "/" || c.Param("any") == "" {
			c.Redirect(http.StatusFound, "/swagger/index.html")
			return
		}
		handler(c)
	})
}
