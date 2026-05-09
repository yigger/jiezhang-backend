package repository

import (
	"context"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
)

// StatementRepository is for command-side writes (create/update/delete) in future.
type StatementRepository interface {
	// Create(ctx context.Context, statement domain.Statement) (domain.Statement, error)
	// Save(ctx context.Context, statement domain.Statement) (domain.Statement, error)
}

// StatementQueryRepository is for read-side complex queries.
type StatementQueryRepository interface {
	ListWithRelations(ctx context.Context, filter StatementListFilter) ([]StatementListItem, error)
}

type StatementListFilter struct {
	UserID            int64
	AccountBookID     int64
	StartDate         *time.Time
	EndDate           *time.Time
	ParentCategoryIDs []int64
	ExceptIDs         []int64
	OrderBy           string
	Limit             int
	Offset            int
}

type StatementListItem struct {
	ID               int64     `json:"id"`
	Type             string    `json:"type"`
	Amount           float64   `json:"amount"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CategoryID       int64     `json:"category_id"`
	CategoryName     string    `json:"category_name"`
	ParentCategoryID int64     `json:"parent_category_id"`
	ParentCategory   string    `json:"parent_category"`
	AssetID          int64     `json:"asset_id"`
	AssetName        string    `json:"asset_name"`
	PayeeID          int64     `json:"payee_id"`
	PayeeName        string    `json:"payee_name"`
}

// Keep domain import alive for future command-side repository evolution.
var _ domain.Statement
