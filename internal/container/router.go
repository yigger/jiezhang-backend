package container

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/http/middleware"
	"github.com/yigger/jiezhang-backend/internal/http/router"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/mcp"
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

	MCPServer *mcp.Server
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

		// MCP endpoint (Streamable HTTP)
		mcpGroup := engine.Group("/mcp")
		if cfg.MCPAPIKey != "" {
			mcpGroup.Use(mcpAPIKeyAuth(cfg.MCPAPIKey))
		}
		mcpGroup.Any("", gin.WrapH(p.MCPServer.Handler()))
	}
}

// mcpAPIKeyAuth validates requests using Bearer token against the configured MCP API key.
//
// Sets WWW-Authenticate: Bearer on 401 responses so MCP clients retry with the
// configured token instead of falling back to OAuth discovery.
func mcpAPIKeyAuth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")

		// Debug: log what we actually received
		fmt.Printf("[mcp-auth] method=%s path=%s Authorization=%q\n", c.Request.Method, c.Request.URL.Path, auth)

		if auth == "" {
			c.Header("WWW-Authenticate", `Bearer`)
			c.AbortWithStatusJSON(401, gin.H{"error": "missing Authorization header"})
			return
		}
		const prefix = "Bearer "
		if len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
			c.Header("WWW-Authenticate", `Bearer`)
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid Authorization format, expected: Bearer <key>"})
			return
		}
		token := auth[len(prefix):]
		if token != expectedKey {
			c.AbortWithStatusJSON(403, gin.H{"error": "invalid API key", "hint": fmt.Sprintf("received=%q expected=%q", token, expectedKey)})
			return
		}
		c.Next()
	}
}
