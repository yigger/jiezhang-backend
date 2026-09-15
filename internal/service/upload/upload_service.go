package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
)

var (
	ErrUploadInvalidInput = errors.New("upload invalid input")
	ErrUploadUnknownType  = errors.New("upload unknown type")
)

type UploadService struct {
	users      repo.UserRepository
	uploads    repo.UploadRepository
	statements repo.OwnerQuery
	urlBuilder repo.URLBuilder
	files      FileStore
}

type UploadInput struct {
	Type          string
	UserID        int64
	AccountBookID int64
	StatementID   int64
	File          io.Reader
	Filename      string
}

func NewUploadService(
	users repo.UserRepository,
	uploads repo.UploadRepository,
	statements repo.OwnerQuery,
	urlBuilder repo.URLBuilder,
	files FileStore,
) UploadService {
	return UploadService{
		users:      users,
		uploads:    uploads,
		statements: statements,
		urlBuilder: urlBuilder, files: files,
	}
}

func (s UploadService) Upload(ctx context.Context, input UploadInput) (types.UploadResult, error) {
	if input.UserID <= 0 || input.File == nil {
		return types.UploadResult{}, ErrUploadInvalidInput
	}

	uploadType := strings.TrimSpace(input.Type)
	switch uploadType {
	case "user_avatar":
		relPath, err := s.files.Save(ctx, filepath.Join("private", fmt.Sprintf("%d", input.UserID), "user"), input.Filename, input.File)
		if err != nil {
			return types.UploadResult{}, err
		}
		avatarURL := s.urlBuilder.BuildPublicURL("/" + relPath)
		if err := s.users.UpdateProfile(ctx, input.UserID, repo.UserProfileUpdateRecord{
			AvatarURL: &avatarURL,
		}); err != nil {
			return types.UploadResult{}, err
		}
		return types.UploadResult{Status: 200, AvatarPath: avatarURL}, nil

	case "bg_avatar":
		relPath, err := s.files.Save(ctx, filepath.Join("private", fmt.Sprintf("%d", input.UserID), "user"), input.Filename, input.File)
		if err != nil {
			return types.UploadResult{}, err
		}
		bgURL := s.urlBuilder.BuildPublicURL("/" + relPath)
		if err := s.users.SetBackgroundAvatarURL(ctx, input.UserID, bgURL); err != nil {
			return types.UploadResult{}, err
		}
		return types.UploadResult{Status: 200, AvatarPath: bgURL}, nil

	case "index_header_bg":
		relPath, err := s.files.Save(ctx, filepath.Join("private", fmt.Sprintf("%d", input.UserID), "user_avatar"), input.Filename, input.File)
		if err != nil {
			return types.UploadResult{}, err
		}
		if err := s.uploads.CreateUserAvatar(ctx, input.UserID, "/"+relPath); err != nil {
			return types.UploadResult{}, err
		}
		return types.UploadResult{Status: 200}, nil

	case "statement_upload":
		if input.AccountBookID <= 0 || input.StatementID <= 0 {
			return types.UploadResult{}, ErrUploadInvalidInput
		}
		ownerID, err := s.statements.GetOwnerID(ctx, input.StatementID, input.AccountBookID)
		if err != nil {
			return types.UploadResult{}, err
		}

		relPath, err := s.files.Save(ctx, filepath.Join("private", fmt.Sprintf("%d", ownerID), "statements", fmt.Sprintf("%d", input.StatementID)), input.Filename, input.File)
		if err != nil {
			return types.UploadResult{}, err
		}
		if err := s.uploads.CreateStatementAvatar(ctx, input.AccountBookID, input.StatementID, "/"+relPath); err != nil {
			return types.UploadResult{}, err
		}
		return types.UploadResult{Status: 200}, nil

	default:
		return types.UploadResult{}, ErrUploadUnknownType
	}
}

type FileStore interface {
	Save(context.Context, string, string, io.Reader) (string, error)
}
