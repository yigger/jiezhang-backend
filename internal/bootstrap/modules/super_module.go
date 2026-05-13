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

func BuildSuperStatementModule(db *gorm.DB, publicBaseURL string) (handler.SuperStatementsHandler, error) {
	superRepo, err := mysqlrepo.NewSuperStatementRepository(db)
	if err != nil {
		return handler.SuperStatementsHandler{}, fmt.Errorf("init super statement repository: %w", err)
	}
	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	rowMapper := statementdto.NewRowMapper(publicURLBuilder)
	superService := service.NewSuperStatementService(superRepo, rowMapper)
	return handler.NewSuperStatementsHandler(superService), nil
}

func BuildSuperChartModule(db *gorm.DB) (handler.SuperChartHandler, error) {
	superChartRepo, err := mysqlrepo.NewSuperChartRepository(db)
	if err != nil {
		return handler.SuperChartHandler{}, fmt.Errorf("init super chart repository: %w", err)
	}
	superChartService := service.NewSuperChartService(superChartRepo)
	return handler.NewSuperChartHandler(superChartService), nil
}
