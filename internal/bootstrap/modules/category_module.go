package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildCategoryModule(db *gorm.DB, publicBaseURL string) (handler.CategoriesHandler, error) {
	categoryRepo, err := mysqlrepo.NewCategoryRepository(db)
	if err != nil {
		return handler.CategoriesHandler{}, fmt.Errorf("init category repository: %w", err)
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	categoryService := service.NewCategoryService(categoryRepo, publicURLBuilder)
	return handler.NewCategoriesHandler(categoryService), nil
}
