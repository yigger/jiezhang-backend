package modules

import (
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
	"gorm.io/gorm"
)

func BuildStatisticModule(db *gorm.DB) (handler.StatisticsHandler, error) {
	statisticsRepo, _ := mysqlrepo.NewStatisticsRepository(db)

	statisticService := service.NewStatisticsService(statisticsRepo)
	statisticHandler := handler.NewStatisticsHandler(statisticService)
	return statisticHandler, nil
}
