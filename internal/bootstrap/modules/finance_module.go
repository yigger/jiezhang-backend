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

func BuildFinanceModule(db *gorm.DB, publicBaseURL string) (handler.FinancesHandler, error) {
	financeRepo, err := mysqlrepo.NewFinanceRepository(db)
	if err != nil {
		return handler.FinancesHandler{}, fmt.Errorf("init finance repository: %w", err)
	}

	statementRepo, err := mysqlrepo.NewStatementRepository(db)
	if err != nil {
		return handler.FinancesHandler{}, fmt.Errorf("init statement repository: %w", err)
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	rowMapper := statementdto.NewRowMapper(publicURLBuilder)
	financeService := service.NewFinanceService(financeRepo, statementRepo, rowMapper)

	return handler.NewFinancesHandler(financeService), nil
}
