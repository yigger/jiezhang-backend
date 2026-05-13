package router

import (
	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
)

func Register(
	engine *gin.Engine,
	authHandler handler.AuthHandler,
	userHandler handler.UserHandler,
	authMiddleware gin.HandlerFunc,
	homeHandler handler.HomeHandler,
	statementsHandler handler.StatementsHandler,
	financesHandler handler.FinancesHandler,
	categoriesHandler handler.CategoriesHandler,
	assetsHandler handler.AssetsHandler,
	accountBookHandler handler.AccountBookHandler,
	budgetsHandler handler.BudgetsHandler,
	messagesHandler handler.MessagesHandler,
	payeesHandler handler.PayeesHandler,
	friendsHandler handler.FriendsHandler,
	settingsHandler handler.SettingsHandler,
	superStatementsHandler handler.SuperStatementsHandler,
	superChartHandler handler.SuperChartHandler,
	statisticsHandler handler.StatisticsHandler,
) {

	api := engine.Group("/api")
	{
		api.POST("/check_openid", authHandler.CheckOpenID)

		authRequired := api.Group("/")
		authRequired.Use(authMiddleware)
		{
			authRequired.POST("/upload", authHandler.Upload)

			authRequired.GET("/header", homeHandler.Header)
			authRequired.GET("/index", homeHandler.Index)
			authRequired.GET("/settings", homeHandler.GetSettings)

			authRequired.GET("/users", userHandler.GetUserInfo)
			authRequired.PUT("/users/update_user", userHandler.UpdateUser)
			authRequired.POST("/users/scan_login", userHandler.ScanLogin)

			authRequired.GET("/statements/categories", statementsHandler.Categories)
			authRequired.GET("/statements/assets", statementsHandler.Assets)
			authRequired.GET("/statements/category_frequent", statementsHandler.CategoryFrequent)
			authRequired.GET("/statements/asset_frequent", statementsHandler.AssetFrequent)
			authRequired.GET("/statements", statementsHandler.List)
			authRequired.GET("/statements/list_by_token", statementsHandler.ListByToken)
			authRequired.POST("/statements", statementsHandler.Create)
			authRequired.PUT("/statements/:statementId", statementsHandler.Update)
			authRequired.GET("/statements/:statementId", statementsHandler.Show)
			authRequired.DELETE("/statements/:statementId", statementsHandler.Delete)
			authRequired.GET("/search", statementsHandler.Search)
			authRequired.GET("/statements/images", statementsHandler.Images)
			authRequired.POST("/statements/generate_share_key", statementsHandler.GenerateShareKey)
			authRequired.POST("/statements/export_check", statementsHandler.ExportCheck)
			authRequired.GET("/statements/target_objects", statementsHandler.TargetObjects)
			authRequired.DELETE("/statements/:statementId/avatar", statementsHandler.RemoveAvatar)
			authRequired.GET("/statements/default_category_asset", statementsHandler.DefaultCategoryAsset)
			authRequired.GET("/statements/export_excel", statementsHandler.ExportExcel)

			authRequired.GET("/categories/category_list", categoriesHandler.List)
			authRequired.GET("/categories/parent", categoriesHandler.Parent)
			authRequired.GET("/categories/category_childs", categoriesHandler.CategoryChilds)
			authRequired.GET("/categories/category_statements", categoriesHandler.CategoryStatements)
			authRequired.GET("/categories/:id", categoriesHandler.Show)
			authRequired.DELETE("/categories/:id", categoriesHandler.Delete)
			authRequired.GET("/icons/categories_with_url", categoriesHandler.Icons)
			authRequired.PUT("/categories/:id", categoriesHandler.Update)
			authRequired.POST("/categories", categoriesHandler.Create)

			authRequired.GET("/assets", assetsHandler.List)
			authRequired.GET("/assets/:id", assetsHandler.Show)
			authRequired.DELETE("/assets/:id", assetsHandler.Delete)
			authRequired.GET("/icons/assets_with_url", assetsHandler.Icons)
			authRequired.PUT("/assets/:id", assetsHandler.Update)
			authRequired.POST("/assets", assetsHandler.Create)
			authRequired.PUT("/wallet/surplus", assetsHandler.UpdateSurplus)

			authRequired.GET("/account_books", accountBookHandler.List)
			authRequired.GET("/account_books/:id", accountBookHandler.Show)
			authRequired.GET("/account_books/types", accountBookHandler.Types)
			authRequired.GET("/account_books/preset_categories", accountBookHandler.PresetCategories)
			authRequired.PUT("/account_books/:id/switch", accountBookHandler.Switch)
			authRequired.POST("/account_books", accountBookHandler.Create)
			authRequired.PUT("/account_books/:id", accountBookHandler.Update)
			authRequired.DELETE("/account_books/:id", accountBookHandler.Delete)

			authRequired.GET("/wallet", financesHandler.Wallet)
			authRequired.GET("/wallet/information", financesHandler.WalletInformation)
			authRequired.GET("/wallet/time_line", financesHandler.WalletTimeline)
			authRequired.GET("/wallet/statement_list", financesHandler.WalletStatementList)

			authRequired.GET("/budgets", budgetsHandler.Summary)
			authRequired.GET("/budgets/parent", budgetsHandler.ParentList)
			authRequired.GET("/budgets/:categoryId", budgetsHandler.CategoryBudget)
			authRequired.PUT("/budgets/0", budgetsHandler.UpdateAmount)

			authRequired.GET("/chart/calendar_data", statisticsHandler.CalendarData)
			authRequired.GET("/chart/overview_header", statisticsHandler.OverviewHeader)
			authRequired.GET("/chart/overview_statements", statisticsHandler.OverviewStatements)
			authRequired.GET("/chart/rate", statisticsHandler.Rate)

			authRequired.GET("/super_statements/time", superStatementsHandler.Time)
			authRequired.GET("/super_statements/list", superStatementsHandler.List)

			authRequired.GET("/super_chart/header", superChartHandler.Header)
			authRequired.GET("/super_chart/get_pie_data", superChartHandler.PieData)
			authRequired.GET("/super_chart/week_data", superChartHandler.WeekData)
			authRequired.GET("/super_chart/line_chart", superChartHandler.LineChart)
			authRequired.GET("/super_chart/categories_list", superChartHandler.CategoriesList)
			authRequired.GET("/super_chart/table_sumary", superChartHandler.TableSummary)

			authRequired.GET("/message", messagesHandler.List)
			authRequired.GET("/message/:id", messagesHandler.Show)

			authRequired.GET("/payees", payeesHandler.List)
			authRequired.POST("/payees", payeesHandler.Create)
			authRequired.PUT("/payees/:id", payeesHandler.Update)
			authRequired.DELETE("/payees/:id", payeesHandler.Delete)

			authRequired.GET("/friends", friendsHandler.List)
			authRequired.POST("/friends/invite", friendsHandler.Invite)
			authRequired.GET("/friends/invite_information", friendsHandler.InviteInformation)
			authRequired.POST("/friends/accept_apply", friendsHandler.AcceptApply)
			authRequired.DELETE("/friends/:collaboratorId", friendsHandler.Remove)
			authRequired.PUT("/friends/:collaboratorId", friendsHandler.Update)

			authRequired.POST("/settings/feedback", settingsHandler.Feedback)
		}
	}
}
