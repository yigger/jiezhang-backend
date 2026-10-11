package model

import (
	"encoding/json"
	"time"
)

// InsightPortfolioSnapshot maps additive analytics storage created by the 20261006 migration.
type InsightPortfolioSnapshot struct {
	ID               int64           `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	AccountBookID    int64           `gorm:"column:account_book_id;type:bigint;not null;index:idx_insight_portfolio_snapshots_book,priority:1"`
	CreatorID        int64           `gorm:"column:creator_id;type:bigint;not null"`
	AssetsCents      int64           `gorm:"column:assets_cents;type:bigint;not null"`
	LiabilitiesCents int64           `gorm:"column:liabilities_cents;type:bigint;not null"`
	Balances         json.RawMessage `gorm:"column:balances;type:json;not null"`
	Note             string          `gorm:"column:note;type:varchar(255);not null"`
	CreatedAt        time.Time       `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;type:datetime;not null"`
}

func (InsightPortfolioSnapshot) TableName() string { return "insight_portfolio_snapshots" }
