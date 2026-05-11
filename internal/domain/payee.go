package domain

import "time"

type Payee struct {
	ID            int64
	Name          string
	UserID        int64
	AccountBookID int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
