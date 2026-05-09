package bootstrap

import (
	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
	"github.com/yigger/jiezhang-backend/internal/http/router"
)

// App represents the HTTP API application.
type App struct {
	cfg    config.Config
	engine *gin.Engine
}

func NewApp() *App {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	engine := gin.New()
	_ = engine.SetTrustedProxies(nil)

	engine.Use(gin.Recovery())
	engine.Use(middleware.AccessLog())

	router.Register(engine)

	return &App{cfg: cfg, engine: engine}
}

func (a *App) Run() error {
	return a.engine.Run(a.cfg.ListenAddr())
}
