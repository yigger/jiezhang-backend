package repo

import (
	"context"
)

type UploadRepository interface {
	CreateUserAvatar(ctx context.Context, userID int64, path string) error
	CreateStatementAvatar(ctx context.Context, accountBookID int64, statementID int64, path string) error
}
