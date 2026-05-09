package repository

import (
	"context"
	"errors"

	"github.com/yigger/jiezhang-backend/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
}
