package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildBudgetModule(db *gorm.DB, publicBaseURL string) (handler.BudgetsHandler, error) {
	budgetRepo, err := mysqlrepo.NewBudgetRepository(db)
	if err != nil {
		return handler.BudgetsHandler{}, fmt.Errorf("init budget repository: %w", err)
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	budgetService := service.NewBudgetService(budgetRepo, publicURLBuilder)
	return handler.NewBudgetsHandler(budgetService), nil
}
