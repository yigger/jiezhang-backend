package modules

import (
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
	"gorm.io/gorm"
)

func BuildStatisticModule(db *gorm.DB, publicBaseURL string) (handler.StatisticsHandler, error) {
	statementRepo, err := mysqlrepo.NewStatementRepository(db)
	if err != nil {
		return handler.StatisticsHandler{}, err
	}
	statisticsRepo, err := mysqlrepo.NewStatisticsRepository(db)
	if err != nil {
		return handler.StatisticsHandler{}, err
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	rowMapper := statementdto.NewRowMapper(publicURLBuilder)
	statisticService := service.NewStatisticsService(statisticsRepo, statementRepo, rowMapper)
	statisticHandler := handler.NewStatisticsHandler(statisticService)
	return statisticHandler, nil
}
