package repository

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/domain"
)

type AccountBookBaseItem struct {
	ID int64 `json:"id"`
}

type AccountBookRepository interface {
	FindByID(ctx context.Context, id int64, userId int64) (domain.AccountBook, error)
}
