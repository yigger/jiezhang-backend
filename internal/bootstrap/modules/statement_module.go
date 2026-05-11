package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildStatementModule(db *gorm.DB, publicBaseURL string) (handler.StatementsHandler, error) {
	statementRepo, err := mysqlrepo.NewStatementRepository(db)
	categoryRepo, err := mysqlrepo.NewCategoryRepository(db)
	assetRepo, err := mysqlrepo.NewAssetRepository(db)
	if err != nil {
		return handler.StatementsHandler{}, fmt.Errorf("init statement repository: %w", err)
	}

	statementService := service.NewStatementService(statementRepo, statementRepo, categoryRepo, assetRepo, publicBaseURL)
	statementsHandler := handler.NewStatementsHandler(statementService)
	return statementsHandler, nil
}
