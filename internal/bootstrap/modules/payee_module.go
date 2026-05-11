package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/http/handler"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
)

func BuildPayeeModule(db *gorm.DB) (handler.PayeesHandler, error) {
	payeeRepo, err := mysqlrepo.NewPayeeRepository(db)
	if err != nil {
		return handler.PayeesHandler{}, fmt.Errorf("init payee repository: %w", err)
	}

	payeeService := service.NewPayeeService(payeeRepo)
	return handler.NewPayeesHandler(payeeService), nil
}
