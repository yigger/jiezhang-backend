package types

import (
	"strconv"
	"time"
)

// UserContext is the normalized authentication context assembled by services.
// The users table is mapped only by model.User.
type UserContext struct {
	ID            int64
	UID           int64
	Nickname      string
	AvatarUrl     string
	ThemeID       int64
	Email         string
	OpenID        string
	SessionKey    string
	ThirdSession  string
	AccountBookId int64

	Country          string
	City             string
	Gender           int
	Language         string
	Province         string
	BGAvatarID       int64
	BGAvatarURL      string
	Remind           int
	HiddenAssetMoney bool
	AlreadyLogin     bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u UserContext) RedisSessionKey() string {
	return "@go:user_" + strconv.FormatInt(u.ID, 10) + "_session_key@"
}
