package repository

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/domain"
)

type StatementRepository interface {
	// FindByID(ctx context.Context, id int64) (domain.Statement, error)
	// FindByOpenID(ctx context.Context, openID string) (domain.Statement, error)
	// FindByThirdSession(ctx context.Context, thirdSession string) (domain.Statement, error)
	List(ctx context.Context, userId int64) ([]domain.Statement, error)
	// Create(ctx context.Context, statement domain.Statement) (domain.Statement, error)
	// Save(ctx context.Context, statement domain.Statement) (domain.Statement, error)
}
