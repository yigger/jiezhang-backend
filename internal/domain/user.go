package domain

import (
	"strconv"
	"time"
)

// User is the business entity and should stay independent from transport/ORM details.
type User struct {
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
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (u User) RedisSessionKey() string {
	return "@user_" + strconv.FormatInt(u.ID, 10) + "_session_key@"
}

type Theme struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ClassName string `json:"class_name"`
}

var DefaultThemes = []Theme{
	{ID: 0, Name: "默认色", ClassName: "jz-theme-default"},
	{ID: 1, Name: "深蓝色", ClassName: "jz-theme-black"},
	{ID: 2, Name: "橄榄绿", ClassName: "jz-theme-green"},
	{ID: 3, Name: "少女粉", ClassName: "jz-theme-pink"},
	{ID: 4, Name: "活力橙", ClassName: "jz-theme-orange"},
	{ID: 5, Name: "动感黄", ClassName: "jz-theme-yellow"},
	{ID: 6, Name: "罗兰紫", ClassName: "jz-theme-purple"},
}
