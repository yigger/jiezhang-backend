package container

import (
	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
	"github.com/yigger/jiezhang-backend/internal/http/router"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/repository"
	"go.uber.org/dig"
)

type routerParams struct {
	dig.In

	AuthHandler             handler.AuthHandler
	UserHandler             handler.UserHandler
	HomeHandler             handler.HomeHandler
	StatementsHandler       handler.StatementsHandler
	FinancesHandler         handler.FinancesHandler
	CategoriesHandler       handler.CategoriesHandler
	AssetsHandler           handler.AssetsHandler
	AccountBookHandler      handler.AccountBookHandler
	BudgetsHandler          handler.BudgetsHandler
	MessagesHandler         handler.MessagesHandler
	PayeesHandler           handler.PayeesHandler
	FriendsHandler          handler.FriendsHandler
	SettingsHandler         handler.SettingsHandler
	SuperStatementsHandler  handler.SuperStatementsHandler
	SuperChartHandler       handler.SuperChartHandler
	StatisticsHandler       handler.StatisticsHandler

	UserRepo        repository.UserRepository
	AccountBookRepo repository.AccountBookRepository
	SessionCache    sessioncache.Cache
}

func RegisterRoutes(engine *gin.Engine, cfg config.Config) interface{} {
	return func(p routerParams) {
		authMiddleware := middleware.AuthenticateAPIV1(
			cfg.Env, cfg.MiniProgramAppID,
			p.UserRepo, p.AccountBookRepo, p.SessionCache,
		)
		router.Register(engine,
			p.AuthHandler, p.UserHandler, authMiddleware,
			p.HomeHandler, p.StatementsHandler, p.FinancesHandler,
			p.CategoriesHandler, p.AssetsHandler, p.AccountBookHandler,
			p.BudgetsHandler, p.MessagesHandler, p.PayeesHandler,
			p.FriendsHandler, p.SettingsHandler,
			p.SuperStatementsHandler, p.SuperChartHandler,
			p.StatisticsHandler,
		)
	}
}
