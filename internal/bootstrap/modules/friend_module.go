package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildFriendModule(db *gorm.DB, publicBaseURL string, tokenSecret string) (handler.FriendsHandler, error) {
	friendRepo, err := mysqlrepo.NewFriendRepository(db)
	if err != nil {
		return handler.FriendsHandler{}, fmt.Errorf("init friend repository: %w", err)
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(publicBaseURL)
	friendService := service.NewFriendService(friendRepo, publicURLBuilder, tokenSecret)
	return handler.NewFriendsHandler(friendService), nil
}
