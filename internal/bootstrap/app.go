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
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/wechat"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
	authservice "github.com/yigger/jiezhang-backend/internal/service/auth"
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
	if cfg.MiniProgramAppID == "" || cfg.MiniProgramSecret == "" {
		log.Fatal("MINIPROGRAM_APPID and MINIPROGRAM_SECRET are required")
	}
	if cfg.SessionTokenSecret == "" {
		log.Fatal("SESSION_TOKEN_SECRET is required")
	}

	mysqlDB, err := db.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v", err)
	}

	userRepo, err := mysqlrepo.NewUserRepository(mysqlDB)
	if err != nil {
		log.Fatalf("failed to init user repository: %v", err)
	}
	statementRepo, err := mysqlrepo.NewStatementRepository(mysqlDB)
	if err != nil {
		log.Fatalf("failed to init statement repository: %v", err)
	}

	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	wechatClient := wechat.NewHTTPClient(cfg.MiniProgramAppID, cfg.MiniProgramSecret)
	sessionCache := sessioncache.NewMemoryCache()
	checkOpenIDService := authservice.NewCheckOpenIDService(
		userRepo,
		wechatClient,
		cfg.SessionTokenSecret,
		sessionCache,
	)
	authHandler := handler.NewAuthHandler(checkOpenIDService)

	// 账单相关
	statementService := service.NewStatementService(statementRepo)
	statementsHandler := handler.NewStatementsHandler(statementService)

	authMiddleware := middleware.AuthenticateAPIV1(cfg.Env, cfg.MiniProgramAppID, userRepo, sessionCache)
	router.Register(engine, authHandler, userHandler, authMiddleware, statementsHandler)

	return &App{cfg: cfg, engine: engine, db: mysqlDB}
}

func (a *App) Run() error {
	return a.engine.Run(a.cfg.ListenAddr())
}
