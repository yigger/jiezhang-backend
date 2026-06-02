package container

import (
	"log"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/signedurl"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/wechat"
	"github.com/yigger/jiezhang-backend/internal/mcp"
	"github.com/yigger/jiezhang-backend/internal/repository"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
	authservice "github.com/yigger/jiezhang-backend/internal/service/auth"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"

	"go.uber.org/dig"
)

func BuildContainer(cfg config.Config, db *gorm.DB, cache sessioncache.Cache) (*dig.Container, error) {
	c := dig.New()

	// Config and infrastructure values
	provide(c, func() config.Config { return cfg })
	provide(c, func() *gorm.DB { return db })
	provide(c, func() sessioncache.Cache { return cache })
	provide(c, func() config.TokenSecret { return config.TokenSecret(cfg.SessionTokenSecret) })
	provide(c, func() config.PublicBaseURL { return config.PublicBaseURL(cfg.PublicBaseURL) })
	provide(c, func() config.MiniProgramAppID { return config.MiniProgramAppID(cfg.MiniProgramAppID) })
	provide(c, func() config.MiniProgramSecret { return config.MiniProgramSecret(cfg.MiniProgramSecret) })

	// Infrastructure
	provide(c, func(secret config.TokenSecret) *signedurl.Signer {
		return signedurl.NewSigner(string(secret))
	})
	provide(c, func(signer *signedurl.Signer, base config.PublicBaseURL) urlbuilder.PublicURLBuilder {
		return urlbuilder.NewPublicURLBuilderWithSigner(string(base), signer)
	})
	provide(c, func(b urlbuilder.PublicURLBuilder) statementdto.URLBuilder { return b })
	provide(c, func(appID config.MiniProgramAppID, secret config.MiniProgramSecret) *wechat.HTTPClient {
		return wechat.NewHTTPClient(string(appID), string(secret))
	})
	provide(c, func(c *wechat.HTTPClient) wechat.Client { return c })

	// Repositories (each mapped to its interface(s))
	provide(c, mysqlrepo.NewUserRepository,
		dig.As(new(repository.UserRepository)),
	)
	provide(c, mysqlrepo.NewUploadRepository,
		dig.As(new(repository.UploadRepository)),
	)
	provide(c, mysqlrepo.NewStatementRepository,
		dig.As(new(repository.StatementRepository)),
		dig.As(new(repository.StatementQueryRepository)),
	)
	provide(c, mysqlrepo.NewCategoryRepository,
		dig.As(new(repository.CategoryRepository)),
	)
	provide(c, mysqlrepo.NewAssetRepository,
		dig.As(new(repository.AssetRepository)),
	)
	provide(c, mysqlrepo.NewFinanceRepository,
		dig.As(new(repository.FinanceRepository)),
	)
	provide(c, mysqlrepo.NewHomeRepository,
		dig.As(new(repository.HomeRepository)),
	)
	provide(c, mysqlrepo.NewBudgetRepository,
		dig.As(new(repository.BudgetRepository)),
	)
	provide(c, mysqlrepo.NewAccountBookRepository,
		dig.As(new(repository.AccountBookRepository)),
	)
	provide(c, mysqlrepo.NewPayeeRepository,
		dig.As(new(repository.PayeeRepository)),
	)
	provide(c, mysqlrepo.NewFriendRepository,
		dig.As(new(repository.FriendRepository)),
	)
	provide(c, mysqlrepo.NewMessageRepository,
		dig.As(new(repository.MessageRepository)),
	)
	provide(c, mysqlrepo.NewSettingRepository,
		dig.As(new(repository.SettingRepository)),
	)
	provide(c, mysqlrepo.NewSuperStatementRepository,
		dig.As(new(repository.SuperStatementRepository)),
	)
	provide(c, mysqlrepo.NewSuperChartRepository,
		dig.As(new(repository.SuperChartRepository)),
	)
	provide(c, mysqlrepo.NewStatisticsRepository,
		dig.As(new(repository.StatisticsRepository)),
	)

	// RowMapper
	provide(c, statementdto.NewRowMapper)

	// Services
	provide(c, service.NewUserService)
	provide(c, authservice.NewCheckOpenIDService)
	provide(c, service.NewUploadService)
	provide(c, service.NewAccountBookService)
	provide(c, service.NewStatementServiceWithSession)
	provide(c, service.NewCategoryService)
	provide(c, service.NewAssetService)
	provide(c, service.NewFinanceService)
	provide(c, service.NewBudgetService)
	provide(c, service.NewHomeService)
	provide(c, service.NewMessageService)
	provide(c, service.NewPayeeService)
	provide(c, service.NewFriendService)
	provide(c, service.NewSettingService)
	provide(c, service.NewSuperStatementService)
	provide(c, service.NewSuperChartService)
	provide(c, service.NewStatisticsService)

	// Handlers
	provide(c, handler.NewAuthHandler)
	provide(c, handler.NewUserHandler)
	provide(c, handler.NewHomeHandler)
	provide(c, handler.NewStatementsHandler)
	provide(c, handler.NewFinancesHandler)
	provide(c, handler.NewCategoriesHandler)
	provide(c, handler.NewAssetsHandler)
	provide(c, handler.NewAccountBookHandler)
	provide(c, handler.NewBudgetsHandler)
	provide(c, handler.NewMessagesHandler)
	provide(c, handler.NewPayeesHandler)
	provide(c, handler.NewFriendsHandler)
	provide(c, handler.NewSettingsHandler)
	provide(c, handler.NewSuperStatementsHandler)
	provide(c, handler.NewSuperChartHandler)
	provide(c, handler.NewStatisticsHandler)

	// MCP
	provide(c, mcp.NewServer)

	return c, nil
}

func provide(c *dig.Container, constructor interface{}, opts ...dig.ProvideOption) {
	if err := c.Provide(constructor, opts...); err != nil {
		log.Fatalf("dig: failed to provide %T: %v", constructor, err)
	}
}
