package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/controller"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/db"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/excel"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/filestore"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessiontoken"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sharetoken"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/signedurl"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/wechat"
	"github.com/yigger/jiezhang-backend/internal/middleware"
	"github.com/yigger/jiezhang-backend/internal/repo/mysql"
	"github.com/yigger/jiezhang-backend/internal/router"
	accountbook "github.com/yigger/jiezhang-backend/internal/service/accountbook"
	asset "github.com/yigger/jiezhang-backend/internal/service/asset"
	auth "github.com/yigger/jiezhang-backend/internal/service/auth"
	budget "github.com/yigger/jiezhang-backend/internal/service/budget"
	category "github.com/yigger/jiezhang-backend/internal/service/category"
	exportservice "github.com/yigger/jiezhang-backend/internal/service/export"
	finance "github.com/yigger/jiezhang-backend/internal/service/finance"
	friend "github.com/yigger/jiezhang-backend/internal/service/friend"
	home "github.com/yigger/jiezhang-backend/internal/service/home"
	message "github.com/yigger/jiezhang-backend/internal/service/message"
	payee "github.com/yigger/jiezhang-backend/internal/service/payee"
	reporting "github.com/yigger/jiezhang-backend/internal/service/reporting"
	setting "github.com/yigger/jiezhang-backend/internal/service/setting"
	"github.com/yigger/jiezhang-backend/internal/service/sharing"
	statement "github.com/yigger/jiezhang-backend/internal/service/statement"
	statistics "github.com/yigger/jiezhang-backend/internal/service/statistics"
	"github.com/yigger/jiezhang-backend/internal/service/upload"
	user "github.com/yigger/jiezhang-backend/internal/service/user"
	"gorm.io/gorm"
)

type App struct {
	server         *http.Server
	closeResources func() error
}

func New(cfg config.Config) (*App, error) {
	if cfg.MySQLDSN == "" {
		return nil, errors.New("MYSQL_DSN is required")
	}
	if cfg.MiniProgramAppID == "" || cfg.MiniProgramSecret == "" {
		return nil, errors.New("MINIPROGRAM_APPID and MINIPROGRAM_SECRET are required")
	}
	if cfg.SessionTokenSecret == "" {
		return nil, errors.New("SESSION_TOKEN_SECRET is required")
	}
	gin.SetMode(cfg.GinMode)
	mysqlDB, err := db.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}
	sqlDB, err := mysqlDB.DB()
	if err != nil {
		return nil, err
	}
	cache := sessioncache.Cache(sessioncache.NewMemoryCache())
	var closeCache func() error
	if cfg.RedisURL != "" {
		rc, e := sessioncache.NewRedisCache(cfg.RedisURL)
		if e != nil {
			log.Printf("redis unavailable, falling back to memory cache: %v", e)
		} else {
			cache = rc
			closeCache = rc.Close
		}
	}
	cleanup := func() error {
		var e error
		if closeCache != nil {
			e = closeCache()
		}
		return errors.Join(e, sqlDB.Close())
	}
	engine := gin.New()
	if err = engine.SetTrustedProxies(nil); err != nil {
		_ = cleanup()
		return nil, err
	}
	engine.Use(gin.Recovery(), middleware.AccessLog(), middleware.RequireSignedURL(signedurl.NewSigner(cfg.SessionTokenSecret)))
	publicDir, err := filepath.Abs("public")
	if err != nil {
		_ = cleanup()
		return nil, err
	}
	if err = os.MkdirAll(publicDir, 0755); err != nil {
		_ = cleanup()
		return nil, err
	}
	engine.Static("/images", filepath.Join(publicDir, "images"))
	engine.Static("/private", filepath.Join(publicDir, "private"))
	engine.Static("/public", publicDir)
	setupRoutes(engine, cfg, mysqlDB, cache)
	return &App{server: &http.Server{Addr: cfg.ListenAddr(), Handler: engine, ReadHeaderTimeout: 10 * time.Second}, closeResources: cleanup}, nil
}

// setupRoutes assembles dependencies explicitly and registers routes without querying the database.
func setupRoutes(engine *gin.Engine, cfg config.Config, db *gorm.DB, cache sessioncache.Cache) {
	users := mysql.NewUserRepository(db)
	books := mysql.NewAccountBookRepository(db)
	statements := mysql.NewStatementRepository(db)
	categories := mysql.NewCategoryRepository(db)
	assets := mysql.NewAssetRepository(db)
	signer := signedurl.NewSigner(cfg.SessionTokenSecret)
	urls := urlbuilder.NewPublicURLBuilderWithSigner(cfg.PublicBaseURL, signer)
	shareTokens := sharetoken.Codec{Secret: cfg.SessionTokenSecret}
	mapper := statement.NewRowMapper(urls)
	reader := statement.NewReader(statements, categories, assets, mapper)
	writer := statement.NewWriter(statements, statements, statements, categories, mapper)
	shares := sharing.New(reader, users, cache, shareTokens, mapper)
	exports := exportservice.New(statements, cache, excel.Renderer{})
	wechatClient := wechat.NewHTTPClient(cfg.MiniProgramAppID, cfg.MiniProgramSecret)
	sessionTokens := sessiontoken.Generator{Secret: cfg.SessionTokenSecret}
	login := auth.NewCheckOpenIDService(users, wechatClient, sessionTokens, cache)
	uploadRepo := mysql.NewUploadRepository(db)
	uploads := upload.NewUploadService(users, uploadRepo, statements, urls, filestore.Local{Root: "public"})
	// Build each service explicitly so its dependencies are visible at the call site.
	userService := user.NewUserService(users, cache)
	sessions := auth.NewSessionService(users, cache, cfg.MiniProgramAppID, cfg.Env == "dev")
	access := accountbook.NewAccessService(books)
	authenticate := middleware.AuthenticateAPIV1(cfg.Env == "dev", sessions, access)

	homeRepo := mysql.NewHomeRepository(db)
	homeService := home.NewHomeService(homeRepo, reader, cfg.PublicBaseURL)
	financeRepo := mysql.NewFinanceRepository(db)
	financeService := finance.NewFinanceService(financeRepo, statements, mapper)
	categoryService := category.NewCategoryService(categories, urls)
	assetService := asset.NewAssetService(assets, urls)
	bookService := accountbook.NewAccountBookService(books)
	budgetRepo := mysql.NewBudgetRepository(db)
	budgetService := budget.NewBudgetService(budgetRepo, urls)
	messageRepo := mysql.NewMessageRepository(db)
	messageService := message.NewMessageService(messageRepo)
	payeeRepo := mysql.NewPayeeRepository(db)
	payeeService := payee.NewPayeeService(payeeRepo)
	friendRepo := mysql.NewFriendRepository(db)
	friendService := friend.NewFriendService(friendRepo, urls, shareTokens)
	settingRepo := mysql.NewSettingRepository(db)
	settingService := setting.NewSettingService(settingRepo)
	superStatementRepo := mysql.NewSuperStatementRepository(db)
	superStatementService := reporting.NewSuperStatementService(superStatementRepo, mapper)
	superChartRepo := mysql.NewSuperChartRepository(db)
	superChartService := reporting.NewSuperChartService(superChartRepo)
	statisticsRepo := mysql.NewStatisticsRepository(db)
	statisticsService := statistics.NewStatisticsService(statisticsRepo, statements, mapper)

	authHandler := controller.NewAuthHandler(login, uploads)
	userHandler := controller.NewUserHandler(userService)
	homeHandler := controller.NewHomeHandler(homeService)
	statementsHandler := controller.NewStatementsHandler(reader, writer, shares, exports)
	financesHandler := controller.NewFinancesHandler(financeService)
	categoriesHandler := controller.NewCategoriesHandler(categoryService)
	assetsHandler := controller.NewAssetsHandler(assetService)
	accountBookHandler := controller.NewAccountBookHandler(bookService)
	budgetsHandler := controller.NewBudgetsHandler(budgetService)
	messagesHandler := controller.NewMessagesHandler(messageService, cfg.PublicBaseURL)
	payeesHandler := controller.NewPayeesHandler(payeeService)
	friendsHandler := controller.NewFriendsHandler(friendService)
	settingsHandler := controller.NewSettingsHandler(settingService)
	superStatementsHandler := controller.NewSuperStatementsHandler(superStatementService)
	superChartHandler := controller.NewSuperChartHandler(superChartService)
	statisticsHandler := controller.NewStatisticsHandler(statisticsService)

	router.Register(engine,
		authHandler, userHandler, authenticate, homeHandler, statementsHandler,
		financesHandler, categoriesHandler, assetsHandler, accountBookHandler,
		budgetsHandler, messagesHandler, payeesHandler, friendsHandler,
		settingsHandler, superStatementsHandler, superChartHandler, statisticsHandler,
	)
}

func (a *App) Close() error { return a.closeResources() }
func (a *App) Run(ctx context.Context) error {
	done := make(chan error, 1)
	go func() { done <- a.server.ListenAndServe() }()
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			_ = a.server.Close()
			return err
		}
		err := <-done
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
func Main() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	a, err := New(config.Load())
	if err != nil {
		return err
	}
	defer a.Close()
	return a.Run(ctx)
}
