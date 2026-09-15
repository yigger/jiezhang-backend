package repo

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (tablemodel.User, error)
	FindByOpenID(ctx context.Context, openID string) (tablemodel.User, error)
	FindByThirdSession(ctx context.Context, thirdSession string) (tablemodel.User, error)
	List(ctx context.Context) ([]tablemodel.User, error)
	Create(ctx context.Context, user tablemodel.User) (tablemodel.User, error)
	CreateWithInit(ctx context.Context, user tablemodel.User, bookInput AccountBookCreateInput) (tablemodel.User, error)
	Save(ctx context.Context, user tablemodel.User) (tablemodel.User, error)

	StatementCounts(ctx context.Context, userID, accountBookID int64) (StatementCounts, error)
	UpdateProfile(ctx context.Context, id int64, input UserProfileUpdateRecord) error
	SetBackgroundAvatarURL(ctx context.Context, id int64, avatarURL string) error
	MarkAlreadyLogin(ctx context.Context, id int64, alreadyLogin bool) error
}

type StatementCounts struct {
	Persist         int64
	StatementsCount int64
}

type UserProfileUpdateRecord struct {
	ThemeID          *int64
	Country          *string
	City             *string
	Gender           *int
	Language         *string
	Province         *string
	BGAvatarID       *int64
	HiddenAssetMoney *bool
	AvatarURL        *string
	Nickname         *string
}
