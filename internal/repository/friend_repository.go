package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrFriendCollaboratorNotFound = errors.New("friend collaborator not found")
	ErrFriendAccountBookNotFound  = errors.New("friend account book not found")
	ErrFriendCollaboratorExists   = errors.New("friend collaborator exists")
)

type FriendRepository interface {
	ListCollaborators(ctx context.Context, accountBookID int64) ([]FriendCollaboratorRecord, error)
	FindCollaboratorByUserID(ctx context.Context, accountBookID int64, userID int64) (FriendCollaboratorRecord, error)
	FindCollaboratorByID(ctx context.Context, accountBookID int64, collaboratorID int64) (FriendCollaboratorRecord, error)

	FindUserByID(ctx context.Context, userID int64) (FriendUserRecord, error)
	FindAccountBookByID(ctx context.Context, accountBookID int64) (FriendAccountBookRecord, error)
	FindAccessibleAccountBookByID(ctx context.Context, userID int64, accountBookID int64) (FriendAccountBookRecord, error)
	FindFirstOwnedAccountBookID(ctx context.Context, userID int64) (*int64, error)
	UpdateUserDefaultAccountBook(ctx context.Context, userID int64, accountBookID *int64) error

	CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error)

	CreateCollaborator(ctx context.Context, input FriendCollaboratorCreateRecord) error
	UpdateCollaborator(ctx context.Context, accountBookID int64, collaboratorID int64, input FriendCollaboratorUpdateRecord) error
	DeleteCollaborator(ctx context.Context, accountBookID int64, collaboratorID int64) (FriendCollaboratorRecord, error)
}

type FriendCollaboratorRecord struct {
	ID            int64
	AccountBookID int64
	UserID        int64
	Role          string
	Remark        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UserNickname  string
	UserAvatarURL string
}

type FriendUserRecord struct {
	ID        int64
	Nickname  string
	AvatarURL string
}

type FriendAccountBookRecord struct {
	ID     int64
	UserID int64
	Name   string
}

type FriendCollaboratorCreateRecord struct {
	AccountBookID int64
	UserID        int64
	Role          string
	Remark        string
}

type FriendCollaboratorUpdateRecord struct {
	Role   *string
	Remark *string
}
