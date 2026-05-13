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
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	UID              int64     `gorm:"column:uid;not null;default:0;uniqueIndex"`
	Nickname         string    `gorm:"column:nickname;type:varchar(100);not null;default:''"`
	Email            string    `gorm:"column:email;type:varchar(255);default:''"`
	ThemeID          int64     `gorm:"column:theme_id;not null;default:0"`
	OpenID           string    `gorm:"column:openid;type:varchar(255);not null;uniqueIndex"`
	SessionKey       string    `gorm:"column:session_key;type:text;not null;default:''"`
	ThirdSession     string    `gorm:"column:third_session;type:varchar(255);not null;default:'';index"`
	AccountBookID    int64     `gorm:"column:account_book_id;not null;default:0"`
	AvatarURL        string    `gorm:"column:avatar_url;type:varchar(512);default:''"`
	Country          string    `gorm:"column:country;type:varchar(255);default:''"`
	City             string    `gorm:"column:city;type:varchar(255);default:''"`
	Gender           int       `gorm:"column:gender;default:0"`
	Language         string    `gorm:"column:language;type:varchar(255);default:''"`
	Province         string    `gorm:"column:province;type:varchar(255);default:''"`
	BGAvatarID       int64     `gorm:"column:bg_avatar_id;default:0"`
	BGAvatarURL      string    `gorm:"column:bg_avatar_url;type:varchar(512);default:''"`
	Remind           int       `gorm:"column:remind;default:0"`
	HiddenAssetMoney bool      `gorm:"column:hidden_asset_money;default:false"`
	AlreadyLogin     bool      `gorm:"column:already_login;default:false"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
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

func (r *UserRepository) GetProfile(ctx context.Context, id int64) (repository.UserProfileRecord, error) {
	var row struct {
		ID               int64  `gorm:"column:id"`
		ThemeID          int64  `gorm:"column:theme_id"`
		AvatarURL        string `gorm:"column:avatar_url"`
		Nickname         string `gorm:"column:nickname"`
		Email            string `gorm:"column:email"`
		Remind           int    `gorm:"column:remind"`
		HiddenAssetMoney bool   `gorm:"column:hidden_asset_money"`
		Persist          int64  `gorm:"column:persist"`
		StatementsCount  int64  `gorm:"column:statements_count"`
	}

	err := r.db.WithContext(ctx).
		Table("users u").
		Select([]string{
			"u.id",
			"u.theme_id",
			"COALESCE(u.avatar_url, '') AS avatar_url",
			"COALESCE(u.nickname, '') AS nickname",
			"COALESCE(u.email, '') AS email",
			"COALESCE(u.remind, 0) AS remind",
			"COALESCE(u.hidden_asset_money, 0) AS hidden_asset_money",
			"(SELECT COUNT(DISTINCT CONCAT(s.year, '-', s.month, '-', s.day)) FROM statements s WHERE s.user_id = u.id AND s.account_book_id = u.account_book_id) AS persist",
			"(SELECT COUNT(1) FROM statements s2 WHERE s2.user_id = u.id AND s2.account_book_id = u.account_book_id) AS statements_count",
		}).
		Where("u.id = ?", id).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.UserProfileRecord{}, repository.ErrUserNotFound
		}
		return repository.UserProfileRecord{}, err
	}

	return repository.UserProfileRecord{
		ID:               row.ID,
		ThemeID:          row.ThemeID,
		AvatarURL:        row.AvatarURL,
		Nickname:         row.Nickname,
		Email:            row.Email,
		Remind:           row.Remind,
		HiddenAssetMoney: row.HiddenAssetMoney,
		Persist:          row.Persist,
		StatementsCount:  row.StatementsCount,
	}, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, input repository.UserProfileUpdateRecord) error {
	updates := map[string]interface{}{"updated_at": time.Now()}
	if input.ThemeID != nil {
		updates["theme_id"] = *input.ThemeID
	}
	if input.Country != nil {
		updates["country"] = *input.Country
	}
	if input.City != nil {
		updates["city"] = *input.City
	}
	if input.Gender != nil {
		updates["gender"] = *input.Gender
	}
	if input.Language != nil {
		updates["language"] = *input.Language
	}
	if input.Province != nil {
		updates["province"] = *input.Province
	}
	if input.BGAvatarID != nil {
		updates["bg_avatar_id"] = *input.BGAvatarID
	}
	if input.HiddenAssetMoney != nil {
		updates["hidden_asset_money"] = *input.HiddenAssetMoney
	}
	if input.AvatarURL != nil {
		updates["avatar_url"] = *input.AvatarURL
	}
	if input.Nickname != nil {
		updates["nickname"] = *input.Nickname
	}

	res := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) SetBackgroundAvatarURL(ctx context.Context, id int64, avatarURL string) error {
	res := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Updates(map[string]interface{}{
		"bg_avatar_url": avatarURL,
		"updated_at":    time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) MarkAlreadyLogin(ctx context.Context, id int64, alreadyLogin bool) error {
	res := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Updates(map[string]interface{}{
		"already_login": alreadyLogin,
		"updated_at":    time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func toDomain(model userModel) domain.User {
	return domain.User{
		ID:               model.ID,
		UID:              model.UID,
		Nickname:         model.Nickname,
		Email:            model.Email,
		OpenID:           model.OpenID,
		SessionKey:       model.SessionKey,
		ThirdSession:     model.ThirdSession,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		AccountBookId:    model.AccountBookID,
		ThemeID:          model.ThemeID,
		AvatarUrl:        model.AvatarURL,
		Country:          model.Country,
		City:             model.City,
		Gender:           model.Gender,
		Language:         model.Language,
		Province:         model.Province,
		BGAvatarID:       model.BGAvatarID,
		BGAvatarURL:      model.BGAvatarURL,
		Remind:           model.Remind,
		HiddenAssetMoney: model.HiddenAssetMoney,
		AlreadyLogin:     model.AlreadyLogin,
	}
}

func fromDomain(user domain.User) userModel {
	return userModel{
		ID:               user.ID,
		Nickname:         user.Nickname,
		Email:            user.Email,
		OpenID:           user.OpenID,
		SessionKey:       user.SessionKey,
		ThirdSession:     user.ThirdSession,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		AccountBookID:    user.AccountBookId,
		ThemeID:          user.ThemeID,
		AvatarURL:        user.AvatarUrl,
		Country:          user.Country,
		City:             user.City,
		Gender:           user.Gender,
		Language:         user.Language,
		Province:         user.Province,
		BGAvatarID:       user.BGAvatarID,
		BGAvatarURL:      user.BGAvatarURL,
		Remind:           user.Remind,
		HiddenAssetMoney: user.HiddenAssetMoney,
		AlreadyLogin:     user.AlreadyLogin,
	}
}
