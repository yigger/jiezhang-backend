package bootstrap

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/bootstrap/modules"
	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
	"github.com/yigger/jiezhang-backend/internal/http/router"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/db"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/signedurl"
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

	signer := signedurl.NewSigner(cfg.SessionTokenSecret)
	engine.Use(middleware.RequireSignedURL(signer))
	registerStaticFiles(engine)

	mysqlDB, err := db.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v", err)
	}

	sessionCache := newSessionCache(cfg.RedisURL)
	userHandler, userRepo, err := modules.BuildUserModule(mysqlDB, sessionCache)
	if err != nil {
		log.Fatalf("failed to build user module: %v", err)
	}

	authModule, err := modules.BuildAuthModule(cfg, mysqlDB, userRepo, sessionCache)
	if err != nil {
		log.Fatalf("failed to build auth module: %v", err)
	}
	homeHandler, err := modules.BuildHomeModule(mysqlDB, cfg.PublicBaseURL)
	if err != nil {
		log.Fatalf("failed to build home module: %v", err)
	}
	statementsHandler, err := modules.BuildStatementModule(mysqlDB, cfg, sessionCache)
	if err != nil {
		log.Fatalf("failed to build statement module: %v", err)
	}
	financesHandler, err := modules.BuildFinanceModule(mysqlDB, cfg.PublicBaseURL)
	if err != nil {
		log.Fatalf("failed to build finance module: %v", err)
	}
	categoriesHandler, err := modules.BuildCategoryModule(mysqlDB, cfg.PublicBaseURL)
	if err != nil {
		log.Fatalf("failed to build category module: %v", err)
	}
	assetsHandler, err := modules.BuildAssetModule(mysqlDB, cfg.PublicBaseURL)
	if err != nil {
		log.Fatalf("failed to build asset module: %v", err)
	}
	payeesHandler, err := modules.BuildPayeeModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build payee module: %v", err)
	}
	friendsHandler, err := modules.BuildFriendModule(mysqlDB, cfg.PublicBaseURL, cfg.SessionTokenSecret)
	if err != nil {
		log.Fatalf("failed to build friend module: %v", err)
	}
	budgetsHandler, err := modules.BuildBudgetModule(mysqlDB, cfg.PublicBaseURL)
	if err != nil {
		log.Fatalf("failed to build budget module: %v", err)
	}
	messagesHandler, err := modules.BuildMessageModule(mysqlDB, cfg.PublicBaseURL)
	if err != nil {
		log.Fatalf("failed to build message module: %v", err)
	}
	settingsHandler, err := modules.BuildSettingModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build setting module: %v", err)
	}
	superStatementsHandler, err := modules.BuildSuperStatementModule(mysqlDB, cfg.PublicBaseURL)
	if err != nil {
		log.Fatalf("failed to build super statement module: %v", err)
	}
	superChartHandler, err := modules.BuildSuperChartModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build super chart module: %v", err)
	}

	accountBookHandler, accountBookRepo, err := modules.BuildAccountBookModule(mysqlDB)
	if err != nil {
		log.Fatalf("failed to build account book module: %v", err)
	}

	statisticHandler, err := modules.BuildStatisticModule(mysqlDB, cfg.PublicBaseURL)
	if err != nil {
		log.Fatalf("failed to build statistic module: %v", err)
	}

	authMiddleware := middleware.AuthenticateAPIV1(cfg.Env, cfg.MiniProgramAppID, userRepo, accountBookRepo, authModule.SessionCache)
	router.Register(engine, authModule.Handler,
		userHandler, authMiddleware,
		homeHandler, statementsHandler, financesHandler, categoriesHandler, assetsHandler,
		accountBookHandler, budgetsHandler, messagesHandler, payeesHandler, friendsHandler, settingsHandler,
		superStatementsHandler, superChartHandler,
		statisticHandler)

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

func registerStaticFiles(engine *gin.Engine) {
	if wd, err := os.Getwd(); err == nil {
		publicDir := filepath.Join(wd, "public")
		if err := os.MkdirAll(publicDir, 0o755); err != nil {
			return
		}
		engine.Static("/images", filepath.Join(publicDir, "images"))
		engine.Static("/private", filepath.Join(publicDir, "private"))
		engine.Static("/public", publicDir)
	}
}

func newSessionCache(redisURL string) sessioncache.Cache {
	if redisURL != "" {
		rc, err := sessioncache.NewRedisCache(redisURL)
		if err != nil {
			log.Printf("redis connect failed, falling back to memory cache: %v", err)
			return sessioncache.NewMemoryCache()
		}
		log.Println("using redis session cache")
		return rc
	}
	return sessioncache.NewMemoryCache()
}
