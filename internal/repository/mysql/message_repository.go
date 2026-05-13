package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) (*MessageRepository, error) {
	return &MessageRepository{db: db}, nil
}

type messageRow struct {
	ID          int64      `gorm:"column:id"`
	Title       string     `gorm:"column:title"`
	Content     string     `gorm:"column:content"`
	TargetType  int        `gorm:"column:target_type"`
	ContentType string     `gorm:"column:content_type"`
	AlreadyRead int        `gorm:"column:already_read"`
	PageURL     string     `gorm:"column:page_url"`
	SubTitle    string     `gorm:"column:sub_title"`
	AvatarURL   string     `gorm:"column:avatar_url"`
	Date        *time.Time `gorm:"column:date"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (r *MessageRepository) ListByUserID(ctx context.Context, userID int64) ([]repository.MessageRecord, error) {
	rows := make([]messageRow, 0)
	err := r.db.WithContext(ctx).
		Table("messages").
		Select([]string{
			"id",
			"COALESCE(title, '') AS title",
			"COALESCE(content, '') AS content",
			"COALESCE(target_type, 0) AS target_type",
			"COALESCE(content_type, '') AS content_type",
			"COALESCE(already_read, 0) AS already_read",
			"COALESCE(page_url, '') AS page_url",
			"COALESCE(sub_title, '') AS sub_title",
			"COALESCE(avatar_url, '') AS avatar_url",
			"date",
			"created_at",
		}).
		Where("target_id = ?", userID).
		Order("created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.MessageRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.MessageRecord{
			ID:          row.ID,
			Title:       row.Title,
			Content:     row.Content,
			TargetType:  row.TargetType,
			ContentType: row.ContentType,
			AlreadyRead: row.AlreadyRead,
			PageURL:     row.PageURL,
			SubTitle:    row.SubTitle,
			AvatarURL:   row.AvatarURL,
			Date:        row.Date,
			CreatedAt:   row.CreatedAt,
		})
	}
	return items, nil
}

func (r *MessageRepository) FindByIDForUser(ctx context.Context, id int64, userID int64) (repository.MessageRecord, error) {
	var row messageRow
	err := r.db.WithContext(ctx).
		Table("messages").
		Select([]string{
			"id",
			"COALESCE(title, '') AS title",
			"COALESCE(content, '') AS content",
			"COALESCE(target_type, 0) AS target_type",
			"COALESCE(content_type, '') AS content_type",
			"COALESCE(already_read, 0) AS already_read",
			"COALESCE(page_url, '') AS page_url",
			"COALESCE(sub_title, '') AS sub_title",
			"COALESCE(avatar_url, '') AS avatar_url",
			"date",
			"created_at",
		}).
		Where("id = ? AND target_id = ?", id, userID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.MessageRecord{}, repository.ErrMessageNotFound
		}
		return repository.MessageRecord{}, err
	}
	return repository.MessageRecord{
		ID:          row.ID,
		Title:       row.Title,
		Content:     row.Content,
		TargetType:  row.TargetType,
		ContentType: row.ContentType,
		AlreadyRead: row.AlreadyRead,
		PageURL:     row.PageURL,
		SubTitle:    row.SubTitle,
		AvatarURL:   row.AvatarURL,
		Date:        row.Date,
		CreatedAt:   row.CreatedAt,
	}, nil
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, id int64, userID int64) error {
	res := r.db.WithContext(ctx).
		Table("messages").
		Where("id = ? AND target_id = ?", id, userID).
		Updates(map[string]interface{}{
			"already_read": 1,
			"updated_at":   time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrMessageNotFound
	}
	return nil
}
