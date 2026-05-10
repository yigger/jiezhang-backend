package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildStatementModule(db *gorm.DB) (handler.StatementsHandler, error) {
	statementRepo, err := mysqlrepo.NewStatementRepository(db)
	categoryRepo, err := mysqlrepo.NewCategoryRepository(db)
	if err != nil {
		return handler.StatementsHandler{}, fmt.Errorf("init statement repository: %w", err)
	}

	statementService := service.NewStatementService(statementRepo, categoryRepo)
	statementsHandler := handler.NewStatementsHandler(statementService)
	return statementsHandler, nil
}
