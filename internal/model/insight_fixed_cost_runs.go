package model

import "time"

// InsightFixedCostRun preserves occurrence identity even when its statement is deleted.
type InsightFixedCostRun struct {
	ID          int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	FixedCostID int64     `gorm:"column:fixed_cost_id;type:bigint;not null;uniqueIndex:idx_insight_fixed_cost_run,priority:1"`
	DueDate     time.Time `gorm:"column:due_date;type:date;not null;uniqueIndex:idx_insight_fixed_cost_run,priority:2"`
	StatementID int64     `gorm:"column:statement_id;type:bigint;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;not null"`
}

func (InsightFixedCostRun) TableName() string { return "insight_fixed_cost_runs" }
