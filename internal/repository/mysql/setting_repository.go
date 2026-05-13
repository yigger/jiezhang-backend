package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) (*SettingRepository, error) {
	return &SettingRepository{db: db}, nil
}

func (r *SettingRepository) CreateFeedback(ctx context.Context, input repository.SettingFeedbackCreateRecord) error {
	row := map[string]interface{}{
		"user_id":    input.UserID,
		"content":    input.Content,
		"type":       input.Type,
		"created_at": input.CreatedAt,
		"updated_at": input.UpdatedAt,
	}
	return r.db.WithContext(ctx).Table("feedbacks").Create(row).Error
}
