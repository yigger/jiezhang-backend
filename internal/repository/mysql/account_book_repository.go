package mysql

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"gorm.io/gorm"
)

type AccountBookRepository struct {
	db *gorm.DB
}

func NewAccountBookRepository(db *gorm.DB) (*AccountBookRepository, error) {
	return &AccountBookRepository{db: db}, nil
}

func (r *AccountBookRepository) FindByID(ctx context.Context, id int64, userId int64) (domain.AccountBook, error) {
	var accountBook domain.AccountBook
	err := r.db.WithContext(ctx).Table("account_books").Where("id = ? and user_id = ?", id, userId).First(&accountBook).Error
	return accountBook, err
}
