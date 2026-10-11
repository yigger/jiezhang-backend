package model

import (
	"encoding/json"
	"time"
)

// InsightStatementAnnotation maps additive analytics storage created by the 20261006 migration.
type InsightStatementAnnotation struct {
	ConsumerID    *int64          `gorm:"column:consumer_id;type:bigint"`
	ID            int64           `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	AccountBookID int64           `gorm:"column:account_book_id;type:bigint;not null;index:idx_insight_statement_annotations_book,priority:1"`
	CreatorID     int64           `gorm:"column:creator_id;type:bigint;not null"`
	StatementID   int64           `gorm:"column:statement_id;type:bigint;not null;uniqueIndex:idx_insight_annotation_statement"`
	ProjectID     *int64          `gorm:"column:project_id;type:bigint"`
	PayerID       *int64          `gorm:"column:payer_id;type:bigint"`
	Allocations   json.RawMessage `gorm:"column:allocations;type:json"`
	FixedCostID   *int64          `gorm:"column:fixed_cost_id;type:bigint"`
	CreatedAt     time.Time       `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;type:datetime;not null"`
}

func (InsightStatementAnnotation) TableName() string { return "insight_statement_annotations" }
