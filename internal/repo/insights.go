package repo

import (
	"context"
	"github.com/yigger/jiezhang-backend/internal/model"
	"time"
)

// InsightStatementRecord is a joined projection over existing statement and lookup tables.
type InsightStatementRecord struct {
	model.Statement
	CategoryName string
	MerchantName string
	MemberName   string
}

type InsightsRepository interface {
	ListRows(ctx context.Context, accountBookID int64, start, end time.Time) ([]InsightStatementRecord, error)
}
