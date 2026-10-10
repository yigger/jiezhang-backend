package model

import (
	"encoding/json"
	"time"
)

// InsightProject maps additive analytics storage created by the 20261006 migration.
type InsightProject struct {
	Icon           *string         `gorm:"column:icon;type:varchar(64)"`
	Color          *string         `gorm:"column:color;type:varchar(7)"`
	ParticipantIDs json.RawMessage `gorm:"column:participant_ids;type:json"`
	StartDate      *time.Time      `gorm:"column:start_date;type:date"`
	EndDate        *time.Time      `gorm:"column:end_date;type:date"`
	ID             int64           `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	AccountBookID  int64           `gorm:"column:account_book_id;type:bigint;not null;index:idx_insight_projects_book,priority:1"`
	CreatorID      int64           `gorm:"column:creator_id;type:bigint;not null"`
	Name           string          `gorm:"column:name;type:varchar(100);not null"`
	BudgetCents    int64           `gorm:"column:budget_cents;type:bigint;not null"`
	Archived       bool            `gorm:"column:archived;type:tinyint(1);not null"`
	CreatedAt      time.Time       `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;type:datetime;not null"`
}

func (InsightProject) TableName() string { return "insight_projects" }
