package repo

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrMessageNotFound = errors.New("message not found")

type MessageRepository interface {
	ListByUserID(ctx context.Context, userID int64) ([]tablemodel.Message, error)
	FindByIDForUser(ctx context.Context, id int64, userID int64) (tablemodel.Message, error)
	MarkAsRead(ctx context.Context, id int64, userID int64) error
}
