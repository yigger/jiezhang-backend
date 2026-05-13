package repository

import (
	"context"
	"errors"
	"time"
)

var ErrStatementNotFound = errors.New("statement not found")

// StatementRepository is for command-side writes (create/update/delete).
type StatementRepository interface {
	Create(ctx context.Context, input StatementWriteRecord) (int64, error)
	GetOwnerID(ctx context.Context, statementID int64, accountBookID int64) (int64, error)
	UpdateByID(ctx context.Context, statementID int64, accountBookID int64, input StatementWriteRecord) error
	DeleteByID(ctx context.Context, statementID int64, accountBookID int64) error
}

// StatementQueryRepository is for read-side complex queries.
type StatementQueryRepository interface {
	ListRowsWithRelations(ctx context.Context, filter StatementListFilter) ([]StatementListRowRecord, error)
	GetRowByIDWithRelations(ctx context.Context, statementID int64, accountBookID int64) (StatementListRowRecord, error)
	GetLatestCategoryAssetByType(ctx context.Context, accountBookID int64, statementType string) (*StatementDefaultCategoryAssetRecord, error)
	ListDistinctTargetObjectsByType(ctx context.Context, accountBookID int64, statementType string) ([]string, error)
}

type StatementListFilter struct {
	UserID            int64
	AccountBookID     int64
	AssetID           int64
	Type              string
	StartDate         *time.Time
	EndDate           *time.Time
	ParentCategoryIDs []int64
	ExceptIDs         []int64
	OrderBy           string
	Limit             int
	Offset            int
}

type StatementRowRecord struct {
	ID           int64
	Type         string
	Amount       float64
	CategoryID   int64
	AssetID      int64
	Description  string
	Remark       string
	Mood         string
	IconPath     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CategoryName string
	AssetName    string
	Location     string
	Nation       string
	Province     string
	City         string
	District     string
	Street       string
}

type StatementListRowRecord struct {
	ID              int64
	Type            string
	Amount          float64
	Description     string
	CategoryID      int64
	AssetID         int64
	Remark          string
	Mood            string
	IconPath        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CategoryName    string
	AssetName       string
	Location        string
	Nation          string
	Province        string
	City            string
	District        string
	Street          string
	HasPic          bool
	PayeeID         int64
	PayeeName       string
	TargetAssetID   int64
	TargetAssetName string
	TargetObject    string
}

type StatementWriteRecord struct {
	UserID        int64
	AccountBookID int64
	Type          string
	Amount        float64
	Description   string
	Mood          string
	CategoryID    int64
	AssetID       int64
	TargetAssetID *int64
	PayeeID       *int64
	TargetObject  string
	Location      string
	Nation        string
	Province      string
	City          string
	District      string
	Street        string
	OccurredAt    time.Time
}

type StatementDefaultCategoryAssetRecord struct {
	CategoryID   int64
	AssetID      int64
	CategoryName string
	AssetName    string
}

// Keep domain import alive for future command-side repository evolution.
// var _ domain.Statement
