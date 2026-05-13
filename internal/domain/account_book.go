package domain

import "time"

type AccountBook struct {
	ID          int64
	UserID      int64
	AccountType int
	Name        string
	Description string
	Budget      float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
