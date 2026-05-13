package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildAssetModule(db *gorm.DB, publicBaseURL string) (handler.AssetsHandler, error) {
	assetRepo, err := mysqlrepo.NewAssetRepository(db)
	if err != nil {
		return handler.AssetsHandler{}, fmt.Errorf("init asset repository: %w", err)
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	assetService := service.NewAssetService(assetRepo, publicURLBuilder)
	return handler.NewAssetsHandler(assetService), nil
}
