package router

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/controller"
)

func TestRouteContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	Register(engine,
		controller.AuthHandler{},
		controller.UserHandler{},
		func(c *gin.Context) { c.Next() },
		controller.HomeHandler{},
		controller.StatementsHandler{},
		controller.FinancesHandler{},
		controller.CategoriesHandler{},
		controller.AssetsHandler{},
		controller.AccountBookHandler{},
		controller.BudgetsHandler{},
		controller.MessagesHandler{},
		controller.PayeesHandler{},
		controller.FriendsHandler{},
		controller.SettingsHandler{},
		controller.SuperStatementsHandler{},
		controller.SuperChartHandler{},
		controller.StatisticsHandler{})
	RegisterInsights(engine, func(c *gin.Context) { c.Next() }, controller.InsightsHandler{})
	RegisterCalendarJournal(engine, func(c *gin.Context) { c.Next() }, controller.CalendarJournalHandler{})
	var routes []string
	for _, r := range engine.Routes() {
		if strings.HasPrefix(r.Path, "/api/") {
			routes = append(routes, r.Method+" "+r.Path)
		}
	}
	sort.Strings(routes)
	expected, err := os.ReadFile("testdata/routes.golden")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(routes, "\n")+"\n" != string(expected) {
		t.Fatalf("route contract changed:\n%s", strings.Join(routes, "\n"))
	}
}
