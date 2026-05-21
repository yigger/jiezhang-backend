package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

var (
	ErrUploadInvalidInput = errors.New("upload invalid input")
	ErrUploadUnknownType  = errors.New("upload unknown type")
)

type UploadService struct {
	users      repository.UserRepository
	uploads    repository.UploadRepository
	statements repository.StatementRepository
	urlBuilder urlbuilder.PublicURLBuilder
}

type UploadInput struct {
	Type          string
	UserID        int64
	AccountBookID int64
	StatementID   int64
	File          *multipart.FileHeader
}

type UploadResult struct {
	Status     int    `json:"status"`
	AvatarPath string `json:"avatar_path,omitempty"`
}

func NewUploadService(
	users repository.UserRepository,
	uploads repository.UploadRepository,
	statements repository.StatementRepository,
	urlBuilder urlbuilder.PublicURLBuilder,
) UploadService {
	return UploadService{
		users:      users,
		uploads:    uploads,
		statements: statements,
		urlBuilder: urlBuilder,
	}
}

func (s UploadService) Upload(ctx context.Context, input UploadInput) (UploadResult, error) {
	if input.UserID <= 0 || input.File == nil {
		return UploadResult{}, ErrUploadInvalidInput
	}

	uploadType := strings.TrimSpace(input.Type)
	switch uploadType {
	case "user_avatar":
		relPath, err := s.saveFile(input.File, filepath.Join("private", fmt.Sprintf("%d", input.UserID), "user"))
		if err != nil {
			return UploadResult{}, err
		}
		avatarURL := s.urlBuilder.BuildPublicURL("/" + relPath)
		if err := s.users.UpdateProfile(ctx, input.UserID, repository.UserProfileUpdateRecord{
			AvatarURL: &avatarURL,
		}); err != nil {
			return UploadResult{}, err
		}
		return UploadResult{Status: 200, AvatarPath: avatarURL}, nil

	case "bg_avatar":
		relPath, err := s.saveFile(input.File, filepath.Join("private", fmt.Sprintf("%d", input.UserID), "user"))
		if err != nil {
			return UploadResult{}, err
		}
		bgURL := s.urlBuilder.BuildPublicURL("/" + relPath)
		if err := s.users.SetBackgroundAvatarURL(ctx, input.UserID, bgURL); err != nil {
			return UploadResult{}, err
		}
		return UploadResult{Status: 200, AvatarPath: bgURL}, nil

	case "index_header_bg":
		relPath, err := s.saveFile(input.File, filepath.Join("private", fmt.Sprintf("%d", input.UserID), "user_avatar"))
		if err != nil {
			return UploadResult{}, err
		}
		if err := s.uploads.CreateUserAvatar(ctx, input.UserID, "/"+relPath); err != nil {
			return UploadResult{}, err
		}
		return UploadResult{Status: 200}, nil

	case "statement_upload":
		if input.AccountBookID <= 0 || input.StatementID <= 0 {
			return UploadResult{}, ErrUploadInvalidInput
		}
		ownerID, err := s.statements.GetOwnerID(ctx, input.StatementID, input.AccountBookID)
		if err != nil {
			return UploadResult{}, err
		}

		relPath, err := s.saveFile(input.File, filepath.Join("private", fmt.Sprintf("%d", ownerID), "statements", fmt.Sprintf("%d", input.StatementID)))
		if err != nil {
			return UploadResult{}, err
		}
		if err := s.uploads.CreateStatementAvatar(ctx, input.AccountBookID, input.StatementID, "/"+relPath); err != nil {
			return UploadResult{}, err
		}
		return UploadResult{Status: 200}, nil

	default:
		return UploadResult{}, ErrUploadUnknownType
	}
}

func (s UploadService) saveFile(file *multipart.FileHeader, dir string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	rootDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	publicDir := filepath.Join(rootDir, "public")
	targetDir := filepath.Join(publicDir, dir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".bin"
	}
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), os.Getpid(), ext)
	targetPath := filepath.Join(targetDir, filename)

	dst, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	rel := filepath.ToSlash(filepath.Join(dir, filename))
	return rel, nil
}
