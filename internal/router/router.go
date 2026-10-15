package router

import (
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/controller"
)

// Handlers names every HTTP entry point so registration does not depend on argument order.
type Handlers struct {
	Auth            controller.AuthHandler
	Users           controller.UserHandler
	Home            controller.HomeHandler
	Statements      controller.StatementsHandler
	Finances        controller.FinancesHandler
	Categories      controller.CategoriesHandler
	Assets          controller.AssetsHandler
	AccountBooks    controller.AccountBookHandler
	Budgets         controller.BudgetsHandler
	Messages        controller.MessagesHandler
	Payees          controller.PayeesHandler
	Friends         controller.FriendsHandler
	Settings        controller.SettingsHandler
	SuperStatements controller.SuperStatementsHandler
	SuperChart      controller.SuperChartHandler
	Statistics      controller.StatisticsHandler
	CalendarJournal controller.CalendarJournalHandler
	Insights        controller.InsightsHandler
}

func Register(engine *gin.Engine, authMiddleware gin.HandlerFunc, handlers Handlers) {
	RegisterCalendarJournal(engine, authMiddleware, handlers.CalendarJournal)
	RegisterInsights(engine, authMiddleware, handlers.Insights)
	RegisterDocumentation(engine)
	api := engine.Group("/api")
	{
		api.POST("/check_openid", handlers.Auth.CheckOpenID)

		authRequired := api.Group("/")
		authRequired.Use(authMiddleware)
		{
			authRequired.POST("/upload", handlers.Auth.Upload)

			authRequired.GET("/header", handlers.Home.Header)
			authRequired.GET("/index", handlers.Home.Index)
			authRequired.GET("/settings", handlers.Home.GetSettings)

			authRequired.GET("/users", handlers.Users.GetUserInfo)
			authRequired.PUT("/users/update_user", handlers.Users.UpdateUser)
			authRequired.POST("/users/scan_login", handlers.Users.ScanLogin)

			authRequired.GET("/statements/categories", handlers.Statements.Categories)
			authRequired.GET("/statements/assets", handlers.Statements.Assets)
			authRequired.GET("/statements/category_frequent", handlers.Statements.CategoryFrequent)
			authRequired.GET("/statements/asset_frequent", handlers.Statements.AssetFrequent)
			authRequired.GET("/statements", handlers.Statements.List)
			authRequired.GET("/statements/list_by_token", handlers.Statements.ListByToken)
			authRequired.POST("/statements", handlers.Statements.Create)
			authRequired.PUT("/statements/:statementId", handlers.Statements.Update)
			authRequired.GET("/statements/:statementId", handlers.Statements.Show)
			authRequired.DELETE("/statements/:statementId", handlers.Statements.Delete)
			authRequired.GET("/search", handlers.Statements.Search)
			authRequired.GET("/statements/images", handlers.Statements.Images)
			authRequired.POST("/statements/generate_share_key", handlers.Statements.GenerateShareKey)
			authRequired.POST("/statements/export_check", handlers.Statements.ExportCheck)
			authRequired.GET("/statements/target_objects", handlers.Statements.TargetObjects)
			authRequired.DELETE("/statements/:statementId/avatar", handlers.Statements.RemoveAvatar)
			authRequired.GET("/statements/default_category_asset", handlers.Statements.DefaultCategoryAsset)
			authRequired.GET("/statements/export_excel", handlers.Statements.ExportExcel)

			authRequired.GET("/categories/category_list", handlers.Categories.List)
			authRequired.GET("/categories/parent", handlers.Categories.Parent)
			authRequired.GET("/categories/category_childs", handlers.Categories.CategoryChilds)
			authRequired.GET("/categories/category_statements", handlers.Categories.CategoryStatements)
			authRequired.GET("/categories/:id", handlers.Categories.Show)
			authRequired.DELETE("/categories/:id", handlers.Categories.Delete)
			authRequired.GET("/icons/categories_with_url", handlers.Categories.Icons)
			authRequired.PUT("/categories/:id", handlers.Categories.Update)
			authRequired.POST("/categories", handlers.Categories.Create)

			authRequired.GET("/assets", handlers.Assets.List)
			authRequired.GET("/assets/:id", handlers.Assets.Show)
			authRequired.DELETE("/assets/:id", handlers.Assets.Delete)
			authRequired.GET("/icons/assets_with_url", handlers.Assets.Icons)
			authRequired.PUT("/assets/:id", handlers.Assets.Update)
			authRequired.POST("/assets", handlers.Assets.Create)
			authRequired.PUT("/wallet/surplus", handlers.Assets.UpdateSurplus)

			authRequired.GET("/account_books", handlers.AccountBooks.List)
			authRequired.GET("/account_books/:id", handlers.AccountBooks.Show)
			authRequired.GET("/account_books/types", handlers.AccountBooks.Types)
			authRequired.GET("/account_books/preset_categories", handlers.AccountBooks.PresetCategories)
			authRequired.PUT("/account_books/:id/switch", handlers.AccountBooks.Switch)
			authRequired.POST("/account_books", handlers.AccountBooks.Create)
			authRequired.PUT("/account_books/:id", handlers.AccountBooks.Update)
			authRequired.DELETE("/account_books/:id", handlers.AccountBooks.Delete)

			authRequired.GET("/wallet", handlers.Finances.Wallet)
			authRequired.GET("/wallet/information", handlers.Finances.WalletInformation)
			authRequired.GET("/wallet/time_line", handlers.Finances.WalletTimeline)
			authRequired.GET("/wallet/statement_list", handlers.Finances.WalletStatementList)

			authRequired.GET("/budgets", handlers.Budgets.Summary)
			authRequired.GET("/budgets/parent", handlers.Budgets.ParentList)
			authRequired.GET("/budgets/:categoryId", handlers.Budgets.CategoryBudget)
			authRequired.PUT("/budgets/0", handlers.Budgets.UpdateAmount)

			authRequired.GET("/chart/calendar_data", handlers.Statistics.CalendarData)
			authRequired.GET("/chart/overview_header", handlers.Statistics.OverviewHeader)
			authRequired.GET("/chart/overview_statements", handlers.Statistics.OverviewStatements)
			authRequired.GET("/chart/rate", handlers.Statistics.Rate)

			authRequired.GET("/super_statements/time", handlers.SuperStatements.Time)
			authRequired.GET("/super_statements/list", handlers.SuperStatements.List)

			authRequired.GET("/super_chart/header", handlers.SuperChart.Header)
			authRequired.GET("/super_chart/get_pie_data", handlers.SuperChart.PieData)
			authRequired.GET("/super_chart/week_data", handlers.SuperChart.WeekData)
			authRequired.GET("/super_chart/line_chart", handlers.SuperChart.LineChart)
			authRequired.GET("/super_chart/categories_list", handlers.SuperChart.CategoriesList)
			authRequired.GET("/super_chart/table_sumary", handlers.SuperChart.TableSummary)

			authRequired.GET("/message", handlers.Messages.List)
			authRequired.GET("/message/:id", handlers.Messages.Show)

			authRequired.GET("/payees", handlers.Payees.List)
			authRequired.POST("/payees", handlers.Payees.Create)
			authRequired.PUT("/payees/:id", handlers.Payees.Update)
			authRequired.DELETE("/payees/:id", handlers.Payees.Delete)

			authRequired.GET("/friends", handlers.Friends.List)
			authRequired.POST("/friends/invite", handlers.Friends.Invite)
			authRequired.GET("/friends/invite_information", handlers.Friends.InviteInformation)
			authRequired.POST("/friends/accept_apply", handlers.Friends.AcceptApply)
			authRequired.DELETE("/friends/:collaboratorId", handlers.Friends.Remove)
			authRequired.PUT("/friends/:collaboratorId", handlers.Friends.Update)

			authRequired.POST("/settings/feedback", handlers.Settings.Feedback)
		}
	}
}
