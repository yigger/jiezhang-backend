package repository

import (
	"context"
	"errors"
	"time"
)

var ErrSettingUserNotFound = errors.New("setting user not found")

type SettingRepository interface {
	CreateFeedback(ctx context.Context, input SettingFeedbackCreateRecord) error
}

type SettingFeedbackCreateRecord struct {
	UserID    int64
	Content   string
	Type      int
	CreatedAt time.Time
	UpdatedAt time.Time
}
