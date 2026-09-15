package message

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type MessageService struct {
	repo repo.MessageRepository
}

func NewMessageService(repo repo.MessageRepository) MessageService {
	return MessageService{repo: repo}
}

func (s MessageService) List(ctx context.Context, userID int64, publicBaseURL string) ([]types.MessageListItem, error) {
	rows, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]types.MessageListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, types.MessageListItem{
			ID:          row.ID,
			Title:       row.Title,
			Content:     row.Content,
			TargetType:  row.TargetType,
			ContentType: row.ContentType,
			AlreadyRead: row.AlreadyRead,
			PageURL:     row.PageURL,
			MsgType:     messageTypeText(row.TargetType),
			SubTitle:    row.SubTitle,
			Time:        messageTimeForList(row.Date, row.CreatedAt),
			ImageURL:    buildImageURL(publicBaseURL, row.AvatarURL),
		})
	}
	return items, nil
}

func (s MessageService) Show(ctx context.Context, userID int64, messageID int64) (types.MessageDetailItem, error) {
	row, err := s.repo.FindByIDForUser(ctx, messageID, userID)
	if err != nil {
		return types.MessageDetailItem{}, err
	}
	if row.AlreadyRead == 0 {
		if err := s.repo.MarkAsRead(ctx, row.ID, userID); err != nil && !errors.Is(err, repo.ErrMessageNotFound) {
			return types.MessageDetailItem{}, err
		}
	}
	return types.MessageDetailItem{
		Title:       row.Title,
		Content:     row.Content,
		Time:        messageTimeForDetail(row.Date, row.CreatedAt),
		ContentType: row.ContentType,
		MsgType:     messageTypeText(row.TargetType),
	}, nil
}

func messageTypeText(targetType int) string {
	if targetType == 0 {
		return "系统消息"
	}
	return ""
}

func messageTimeForList(date *time.Time, createdAt time.Time) string {
	if date != nil {
		return date.Format("2006-01-02 15:04")
	}
	return createdAt.Format("2006-01-02 15:04")
}

func messageTimeForDetail(date *time.Time, createdAt time.Time) string {
	if date != nil {
		return date.Format("2006-01-02")
	}
	return createdAt.Format("2006-01-02")
}

func buildImageURL(publicBaseURL string, avatarURL string) string {
	avatarURL = strings.TrimSpace(avatarURL)
	if avatarURL == "" {
		return ""
	}
	if strings.HasPrefix(avatarURL, "http://") || strings.HasPrefix(avatarURL, "https://") {
		return avatarURL
	}
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	path := strings.TrimLeft(avatarURL, "/")
	if base == "" {
		return "/" + path
	}
	return base + "/" + path
}
