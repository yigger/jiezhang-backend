package repository

import (
	"context"
	"time"
)

type HomeRepository interface {
	SumExpendInRange(ctx context.Context, accountBookID int64, start time.Time, end time.Time) (float64, error)
	GetAccountBookBudget(ctx context.Context, accountBookID int64) (float64, error)
	FindLatestUnreadMessage(ctx context.Context, userID int64) (*HomeMessageRecord, error)
	CountUserPersistDays(ctx context.Context, userID int64, accountBookID int64) (int64, error)
}

type HomeMessageRecord struct {
	ID       int64
	Title    string
	SubTitle string
}
