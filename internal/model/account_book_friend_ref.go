package model

import (
	"time"
)

// AccountBookFriendRef maps the account_book_friend_refs table.
type AccountBookFriendRef struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	AccountBookID *int64    `gorm:"column:account_book_id;type:bigint;index:index_account_book_friend_refs_on_account_book_id,priority:1"`
	FriendID      *int64    `gorm:"column:friend_id;type:bigint;index:index_account_book_friend_refs_on_friend_id,priority:1"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (AccountBookFriendRef) TableName() string { return "account_book_friend_refs" }
