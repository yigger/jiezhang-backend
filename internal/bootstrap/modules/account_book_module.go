package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildAccountBookModule(db *gorm.DB) (handler.AccountBookHandler, *mysqlrepo.AccountBookRepository, error) {
	accountBookRepo, err := mysqlrepo.NewAccountBookRepository(db)
	if err != nil {
		return handler.AccountBookHandler{}, nil, fmt.Errorf("init account book repository: %w", err)
	}

	accountBookService := service.NewAccountBookService(accountBookRepo)
	accountBookHandler := handler.NewAccountBookHandler(accountBookService)
	return accountBookHandler, accountBookRepo, nil
}
