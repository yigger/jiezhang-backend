package mysql

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type FriendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) (*FriendRepository, error) {
	return &FriendRepository{db: db}, nil
}

type friendCollaboratorRow struct {
	ID            int64     `gorm:"column:id"`
	AccountBookID int64     `gorm:"column:account_book_id"`
	UserID        int64     `gorm:"column:user_id"`
	Role          string    `gorm:"column:role"`
	Remark        string    `gorm:"column:remark"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
	UserNickname  string    `gorm:"column:user_nickname"`
	UserAvatarURL string    `gorm:"column:user_avatar_url"`
}

func (r *FriendRepository) ListCollaborators(ctx context.Context, accountBookID int64) ([]repository.FriendCollaboratorRecord, error) {
	rows := make([]friendCollaboratorRow, 0)
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
			"COALESCE(u.nickname, '') AS user_nickname",
			"COALESCE(u.avatar_url, '') AS user_avatar_url",
		}).
		Where("abc.account_book_id = ?", accountBookID).
		Order("abc.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toFriendCollaboratorRecords(rows), nil
}

func (r *FriendRepository) FindCollaboratorByUserID(ctx context.Context, accountBookID int64, userID int64) (repository.FriendCollaboratorRecord, error) {
	var row friendCollaboratorRow
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
			"COALESCE(u.nickname, '') AS user_nickname",
			"COALESCE(u.avatar_url, '') AS user_avatar_url",
		}).
		Where("abc.account_book_id = ? AND abc.user_id = ?", accountBookID, userID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.FriendCollaboratorRecord{}, repository.ErrFriendCollaboratorNotFound
		}
		return repository.FriendCollaboratorRecord{}, err
	}
	return toFriendCollaboratorRecord(row), nil
}

func (r *FriendRepository) FindCollaboratorByID(ctx context.Context, accountBookID int64, collaboratorID int64) (repository.FriendCollaboratorRecord, error) {
	var row friendCollaboratorRow
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
			"COALESCE(u.nickname, '') AS user_nickname",
			"COALESCE(u.avatar_url, '') AS user_avatar_url",
		}).
		Where("abc.id = ? AND abc.account_book_id = ?", collaboratorID, accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.FriendCollaboratorRecord{}, repository.ErrFriendCollaboratorNotFound
		}
		return repository.FriendCollaboratorRecord{}, err
	}
	return toFriendCollaboratorRecord(row), nil
}

func (r *FriendRepository) FindUserByID(ctx context.Context, userID int64) (repository.FriendUserRecord, error) {
	var row struct {
		ID        int64  `gorm:"column:id"`
		Nickname  string `gorm:"column:nickname"`
		AvatarURL string `gorm:"column:avatar_url"`
	}
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, COALESCE(nickname, '') AS nickname, COALESCE(avatar_url, '') AS avatar_url").
		Where("id = ?", userID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.FriendUserRecord{}, repository.ErrUserNotFound
		}
		return repository.FriendUserRecord{}, err
	}
	return repository.FriendUserRecord{
		ID:        row.ID,
		Nickname:  row.Nickname,
		AvatarURL: row.AvatarURL,
	}, nil
}

func (r *FriendRepository) FindAccountBookByID(ctx context.Context, accountBookID int64) (repository.FriendAccountBookRecord, error) {
	var row struct {
		ID     int64  `gorm:"column:id"`
		UserID int64  `gorm:"column:user_id"`
		Name   string `gorm:"column:name"`
	}
	err := r.db.WithContext(ctx).
		Table("account_books").
		Select("id, user_id, COALESCE(name, '') AS name").
		Where("id = ?", accountBookID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.FriendAccountBookRecord{}, repository.ErrFriendAccountBookNotFound
		}
		return repository.FriendAccountBookRecord{}, err
	}
	return repository.FriendAccountBookRecord{
		ID:     row.ID,
		UserID: row.UserID,
		Name:   row.Name,
	}, nil
}

func (r *FriendRepository) FindAccessibleAccountBookByID(ctx context.Context, userID int64, accountBookID int64) (repository.FriendAccountBookRecord, error) {
	var row struct {
		ID     int64  `gorm:"column:id"`
		UserID int64  `gorm:"column:user_id"`
		Name   string `gorm:"column:name"`
	}
	err := r.db.WithContext(ctx).
		Table("account_books ab").
		Joins("LEFT JOIN account_book_collaborators abc ON abc.account_book_id = ab.id").
		Select("ab.id, ab.user_id, COALESCE(ab.name, '') AS name").
		Where("ab.id = ? AND (ab.user_id = ? OR abc.user_id = ?)", accountBookID, userID, userID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.FriendAccountBookRecord{}, repository.ErrFriendAccountBookNotFound
		}
		return repository.FriendAccountBookRecord{}, err
	}
	return repository.FriendAccountBookRecord{
		ID:     row.ID,
		UserID: row.UserID,
		Name:   row.Name,
	}, nil
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
		return repository.ErrUserNotFound
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

func (r *FriendRepository) CreateCollaborator(ctx context.Context, input repository.FriendCollaboratorCreateRecord) error {
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
		return repository.ErrFriendCollaboratorExists
	}
	return err
}

func (r *FriendRepository) UpdateCollaborator(ctx context.Context, accountBookID int64, collaboratorID int64, input repository.FriendCollaboratorUpdateRecord) error {
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
		return repository.ErrFriendCollaboratorNotFound
	}
	return nil
}

func (r *FriendRepository) DeleteCollaborator(ctx context.Context, accountBookID int64, collaboratorID int64) (repository.FriendCollaboratorRecord, error) {
	var deleted repository.FriendCollaboratorRecord
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row friendCollaboratorRow
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
				"COALESCE(u.nickname, '') AS user_nickname",
				"COALESCE(u.avatar_url, '') AS user_avatar_url",
			}).
			Where("abc.id = ? AND abc.account_book_id = ?", collaboratorID, accountBookID).
			Take(&row).Error
		if findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return repository.ErrFriendCollaboratorNotFound
			}
			return findErr
		}

		res := tx.Table("account_book_collaborators").Where("id = ? AND account_book_id = ?", collaboratorID, accountBookID).Delete(&struct{}{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return repository.ErrFriendCollaboratorNotFound
		}

		deleted = toFriendCollaboratorRecord(row)
		return nil
	})
	if err != nil {
		return repository.FriendCollaboratorRecord{}, err
	}
	return deleted, nil
}

func toFriendCollaboratorRecords(rows []friendCollaboratorRow) []repository.FriendCollaboratorRecord {
	items := make([]repository.FriendCollaboratorRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, toFriendCollaboratorRecord(row))
	}
	return items
}

func toFriendCollaboratorRecord(row friendCollaboratorRow) repository.FriendCollaboratorRecord {
	return repository.FriendCollaboratorRecord{
		ID:            row.ID,
		AccountBookID: row.AccountBookID,
		UserID:        row.UserID,
		Role:          row.Role,
		Remark:        row.Remark,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		UserNickname:  row.UserNickname,
		UserAvatarURL: row.UserAvatarURL,
	}
}
