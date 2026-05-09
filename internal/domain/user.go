package domain

import "time"

// User is the business entity and should stay independent from transport/ORM details.
type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
