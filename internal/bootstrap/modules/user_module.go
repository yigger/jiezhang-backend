package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildUserModule(db *gorm.DB, cache sessioncache.Cache) (handler.UserHandler, *mysqlrepo.UserRepository, error) {
	userRepo, err := mysqlrepo.NewUserRepository(db)
	if err != nil {
		return handler.UserHandler{}, nil, fmt.Errorf("init user repository: %w", err)
	}

	userService := service.NewUserService(userRepo, cache)
	userHandler := handler.NewUserHandler(userService)
	return userHandler, userRepo, nil
}
