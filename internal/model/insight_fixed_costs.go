package model

import (
	"time"
)

// InsightFixedCost maps additive analytics storage created by the 20261006 migration.
type InsightFixedCost struct {
	NextRunDate    *time.Time `gorm:"column:next_run_date;type:date"`
	ID             int64      `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	AccountBookID  int64      `gorm:"column:account_book_id;type:bigint;not null;index:idx_insight_fixed_costs_book,priority:1"`
	CreatorID      int64      `gorm:"column:creator_id;type:bigint;not null"`
	Name           string     `gorm:"column:name;type:varchar(100);not null"`
	AmountCents    int64      `gorm:"column:amount_cents;type:bigint;not null"`
	CategoryID     int64      `gorm:"column:category_id;type:bigint;not null"`
	AssetID        int64      `gorm:"column:asset_id;type:bigint;not null"`
	IntervalMonths int        `gorm:"column:interval_months;type:int;not null"`
	DueDay         int        `gorm:"column:due_day;type:int;not null"`
	CandidateKey   string     `gorm:"column:candidate_key;type:varchar(255);not null"`
	Active         bool       `gorm:"column:active;type:tinyint(1);not null"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:datetime;not null"`
}

func (InsightFixedCost) TableName() string { return "insight_fixed_costs" }
