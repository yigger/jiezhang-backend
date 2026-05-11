package bootstrap

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/bootstrap/modules"
	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
	"github.com/yigger/jiezhang-backend/internal/http/router"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/db"
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

	validateRequiredConfig(cfg)

	mysqlDB, err := db.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v", err)
	}

	userHandler, userRepo, err := modules.BuildUserModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build user module: %v", err)
	}

	authModule := modules.BuildAuthModule(cfg, userRepo)
	homeHandler, err := modules.BuildHomeModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build home module: %v", err)
	}
	statementsHandler, err := modules.BuildStatementModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build statement module: %v", err)
	}
	payeesHandler, err := modules.BuildPayeeModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build payee module: %v", err)
	}

	accountBookHandler, accountBookRepo, err := modules.BuildAccountBookModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build account book module: %v", err)
	}

	authMiddleware := middleware.AuthenticateAPIV1(cfg.Env, cfg.MiniProgramAppID, userRepo, accountBookRepo, authModule.SessionCache)
	router.Register(engine, authModule.Handler, userHandler, authMiddleware, homeHandler, statementsHandler, accountBookHandler, payeesHandler)

	return &App{cfg: cfg, engine: engine, db: mysqlDB}
}

func (a *App) Run() error {
	return a.engine.Run(a.cfg.ListenAddr())
}

func validateRequiredConfig(cfg config.Config) {
	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN is required")
	}
	if cfg.MiniProgramAppID == "" || cfg.MiniProgramSecret == "" {
		log.Fatal("MINIPROGRAM_APPID and MINIPROGRAM_SECRET are required")
	}
	if cfg.SessionTokenSecret == "" {
		log.Fatal("SESSION_TOKEN_SECRET is required")
	}
}
