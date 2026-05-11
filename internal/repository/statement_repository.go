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
	ListRowsWithRelations(ctx context.Context, filter StatementListFilter) ([]StatementListRowRecord, error)
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
	Province        string
	City            string
	Street          string
	HasPic          bool
	PayeeID         int64
	PayeeName       string
	TargetAssetName string
}

// Keep domain import alive for future command-side repository evolution.
var _ domain.Statement
