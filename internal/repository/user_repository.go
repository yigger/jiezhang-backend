package repository

import (
	"context"
	"errors"

	"github.com/yigger/jiezhang-backend/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (domain.User, error)
	FindByOpenID(ctx context.Context, openID string) (domain.User, error)
	FindByThirdSession(ctx context.Context, thirdSession string) (domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
	Save(ctx context.Context, user domain.User) (domain.User, error)

	GetProfile(ctx context.Context, id int64) (UserProfileRecord, error)
	UpdateProfile(ctx context.Context, id int64, input UserProfileUpdateRecord) error
	SetBackgroundAvatarURL(ctx context.Context, id int64, avatarURL string) error
	MarkAlreadyLogin(ctx context.Context, id int64, alreadyLogin bool) error
}

type UserProfileRecord struct {
	ID               int64
	ThemeID          int64
	AvatarURL        string
	Nickname         string
	Email            string
	Remind           int
	HiddenAssetMoney bool
	Persist          int64
	StatementsCount  int64
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
