package bootstrap

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
	"github.com/yigger/jiezhang-backend/internal/http/router"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/db"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

// App represents the HTTP API application.
type App struct {
	cfg    config.Config
	engine *gin.Engine
	db     *gorm.DB
}

func NewApp() *App {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	engine := gin.New()
	_ = engine.SetTrustedProxies(nil)

	engine.Use(gin.Recovery())
	engine.Use(middleware.AccessLog())

	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN is required")
	}

	mysqlDB, err := db.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v", err)
	}

	userRepo, err := mysqlrepo.NewUserRepository(mysqlDB)
	if err != nil {
		log.Fatalf("failed to init user repository: %v", err)
	}

	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	usersAPIHandler := handler.NewUsersAPIHandler(userHandler)

	router.Register(engine, usersAPIHandler)

	return &App{cfg: cfg, engine: engine, db: mysqlDB}
}

func (a *App) Run() error {
	return a.engine.Run(a.cfg.ListenAddr())
}
