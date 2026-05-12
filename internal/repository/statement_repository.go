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
	StartDate         *time.Time
	EndDate           *time.Time
	ParentCategoryIDs []int64
	ExceptIDs         []int64
	OrderBy           string
	Limit             int
	Offset            int
}

// :id, :type, :description, :title, :amount, :target_object, :mood

type Payee struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementBaseItem struct {
	ID           int64   `json:"id"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	Title        string  `json:"title"`
	TargetObject string  `json:"target_object"`
	Mood         string  `json:"mood"`
	Money        string  `json:"money"`
	Category     string  `json:"category"`
	IconPath     string  `json:"icon_path"`
	Asset        string  `json:"asset"`
	Date         string  `json:"date"`
	Time         string  `json:"time"`
	TimeStr      string  `json:"timeStr"`
	Week         string  `json:"week"`
	Payee        Payee   `json:"payee"`
	Remark       string  `json:"remark"`
	CategoryID   int64   `json:"category_id"`
	AssetID      int64   `json:"asset_id"`
}

type StatementListItem struct {
	StatementBaseItem
	Location  string    `json:"location"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	Street    string    `json:"street"`
	MonthDay  string    `json:"month_day"`
	HasPic    bool      `json:"has_pic"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StatementDetailItem struct {
	StatementBaseItem
	Location    string        `json:"location"`
	Province    string        `json:"province"`
	City        string        `json:"city"`
	Street      string        `json:"street"`
	MonthDay    string        `json:"month_day"`
	HasPic      bool          `json:"has_pic"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	UploadFIles []interface{} `json:"upload_files"`
}

type StatementListRowRecord struct {
	ID              int64
	Type            string
	Amount          float64
	Description     string
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
	CategoryID      int64
	AssetID         int64
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
