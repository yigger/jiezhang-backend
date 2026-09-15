package mysql

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) FindByID(ctx context.Context, id int64) (tablemodel.User, error) {
	var model tablemodel.User
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.User{}, repo.ErrUserNotFound
		}
		return tablemodel.User{}, err
	}

	return model, nil
}

func (r *UserRepository) FindByOpenID(ctx context.Context, openID string) (tablemodel.User, error) {
	var model tablemodel.User
	if err := r.db.WithContext(ctx).Where("openid = ?", openID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.User{}, repo.ErrUserNotFound
		}
		return tablemodel.User{}, err
	}
	return model, nil
}

func (r *UserRepository) FindByThirdSession(ctx context.Context, thirdSession string) (tablemodel.User, error) {
	var model tablemodel.User
	if err := r.db.WithContext(ctx).Where("third_session = ?", thirdSession).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.User{}, repo.ErrUserNotFound
		}
		return tablemodel.User{}, err
	}
	return model, nil
}

func (r *UserRepository) List(ctx context.Context) ([]tablemodel.User, error) {
	var models []tablemodel.User
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
}

func (r *UserRepository) Create(ctx context.Context, user tablemodel.User) (tablemodel.User, error) {
	model := user
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return tablemodel.User{}, err
	}

	return model, nil
}

func (r *UserRepository) CreateWithInit(ctx context.Context, user tablemodel.User, bookInput repo.AccountBookCreateInput) (tablemodel.User, error) {
	var result tablemodel.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := user
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		bookInput.UserID = model.ID
		if user.Nickname != nil {
			bookInput.UserNickname = *user.Nickname
		}
		created, err := createAccountBookInTx(tx, bookInput)
		if err != nil {
			return err
		}
		model.UID = model.ID + 10000
		model.AccountBookID = created.ID
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		result = model
		return nil
	})
	if err != nil {
		return tablemodel.User{}, err
	}
	return result, nil
}

func (r *UserRepository) Save(ctx context.Context, user tablemodel.User) (tablemodel.User, error) {
	model := user
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return tablemodel.User{}, err
	}

	return model, nil
}

func (r *UserRepository) StatementCounts(ctx context.Context, userID, accountBookID int64) (repo.StatementCounts, error) {
	var counts repo.StatementCounts
	err := r.db.WithContext(ctx).Table("statements").Select("COUNT(DISTINCT CONCAT(year, '-', month, '-', day)) AS persist, COUNT(1) AS statements_count").
		Where("user_id = ? AND account_book_id = ?", userID, accountBookID).Scan(&counts).Error
	return counts, err
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, input repo.UserProfileUpdateRecord) error {
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
		return repo.ErrUserNotFound
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
		return repo.ErrUserNotFound
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
		return repo.ErrUserNotFound
	}
	return nil
}
