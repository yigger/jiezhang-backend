package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

type userModel struct {
	ID            int64     `gorm:"primaryKey;autoIncrement"`
	Name          string    `gorm:"type:varchar(100);not null;default:''"`
	Email         string    `gorm:"type:varchar(255);default:''"`
	OpenID        string    `gorm:"column:openid;type:varchar(255);not null;uniqueIndex"`
	SessionKey    string    `gorm:"column:session_key;type:text;not null;default:''"`
	ThirdSession  string    `gorm:"column:third_session;type:varchar(255);not null;default:'';index"`
	AccountBookId int64     `gorm:"column:account_book_id;not null;default:0"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (userModel) TableName() string {
	return "users"
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) (*UserRepository, error) {
	return &UserRepository{db: db}, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	var model userModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return toDomain(model), nil
}

func (r *UserRepository) FindByOpenID(ctx context.Context, openID string) (domain.User, error) {
	var model userModel
	if err := r.db.WithContext(ctx).Where("openid = ?", openID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomain(model), nil
}

func (r *UserRepository) FindByThirdSession(ctx context.Context, thirdSession string) (domain.User, error) {
	var model userModel
	if err := r.db.WithContext(ctx).Where("third_session = ?", thirdSession).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomain(model), nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	var models []userModel
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(models))
	for _, model := range models {
		users = append(users, toDomain(model))
	}

	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	model := fromDomain(user)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.User{}, err
	}

	return toDomain(model), nil
}

func (r *UserRepository) Save(ctx context.Context, user domain.User) (domain.User, error) {
	model := fromDomain(user)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.User{}, err
	}

	return toDomain(model), nil
}

func toDomain(model userModel) domain.User {
	return domain.User{
		ID:            model.ID,
		Name:          model.Name,
		Email:         model.Email,
		OpenID:        model.OpenID,
		SessionKey:    model.SessionKey,
		ThirdSession:  model.ThirdSession,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		AccountBookId: model.AccountBookId,
	}
}

func fromDomain(user domain.User) userModel {
	return userModel{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		OpenID:        user.OpenID,
		SessionKey:    user.SessionKey,
		ThirdSession:  user.ThirdSession,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		AccountBookId: user.AccountBookId,
	}
}
