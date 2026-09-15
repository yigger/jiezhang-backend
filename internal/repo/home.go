package repo

import (
	"context"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

type HomeRepository interface {
	SumExpendInRange(ctx context.Context, accountBookID int64, start time.Time, end time.Time) (float64, error)
	GetAccountBookBudget(ctx context.Context, accountBookID int64) (float64, error)
	FindLatestUnreadMessage(ctx context.Context, userID int64) (*tablemodel.Message, error)
	CountUserPersistDays(ctx context.Context, userID int64, accountBookID int64) (int64, error)
}
