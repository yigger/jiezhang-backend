package repository

import (
	"context"
	"errors"

	"github.com/yigger/jiezhang-backend/internal/domain"
)

var ErrPayeeNotFound = errors.New("payee not found")

type PayeeRepository interface {
	ListByAccountBookID(ctx context.Context, accountBookID int64) ([]domain.Payee, error)
	FindByIDAndUserID(ctx context.Context, payeeID int64, userID int64) (domain.Payee, error)
	Create(ctx context.Context, payee domain.Payee) (domain.Payee, error)
	UpdateNameByID(ctx context.Context, payeeID int64, userID int64, name string) (domain.Payee, error)
	DeleteByID(ctx context.Context, payeeID int64, userID int64) error
}
