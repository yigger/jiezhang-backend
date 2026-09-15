package repo

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrPayeeNotFound = errors.New("payee not found")

type PayeeRepository interface {
	ListByAccountBookID(ctx context.Context, accountBookID int64) ([]tablemodel.Payee, error)
	FindByIDAndUserID(ctx context.Context, payeeID int64, userID int64) (tablemodel.Payee, error)
	Create(ctx context.Context, payee tablemodel.Payee) (tablemodel.Payee, error)
	UpdateNameByID(ctx context.Context, payeeID int64, userID int64, name string) (tablemodel.Payee, error)
	DeleteByID(ctx context.Context, payeeID int64, userID int64) error
}
