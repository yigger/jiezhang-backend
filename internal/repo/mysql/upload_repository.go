package mysql

import (
	"context"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type UploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepository { return &UploadRepository{db: db} }

func (r *UploadRepository) CreateUserAvatar(ctx context.Context, userID int64, path string) error {
	now := time.Now()
	model := tablemodel.UserAsset{
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
		return repo.ErrStatementNotFound
	}

	now := time.Now()
	model := tablemodel.UserAsset{
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
