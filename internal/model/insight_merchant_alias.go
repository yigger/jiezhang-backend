package model

import "time"

// InsightMerchantAlias groups merchants for analytics without rewriting payees or statements.
type InsightMerchantAlias struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	AccountBookID int64     `gorm:"column:account_book_id;type:bigint;not null;index:idx_insight_merchant_aliases_book"`
	CreatorID     int64     `gorm:"column:creator_id;type:bigint;not null"`
	PayeeID       int64     `gorm:"column:payee_id;type:bigint;not null;uniqueIndex:idx_insight_merchant_aliases_payee"`
	Name          string    `gorm:"column:name;type:varchar(100);not null"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (InsightMerchantAlias) TableName() string { return "insight_merchant_aliases" }
