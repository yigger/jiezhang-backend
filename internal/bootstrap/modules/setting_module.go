package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildSettingModule(db *gorm.DB) (handler.SettingsHandler, error) {
	settingRepo, err := mysqlrepo.NewSettingRepository(db)
	if err != nil {
		return handler.SettingsHandler{}, fmt.Errorf("init setting repository: %w", err)
	}

	settingService := service.NewSettingService(settingRepo)
	return handler.NewSettingsHandler(settingService), nil
}
