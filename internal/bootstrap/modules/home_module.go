package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildHomeModule(db *gorm.DB) (handler.HomeHandler, error) {
	statementRepo, err := mysqlrepo.NewStatementRepository(db)
	if err != nil {
		return handler.HomeHandler{}, fmt.Errorf("init statement repository: %w", err)
	}
	categoryRepo, err := mysqlrepo.NewCategoryRepository(db)
	if err != nil {
		return handler.HomeHandler{}, fmt.Errorf("init category repository: %w", err)
	}
	assetRepo, err := mysqlrepo.NewAssetRepository(db)
	if err != nil {
		return handler.HomeHandler{}, fmt.Errorf("init asset repository: %w", err)
	}

	statementService := service.NewStatementService(statementRepo, statementRepo, categoryRepo, assetRepo)
	return handler.NewHomeHandler(db, statementService), nil
}
