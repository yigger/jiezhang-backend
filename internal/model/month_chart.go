package model

import (
	"encoding/json"
	"time"
)

// MonthChart maps the month_charts table.
type MonthChart struct {
	ID            int64           `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID        *int64          `gorm:"column:user_id;type:int"`
	Year          *int            `gorm:"column:year;type:int"`
	Month         *int            `gorm:"column:month;type:int"`
	Dashboard     json.RawMessage `gorm:"column:dashboard;type:json"`
	ExpendCompare json.RawMessage `gorm:"column:expend_compare;type:json"`
	DayAvg        json.RawMessage `gorm:"column:day_avg;type:json"`
	WeekAvg       json.RawMessage `gorm:"column:week_avg;type:json"`
	MonthSurplus  json.RawMessage `gorm:"column:month_surplus;type:json"`
	BudgetUsed    json.RawMessage `gorm:"column:budget_used;type:json"`
	AssetTotal    json.RawMessage `gorm:"column:asset_total;type:json"`
	MonthLast10   json.RawMessage `gorm:"column:month_last_10;type:json"`
	BeginText     *string         `gorm:"column:begin_text;type:text"`
	EndText       *string         `gorm:"column:end_text;type:text"`
	Cover         *string         `gorm:"column:cover;type:text"`
	CreatedAt     time.Time       `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;type:datetime;not null"`
}

func (MonthChart) TableName() string { return "month_charts" }
