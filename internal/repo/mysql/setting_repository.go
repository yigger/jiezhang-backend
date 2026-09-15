package mysql

import (
	"context"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"gorm.io/gorm"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) *SettingRepository { return &SettingRepository{db: db} }

func (r *SettingRepository) CreateFeedback(ctx context.Context, input tablemodel.Feedback) error {
	return r.db.WithContext(ctx).Create(&input).Error
}
