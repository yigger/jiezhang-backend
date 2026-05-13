package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

func BuildHomeModule(db *gorm.DB, publicBaseURL string) (handler.HomeHandler, error) {
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
	homeRepo, err := mysqlrepo.NewHomeRepository(db)
	if err != nil {
		return handler.HomeHandler{}, fmt.Errorf("init home repository: %w", err)
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	rowMapper := statementdto.NewRowMapper(publicURLBuilder)
	statementService := service.NewStatementService(statementRepo, statementRepo, categoryRepo, assetRepo, rowMapper)
	homeService := service.NewHomeService(homeRepo, statementService)
	return handler.NewHomeHandler(homeService), nil
}
