package mysql

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type UploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) (*UploadRepository, error) {
	return &UploadRepository{db: db}, nil
}

type userAssetModel struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name          string    `gorm:"column:name"`
	Path          string    `gorm:"column:path"`
	Type          string    `gorm:"column:type"`
	ImageableType string    `gorm:"column:imageable_type"`
	ImageableID   int64     `gorm:"column:imageable_id"`
	Score         int       `gorm:"column:score"`
	Locked        int       `gorm:"column:locked"`
	System        int       `gorm:"column:system"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (userAssetModel) TableName() string {
	return "user_assets"
}

func (r *UploadRepository) CreateUserAvatar(ctx context.Context, userID int64, path string) error {
	now := time.Now()
	model := userAssetModel{
		Name:          "user-avatar",
		Path:          path,
		Type:          "UserAvatar",
		ImageableType: "User",
		ImageableID:   userID,
		Score:         0,
		Locked:        0,
		System:        0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *UploadRepository) CreateStatementAvatar(ctx context.Context, accountBookID int64, statementID int64, path string) error {
	var count int64
	if err := r.db.WithContext(ctx).
		Table("statements").
		Where("id = ? AND account_book_id = ?", statementID, accountBookID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return repository.ErrStatementNotFound
	}

	now := time.Now()
	model := userAssetModel{
		Name:          "statement-avatar",
		Path:          path,
		Type:          "StatementAvatar",
		ImageableType: "Statement",
		ImageableID:   statementID,
		Score:         0,
		Locked:        0,
		System:        0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}
