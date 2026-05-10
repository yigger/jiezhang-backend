package domain

import (
	"strconv"
	"time"
)

// User is the business entity and should stay independent from transport/ORM details.
type User struct {
	ID            int64
	Name          string
	Email         string
	OpenID        string
	SessionKey    string
	ThirdSession  string
	AccountBookId int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (u User) RedisSessionKey() string {
	return "@user_" + strconv.FormatInt(u.ID, 10) + "_session_key@"
}
