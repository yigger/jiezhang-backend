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
	"github.com/yigger/jiezhang-backend/internal/infrastructure/billfile"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/db"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/excel"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/filestore"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/fixedcostscheduler"
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
	"github.com/yigger/jiezhang-backend/internal/service/billimport"
	budget "github.com/yigger/jiezhang-backend/internal/service/budget"
	"github.com/yigger/jiezhang-backend/internal/service/calendarjournal"
	category "github.com/yigger/jiezhang-backend/internal/service/category"
	exportservice "github.com/yigger/jiezhang-backend/internal/service/export"
	finance "github.com/yigger/jiezhang-backend/internal/service/finance"
	friend "github.com/yigger/jiezhang-backend/internal/service/friend"
	home "github.com/yigger/jiezhang-backend/internal/service/home"
	"github.com/yigger/jiezhang-backend/internal/service/insights"
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
	fixedCosts     *fixedcostscheduler.Scheduler
	server         *http.Server
	closeResources func() error
}

// NewBillImporter assembles the standalone CLI without starting HTTP, Redis or cron.
func NewBillImporter(dsn string) (*billimport.Service, func() error, error) {
	if dsn == "" {
		return nil, nil, errors.New("MYSQL_DSN is required")
	}
	database, err := db.NewMySQLQuiet(dsn)
	if err != nil {
		return nil, nil, errors.New("cannot connect to MySQL; check MYSQL_DSN")
	}
	sqlDB, err := database.DB()
	if err != nil {
		return nil, nil, errors.New("cannot initialize MySQL connection")
	}
	imports := mysql.NewStatementImportRepository(database)
	service := billimport.New(imports, billfile.Reader{})
	return service, sqlDB.Close, nil
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
	dependencies := assembleDependencies(cfg, mysqlDB, cache)
	setupRoutes(engine, dependencies.authenticate, dependencies.handlers)
	fixedCosts, err := fixedcostscheduler.New(dependencies.runDueFixedCosts)
	if err != nil {
		_ = cleanup()
		return nil, err
	}
	return &App{
		fixedCosts: fixedCosts,
		server: &http.Server{
			Addr:              cfg.ListenAddr(),
			Handler:           engine,
			ReadHeaderTimeout: 10 * time.Second,
		},
		closeResources: cleanup,
	}, nil
}

// applicationDependencies contains the entry points consumed by HTTP and scheduled jobs.
type applicationDependencies struct {
	handlers         router.Handlers
	authenticate     gin.HandlerFunc
	runDueFixedCosts func(context.Context, time.Time) (int, error)
}

// assembleDependencies constructs the application without querying the database.
func assembleDependencies(cfg config.Config, db *gorm.DB, cache sessioncache.Cache) applicationDependencies {
	// Shared repositories and capabilities.
	users := mysql.NewUserRepository(db)
	books := mysql.NewAccountBookRepository(db)
	statements := mysql.NewStatementRepository(db)
	categories := mysql.NewCategoryRepository(db)
	assets := mysql.NewAssetRepository(db)
	signer := signedurl.NewSigner(cfg.SessionTokenSecret)
	urls := urlbuilder.NewPublicURLBuilderWithSigner(cfg.PublicBaseURL, signer)
	shareTokens := sharetoken.Codec{Secret: cfg.SessionTokenSecret}
	mapper := statement.NewRowMapper(urls)

	// Authentication, uploads and user profile.
	wechatClient := wechat.NewHTTPClient(cfg.MiniProgramAppID, cfg.MiniProgramSecret)
	sessionTokens := sessiontoken.Generator{Secret: cfg.SessionTokenSecret}
	login := auth.NewCheckOpenIDService(users, wechatClient, sessionTokens, cache)
	uploadRepo := mysql.NewUploadRepository(db)
	uploads := upload.NewUploadService(users, uploadRepo, statements, urls, filestore.Local{Root: "public"})
	authHandler := controller.NewAuthHandler(login, uploads)
	userService := user.NewUserService(users, cache)
	userHandler := controller.NewUserHandler(userService)
	sessions := auth.NewSessionService(users, cache, cfg.MiniProgramAppID, cfg.Env == "dev")
	access := accountbook.NewAccessService(books)
	authenticate := middleware.AuthenticateAPIV1(cfg.Env == "dev", sessions, access)

	// Statements, sharing and exports.
	reader := statement.NewReader(statements, categories, assets, mapper)
	// The statement repository implements transactions, attachment writes and detail reads.
	writer := statement.NewWriter(statements, statements, statements, categories, mapper)
	shares := sharing.New(reader, users, cache, shareTokens, mapper)
	exports := exportservice.New(statements, cache, excel.Renderer{})
	statementsHandler := controller.NewStatementsHandler(reader, writer, shares, exports)

	// Home and finance.
	homeRepo := mysql.NewHomeRepository(db)
	homeService := home.NewHomeService(homeRepo, reader, cfg.PublicBaseURL)
	homeHandler := controller.NewHomeHandler(homeService)
	financeRepo := mysql.NewFinanceRepository(db)
	financeService := finance.NewFinanceService(financeRepo, statements, mapper)
	financesHandler := controller.NewFinancesHandler(financeService)

	// Categories, assets and account books.
	categoryService := category.NewCategoryService(categories, urls)
	categoriesHandler := controller.NewCategoriesHandler(categoryService)
	assetService := asset.NewAssetService(assets, urls)
	assetsHandler := controller.NewAssetsHandler(assetService)
	bookService := accountbook.NewAccountBookService(books)
	accountBookHandler := controller.NewAccountBookHandler(bookService)

	// Budgets, messages, payees, friends and settings.
	budgetRepo := mysql.NewBudgetRepository(db)
	budgetService := budget.NewBudgetService(budgetRepo, urls)
	budgetsHandler := controller.NewBudgetsHandler(budgetService)
	messageRepo := mysql.NewMessageRepository(db)
	messageService := message.NewMessageService(messageRepo)
	messagesHandler := controller.NewMessagesHandler(messageService, cfg.PublicBaseURL)
	payeeRepo := mysql.NewPayeeRepository(db)
	payeeService := payee.NewPayeeService(payeeRepo)
	payeesHandler := controller.NewPayeesHandler(payeeService)
	friendRepo := mysql.NewFriendRepository(db)
	friendService := friend.NewFriendService(friendRepo, urls, shareTokens)
	friendsHandler := controller.NewFriendsHandler(friendService)
	settingRepo := mysql.NewSettingRepository(db)
	settingService := setting.NewSettingService(settingRepo)
	settingsHandler := controller.NewSettingsHandler(settingService)

	// Reports and statistics.
	superStatementRepo := mysql.NewSuperStatementRepository(db)
	superStatementService := reporting.NewSuperStatementService(superStatementRepo, mapper)
	superStatementsHandler := controller.NewSuperStatementsHandler(superStatementService)
	superChartRepo := mysql.NewSuperChartRepository(db)
	superChartService := reporting.NewSuperChartService(superChartRepo)
	superChartHandler := controller.NewSuperChartHandler(superChartService)
	statisticsRepo := mysql.NewStatisticsRepository(db)
	statisticsService := statistics.NewStatisticsService(statisticsRepo, statements, mapper)
	statisticsHandler := controller.NewStatisticsHandler(statisticsService)

	// Calendar journal.
	journalRepo := mysql.NewCalendarJournalRepository(db)
	journalService := calendarjournal.New(journalRepo)
	journalHandler := controller.NewCalendarJournalHandler(journalService)

	// Insights and the recurring-cost job share the same storage service.
	insightsRepo := mysql.NewInsightsRepository(db)
	insightsStorageRepo := mysql.NewInsightsStorage(db)
	insightsService := insights.New(insightsRepo)
	insightsStorage := insights.NewStorage(insightsStorageRepo, insightsRepo)
	insightsHandler := controller.NewInsightsHandler(insightsService, insightsStorage)

	return applicationDependencies{
		authenticate:     authenticate,
		runDueFixedCosts: insightsStorage.RunDueFixedCosts,
		handlers: router.Handlers{
			Auth:            authHandler,
			Users:           userHandler,
			Home:            homeHandler,
			Statements:      statementsHandler,
			Finances:        financesHandler,
			Categories:      categoriesHandler,
			Assets:          assetsHandler,
			AccountBooks:    accountBookHandler,
			Budgets:         budgetsHandler,
			Messages:        messagesHandler,
			Payees:          payeesHandler,
			Friends:         friendsHandler,
			Settings:        settingsHandler,
			SuperStatements: superStatementsHandler,
			SuperChart:      superChartHandler,
			Statistics:      statisticsHandler,
			CalendarJournal: journalHandler,
			Insights:        insightsHandler,
		},
	}
}

// setupRoutes only binds the assembled HTTP handlers and authentication middleware.
func setupRoutes(engine *gin.Engine, authenticate gin.HandlerFunc, handlers router.Handlers) {
	router.Register(engine, authenticate, handlers)
}

func (a *App) Close() error {
	if a.fixedCosts != nil {
		a.fixedCosts.Stop()
	}
	return a.closeResources()
}
func (a *App) Run(ctx context.Context) error {
	if a.fixedCosts != nil {
		if e := a.fixedCosts.Start(ctx); e != nil {
			return e
		}
		defer a.fixedCosts.Stop()
	}
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
