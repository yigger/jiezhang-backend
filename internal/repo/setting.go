package repo

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrSettingUserNotFound = errors.New("setting user not found")

type SettingRepository interface {
	CreateFeedback(ctx context.Context, input tablemodel.Feedback) error
}
