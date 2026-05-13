package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildMessageModule(db *gorm.DB, publicBaseURL string) (handler.MessagesHandler, error) {
	messageRepo, err := mysqlrepo.NewMessageRepository(db)
	if err != nil {
		return handler.MessagesHandler{}, fmt.Errorf("init message repository: %w", err)
	}

	messageService := service.NewMessageService(messageRepo)
	return handler.NewMessagesHandler(messageService, publicBaseURL), nil
}
