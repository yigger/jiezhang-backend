package model

import (
	"strconv"
	"time"
)

// User maps the users table.
type User struct {
	ID               int64     `gorm:"column:id;type:int;not null;primaryKey;autoIncrement"`
	Email            *string   `gorm:"column:email;type:varchar(255);default:''"`
	ThemeID          int64     `gorm:"column:theme_id;type:int;default:0"`
	OpenID           string    `gorm:"column:openid;type:varchar(255);not null;uniqueIndex:idx_users_open_id,priority:1"`
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime(3);not null"`
	UpdatedAt        time.Time `gorm:"column:updated_at;type:datetime(3);not null"`
	Nickname         *string   `gorm:"column:nickname;type:varchar(255)"`
	Language         *string   `gorm:"column:language;type:varchar(255)"`
	City             *string   `gorm:"column:city;type:varchar(255)"`
	Province         *string   `gorm:"column:province;type:varchar(255)"`
	AvatarURL        *string   `gorm:"column:avatar_url;type:varchar(512)"`
	Country          *string   `gorm:"column:country;type:varchar(255)"`
	SessionKey       *string   `gorm:"column:session_key;type:varchar(255)"`
	Gender           int       `gorm:"column:gender;type:int"`
	UID              int64     `gorm:"column:uid;type:int;not null;default:0"`
	ThirdSession     *string   `gorm:"column:third_session;type:varchar(255)"`
	Phone            *string   `gorm:"column:phone;type:varchar(255)"`
	Budget           *float64  `gorm:"column:budget;type:decimal(12,2);default:0.00"`
	BGAvatarURL      *string   `gorm:"column:bg_avatar_url;type:varchar(255)"`
	BonusPoints      *int      `gorm:"column:bonus_points;type:int;default:0"`
	HeaderPosition1  *string   `gorm:"column:header_position_1;type:varchar(255)"`
	HeaderPosition2  *string   `gorm:"column:header_position_2;type:varchar(255)"`
	HeaderPosition3  *string   `gorm:"column:header_position_3;type:varchar(255)"`
	BGAvatarID       int64     `gorm:"column:bg_avatar_id;type:int"`
	Remind           int       `gorm:"column:remind;type:int;default:0"`
	HiddenAssetMoney bool      `gorm:"column:hidden_asset_money;type:tinyint(1);default:0"`
	AlreadyLogin     bool      `gorm:"column:already_login;type:tinyint(1);default:0"`
	Admin            *bool     `gorm:"column:admin;type:tinyint(1);default:0"`
	Password         *string   `gorm:"column:password;type:varchar(255)" json:"-"`
	Salt             *string   `gorm:"column:salt;type:varchar(255)" json:"-"`
	AccountBookID    int64     `gorm:"column:account_book_id;type:int"`
	Name             string    `gorm:"column:name;type:varchar(100);not null;default:''"`
}

func (User) TableName() string { return "users" }

func (u User) RedisSessionKey() string {
	return "@go:user_" + strconv.FormatInt(u.ID, 10) + "_session_key@"
}
