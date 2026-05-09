package mysql

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"gorm.io/gorm"
)

type StatementModel struct {
	ID          int64   `gorm:"primaryKey;autoIncrement"`
	UserID      int64   `gorm:"column:user_id;not null;index"`
	Amount      float64 `gorm:"type:decimal(10,2);not null;default:0"`
	Category    string  `gorm:"type:varchar(100);not null;default:''"`
	Asset       string  `gorm:"type:varchar(100);not null;default:''"`
	Description string  `gorm:"type:text;not null;default:''"`
}

type StatementRepository struct {
	db *gorm.DB
}

func (StatementModel) TableName() string {
	return "statements"
}

func NewStatementRepository(db *gorm.DB) (*StatementRepository, error) {
	return &StatementRepository{db: db}, nil
}

func (r *StatementRepository) List(ctx context.Context, userId int64) ([]domain.Statement, error) {
	var statements []domain.Statement
	if err := r.db.WithContext(ctx).Where("user_id = ?", userId).Find(&statements).Error; err != nil {
		return nil, err
	}
	return statements, nil
}
