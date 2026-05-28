package bootstrap

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/container"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
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

	c, err := container.BuildContainer(cfg, mysqlDB, sessionCache)
	if err != nil {
		log.Fatalf("failed to build container: %v", err)
	}

	if err := c.Invoke(container.RegisterRoutes(engine, cfg)); err != nil {
		log.Fatalf("failed to register routes: %v", err)
	}

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
