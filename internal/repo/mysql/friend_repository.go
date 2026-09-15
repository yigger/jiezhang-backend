package mysql

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type FriendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) *FriendRepository { return &FriendRepository{db: db} }

func (r *FriendRepository) ListCollaborators(ctx context.Context, accountBookID int64) ([]tablemodel.AccountBookCollaborator, error) {
	rows := make([]tablemodel.AccountBookCollaborator, 0)
	err := r.db.WithContext(ctx).
		Table("account_book_collaborators abc").
		Joins("INNER JOIN users u ON u.id = abc.user_id").
		Select([]string{
			"abc.id",
			"abc.account_book_id",
			"abc.user_id",
			"abc.role",
			"COALESCE(abc.remark, '') AS remark",
			"abc.created_at",
			"abc.updated_at",
		}).
		Where("abc.account_book_id = ?", accountBookID).
		Order("abc.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *FriendRepository) FindCollaboratorByUserID(ctx context.Context, accountBookID int64, userID int64) (tablemodel.AccountBookCollaborator, error) {
	var row tablemodel.AccountBookCollaborator
	err := r.db.WithContext(ctx).
		Table("account_book_collaborators abc").
		Joins("INNER JOIN users u ON u.id = abc.user_id").
		Select([]string{
			"abc.id",
			"abc.account_book_id",
			"abc.user_id",
			"abc.role",
			"COALESCE(abc.remark, '') AS remark",
			"abc.created_at",
			"abc.updated_at",
		}).
		Where("abc.account_book_id = ? AND abc.user_id = ?", accountBookID, userID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.AccountBookCollaborator{}, repo.ErrFriendCollaboratorNotFound
		}
		return tablemodel.AccountBookCollaborator{}, err
	}
	return row, nil
}

func (r *FriendRepository) FindCollaboratorByID(ctx context.Context, accountBookID int64, collaboratorID int64) (tablemodel.AccountBookCollaborator, error) {
	var row tablemodel.AccountBookCollaborator
	err := r.db.WithContext(ctx).
		Table("account_book_collaborators abc").
		Joins("INNER JOIN users u ON u.id = abc.user_id").
		Select([]string{
			"abc.id",
			"abc.account_book_id",
			"abc.user_id",
			"abc.role",
			"COALESCE(abc.remark, '') AS remark",
			"abc.created_at",
			"abc.updated_at",
		}).
		Where("abc.id = ? AND abc.account_book_id = ?", collaboratorID, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.AccountBookCollaborator{}, repo.ErrFriendCollaboratorNotFound
		}
		return tablemodel.AccountBookCollaborator{}, err
	}
	return row, nil
}

func (r *FriendRepository) FindUserByID(ctx context.Context, userID int64) (tablemodel.User, error) {
	var row tablemodel.User
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, COALESCE(nickname, '') AS nickname, COALESCE(avatar_url, '') AS avatar_url").
		Where("id = ?", userID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.User{}, repo.ErrUserNotFound
		}
		return tablemodel.User{}, err
	}
	return row, nil
}

func (r *FriendRepository) FindAccountBookByID(ctx context.Context, accountBookID int64) (tablemodel.AccountBook, error) {
	var row tablemodel.AccountBook
	err := r.db.WithContext(ctx).
		Table("account_books").
		Select("id, user_id, COALESCE(name, '') AS name").
		Where("id = ?", accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.AccountBook{}, repo.ErrFriendAccountBookNotFound
		}
		return tablemodel.AccountBook{}, err
	}
	return row, nil
}

func (r *FriendRepository) FindAccessibleAccountBookByID(ctx context.Context, userID int64, accountBookID int64) (tablemodel.AccountBook, error) {
	var row tablemodel.AccountBook
	err := r.db.WithContext(ctx).
		Table("account_books ab").
		Joins("LEFT JOIN account_book_collaborators abc ON abc.account_book_id = ab.id").
		Select("ab.id, ab.user_id, COALESCE(ab.name, '') AS name").
		Where("ab.id = ? AND (ab.user_id = ? OR abc.user_id = ?)", accountBookID, userID, userID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.AccountBook{}, repo.ErrFriendAccountBookNotFound
		}
		return tablemodel.AccountBook{}, err
	}
	return row, nil
}

func (r *FriendRepository) FindFirstOwnedAccountBookID(ctx context.Context, userID int64) (*int64, error) {
	var row struct {
		ID int64 `gorm:"column:id"`
	}
	err := r.db.WithContext(ctx).
		Table("account_books").
		Select("id").
		Where("user_id = ?", userID).
		Order("id ASC").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	id := row.ID
	return &id, nil
}

func (r *FriendRepository) UpdateUserDefaultAccountBook(ctx context.Context, userID int64, accountBookID *int64) error {
	var value interface{}
	if accountBookID == nil || *accountBookID <= 0 {
		value = nil
	} else {
		value = *accountBookID
	}

	res := r.db.WithContext(ctx).
		Table("users").
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"account_book_id": value,
			"updated_at":      time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrUserNotFound
	}
	return nil
}

func (r *FriendRepository) CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error) {
	var ownerCount int64
	if err := r.db.WithContext(ctx).
		Table("account_books").
		Where("id = ? AND user_id = ?", accountBookID, userID).
		Count(&ownerCount).Error; err != nil {
		return false, err
	}
	if ownerCount > 0 {
		return true, nil
	}

	var collaboratorCount int64
	if err := r.db.WithContext(ctx).
		Table("account_book_collaborators").
		Where("account_book_id = ? AND user_id = ? AND role IN ?", accountBookID, userID, []string{"owner", "admin"}).
		Count(&collaboratorCount).Error; err != nil {
		return false, err
	}
	return collaboratorCount > 0, nil
}

func (r *FriendRepository) CreateCollaborator(ctx context.Context, input tablemodel.AccountBookCollaborator) error {
	row := map[string]interface{}{
		"account_book_id": input.AccountBookID,
		"user_id":         input.UserID,
		"role":            strings.TrimSpace(input.Role),
		"remark":          strings.TrimSpace(input.Remark),
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}
	err := r.db.WithContext(ctx).Table("account_book_collaborators").Create(row).Error
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return repo.ErrFriendCollaboratorExists
	}
	return err
}

func (r *FriendRepository) UpdateCollaborator(ctx context.Context, accountBookID int64, collaboratorID int64, input repo.FriendCollaboratorUpdateRecord) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if input.Role != nil {
		updates["role"] = strings.TrimSpace(*input.Role)
	}
	if input.Remark != nil {
		updates["remark"] = strings.TrimSpace(*input.Remark)
	}

	res := r.db.WithContext(ctx).
		Table("account_book_collaborators").
		Where("id = ? AND account_book_id = ?", collaboratorID, accountBookID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrFriendCollaboratorNotFound
	}
	return nil
}

func (r *FriendRepository) DeleteCollaborator(ctx context.Context, accountBookID int64, collaboratorID int64) (tablemodel.AccountBookCollaborator, error) {
	var deleted tablemodel.AccountBookCollaborator
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row tablemodel.AccountBookCollaborator
		findErr := tx.
			Table("account_book_collaborators abc").
			Joins("INNER JOIN users u ON u.id = abc.user_id").
			Select([]string{
				"abc.id",
				"abc.account_book_id",
				"abc.user_id",
				"abc.role",
				"COALESCE(abc.remark, '') AS remark",
				"abc.created_at",
				"abc.updated_at",
			}).
			Where("abc.id = ? AND abc.account_book_id = ?", collaboratorID, accountBookID).
			Take(&row).Error
		if findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return repo.ErrFriendCollaboratorNotFound
			}
			return findErr
		}

		res := tx.Table("account_book_collaborators").Where("id = ? AND account_book_id = ?", collaboratorID, accountBookID).Delete(&struct{}{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return repo.ErrFriendCollaboratorNotFound
		}

		deleted = row
		return nil
	})
	if err != nil {
		return tablemodel.AccountBookCollaborator{}, err
	}
	return deleted, nil
}

func (r *FriendRepository) ListUsersByIDs(ctx context.Context, ids []int64) ([]tablemodel.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []tablemodel.User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}
