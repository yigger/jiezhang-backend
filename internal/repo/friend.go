package repo

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var (
	ErrFriendCollaboratorNotFound = errors.New("friend collaborator not found")
	ErrFriendAccountBookNotFound  = errors.New("friend account book not found")
	ErrFriendCollaboratorExists   = errors.New("friend collaborator exists")
)

type FriendRepository interface {
	ListUsersByIDs(context.Context, []int64) ([]tablemodel.User, error)
	ListCollaborators(ctx context.Context, accountBookID int64) ([]tablemodel.AccountBookCollaborator, error)
	FindCollaboratorByUserID(ctx context.Context, accountBookID int64, userID int64) (tablemodel.AccountBookCollaborator, error)
	FindCollaboratorByID(ctx context.Context, accountBookID int64, collaboratorID int64) (tablemodel.AccountBookCollaborator, error)

	FindUserByID(ctx context.Context, userID int64) (tablemodel.User, error)
	FindAccountBookByID(ctx context.Context, accountBookID int64) (tablemodel.AccountBook, error)
	FindAccessibleAccountBookByID(ctx context.Context, userID int64, accountBookID int64) (tablemodel.AccountBook, error)
	FindFirstOwnedAccountBookID(ctx context.Context, userID int64) (*int64, error)
	UpdateUserDefaultAccountBook(ctx context.Context, userID int64, accountBookID *int64) error

	CanAdmin(ctx context.Context, accountBookID int64, userID int64) (bool, error)

	CreateCollaborator(ctx context.Context, input tablemodel.AccountBookCollaborator) error
	UpdateCollaborator(ctx context.Context, accountBookID int64, collaboratorID int64, input FriendCollaboratorUpdateRecord) error
	DeleteCollaborator(ctx context.Context, accountBookID int64, collaboratorID int64) (tablemodel.AccountBookCollaborator, error)
}

type FriendCollaboratorUpdateRecord struct {
	Role   *string
	Remark *string
}
