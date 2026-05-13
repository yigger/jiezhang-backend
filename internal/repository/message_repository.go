package repository

import (
	"context"
	"errors"
	"time"
)

var ErrMessageNotFound = errors.New("message not found")

type MessageRepository interface {
	ListByUserID(ctx context.Context, userID int64) ([]MessageRecord, error)
	FindByIDForUser(ctx context.Context, id int64, userID int64) (MessageRecord, error)
	MarkAsRead(ctx context.Context, id int64, userID int64) error
}

type MessageRecord struct {
	ID          int64
	Title       string
	Content     string
	TargetType  int
	ContentType string
	AlreadyRead int
	PageURL     string
	SubTitle    string
	AvatarURL   string
	Date        *time.Time
	CreatedAt   time.Time
}
