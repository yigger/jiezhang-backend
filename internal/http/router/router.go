package router

import (
	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
)

func Register(
	engine *gin.Engine,
	authHandler handler.AuthHandler,
	userHandler handler.UserHandler) {

	homeHandler := handler.NewHomeHandler()
	statementsHandler := handler.NewStatementsHandler()
	categoriesHandler := handler.NewCategoriesHandler()
	assetsHandler := handler.NewAssetsHandler()
	accountBooksHandler := handler.NewAccountBooksHandler()
	financesHandler := handler.NewFinancesHandler()
	budgetsHandler := handler.NewBudgetsHandler()
	statisticsHandler := handler.NewStatisticsHandler()
	superStatementsHandler := handler.NewSuperStatementsHandler()
	superChartHandler := handler.NewSuperChartHandler()
	messagesHandler := handler.NewMessagesHandler()
	payeesHandler := handler.NewPayeesHandler()
	friendsHandler := handler.NewFriendsHandler()
	settingsHandler := handler.NewSettingsHandler()

	api := engine.Group("/api/v1")
	{
		api.POST("/check_openid", authHandler.CheckOpenID)
		api.POST("/upload", authHandler.Upload)

		api.GET("/header", homeHandler.Header)
		api.GET("/index", homeHandler.Index)

		api.GET("/settings", userHandler.GetSettings)
		api.GET("/users", userHandler.GetUserInfo)
		api.PUT("/users/update_user", userHandler.UpdateUser)
		api.POST("/users/scan_login", userHandler.ScanLogin)

		api.GET("/statements/categories", statementsHandler.Categories)
		api.GET("/statements/assets", statementsHandler.Assets)
		api.GET("/statements/category_frequent", statementsHandler.CategoryFrequent)
		api.GET("/statements/asset_frequent", statementsHandler.AssetFrequent)
		api.GET("/statements", statementsHandler.List)
		api.GET("/statements/list_by_token", statementsHandler.ListByToken)
		api.POST("/statements", statementsHandler.Create)
		api.PUT("/statements/:statementId", statementsHandler.Update)
		api.GET("/statements/:statementId", statementsHandler.Show)
		api.DELETE("/statements/:statementId", statementsHandler.Delete)
		api.GET("/search", statementsHandler.Search)
		api.GET("/statements/images", statementsHandler.Images)
		api.POST("/statements/generate_share_key", statementsHandler.GenerateShareKey)
		api.POST("/statements/export_check", statementsHandler.ExportCheck)
		api.GET("/statements/target_objects", statementsHandler.TargetObjects)
		api.DELETE("/statements/:statementId/avatar", statementsHandler.RemoveAvatar)
		api.GET("/statements/default_category_asset", statementsHandler.DefaultCategoryAsset)
		api.GET("/statements/export_excel", statementsHandler.ExportExcel)

		api.GET("/categories/category_list", categoriesHandler.List)
		api.GET("/categories/:id", categoriesHandler.Show)
		api.DELETE("/categories/:id", categoriesHandler.Delete)
		api.GET("/icons/categories_with_url", categoriesHandler.Icons)
		api.PUT("/categories/:id", categoriesHandler.Update)
		api.POST("/categories", categoriesHandler.Create)

		api.GET("/assets", assetsHandler.List)
		api.GET("/assets/:id", assetsHandler.Show)
		api.DELETE("/assets/:id", assetsHandler.Delete)
		api.GET("/icons/assets_with_url", assetsHandler.Icons)
		api.PUT("/assets/:id", assetsHandler.Update)
		api.POST("/assets", assetsHandler.Create)
		api.PUT("/wallet/surplus", assetsHandler.UpdateSurplus)

		api.GET("/account_books", accountBooksHandler.List)
		api.GET("/account_books/:id", accountBooksHandler.Show)
		api.GET("/account_books/types", accountBooksHandler.Types)
		api.GET("/account_books/preset_categories", accountBooksHandler.PresetCategories)
		api.PUT("/account_books/:id/switch", accountBooksHandler.Switch)
		api.POST("/account_books", accountBooksHandler.Create)
		api.PUT("/account_books/:id", accountBooksHandler.Update)
		api.DELETE("/account_books/:id", accountBooksHandler.Delete)

		api.GET("/wallet", financesHandler.Wallet)
		api.GET("/wallet/information", financesHandler.WalletInformation)
		api.GET("/wallet/time_line", financesHandler.WalletTimeline)
		api.GET("/wallet/statement_list", financesHandler.WalletStatementList)

		api.GET("/budgets", budgetsHandler.Summary)
		api.GET("/budgets/parent", budgetsHandler.ParentList)
		api.GET("/budgets/:categoryId", budgetsHandler.CategoryBudget)
		api.PUT("/budgets/0", budgetsHandler.UpdateAmount)

		api.GET("/chart/calendar_data", statisticsHandler.CalendarData)
		api.GET("/chart/overview_header", statisticsHandler.OverviewHeader)
		api.GET("/chart/overview_statements", statisticsHandler.OverviewStatements)
		api.GET("/chart/rate", statisticsHandler.Rate)

		api.GET("/super_statements/time", superStatementsHandler.Time)
		api.GET("/super_statements/list", superStatementsHandler.List)

		api.GET("/super_chart/header", superChartHandler.Header)
		api.GET("/super_chart/get_pie_data", superChartHandler.PieData)
		api.GET("/super_chart/week_data", superChartHandler.WeekData)
		api.GET("/super_chart/line_chart", superChartHandler.LineChart)
		api.GET("/super_chart/categories_list", superChartHandler.CategoriesList)
		api.GET("/super_chart/table_sumary", superChartHandler.TableSummary)

		api.GET("/message", messagesHandler.List)
		api.GET("/message/:id", messagesHandler.Show)

		api.GET("/payees", payeesHandler.List)
		api.POST("/payees", payeesHandler.Create)
		api.PUT("/payees/:id", payeesHandler.Update)
		api.DELETE("/payees/:id", payeesHandler.Delete)

		api.GET("/friends", friendsHandler.List)
		api.POST("/friends/invite", friendsHandler.Invite)
		api.GET("/friends/invite_information", friendsHandler.InviteInformation)
		api.POST("/friends/accept_apply", friendsHandler.AcceptApply)
		api.DELETE("/friends/:collaboratorId", friendsHandler.Remove)
		api.PUT("/friends/:collaboratorId", friendsHandler.Update)

		api.POST("/settings/feedback", settingsHandler.Feedback)
	}
}
