package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

func BuildFinanceModule(db *gorm.DB, publicBaseURL string) (handler.FinancesHandler, error) {
	statementRepo, err := mysqlrepo.NewStatementRepository(db)
	if err != nil {
		return handler.FinancesHandler{}, fmt.Errorf("init statement repository: %w", err)
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	rowMapper := statementdto.NewRowMapper(publicURLBuilder)
	return handler.NewFinancesHandler(db, statementRepo, rowMapper), nil
}
