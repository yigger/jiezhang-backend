package user

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	"github.com/yigger/jiezhang-backend/internal/types"
	"strconv"
	"strings"
	"time"
)

var (
	ErrUserInvalidInput  = errors.New("user invalid input")
	ErrUserQRCodeExpired = errors.New("qr code expired")
)

type UserService struct {
	repo  repo.UserRepository
	cache repo.Cache
}

func NewUserService(repo repo.UserRepository, cache repo.Cache) UserService {
	return UserService{repo: repo, cache: cache}
}

func (s UserService) FindByID(ctx context.Context, id int64) (types.UserContext, error) {
	row, err := s.repo.FindByID(ctx, id)
	return helperservice.ToUserContext(row), err
}

func (s UserService) List(ctx context.Context) ([]types.UserContext, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]types.UserContext, 0, len(rows))
	for _, row := range rows {
		users = append(users, helperservice.ToUserContext(row))
	}
	return users, nil
}

func (s UserService) Create(ctx context.Context, name, email string) (types.UserContext, error) {
	user := types.UserContext{
		Nickname: strings.TrimSpace(name),
		Email:    strings.TrimSpace(strings.ToLower(email)),
	}
	row, err := s.repo.Create(ctx, helperservice.UserModel(user))
	return helperservice.ToUserContext(row), err
}

type UserProfileUpdateInput struct {
	ThemeID          *int64
	Country          *string
	City             *string
	Gender           *int
	Language         *string
	Province         *string
	BGAvatarID       *int64
	HiddenAssetMoney *bool
	AvatarURL        *string
	Nickname         *string
	BGAvatar         *string
}

func (s UserService) GetProfile(ctx context.Context, userID int64) (types.UserProfile, error) {
	row, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return types.UserProfile{}, err
	}

	record := helperservice.ToUserContext(row)
	counts, err := s.repo.StatementCounts(ctx, userID, row.AccountBookID)
	if err != nil {
		return types.UserProfile{}, err
	}

	return types.UserProfile{
		ID:               record.ID,
		ThemeID:          record.ThemeID,
		AvatarURL:        strings.TrimSpace(record.AvatarUrl),
		Nickname:         strings.TrimSpace(record.Nickname),
		Persist:          counts.Persist,
		StatementsCount:  counts.StatementsCount,
		Email:            strings.TrimSpace(record.Email),
		Remind:           record.Remind > 0,
		HiddenAssetMoney: record.HiddenAssetMoney,
	}, nil
}

func (s UserService) UpdateProfile(ctx context.Context, userID int64, input UserProfileUpdateInput) error {
	if input.BGAvatar != nil {
		if err := s.repo.SetBackgroundAvatarURL(ctx, userID, strings.TrimSpace(*input.BGAvatar)); err != nil {
			return err
		}
	}

	update := repo.UserProfileUpdateRecord{
		ThemeID:          input.ThemeID,
		Country:          trimStringPtr(input.Country),
		City:             trimStringPtr(input.City),
		Gender:           input.Gender,
		Language:         trimStringPtr(input.Language),
		Province:         trimStringPtr(input.Province),
		BGAvatarID:       input.BGAvatarID,
		HiddenAssetMoney: input.HiddenAssetMoney,
		AvatarURL:        trimStringPtr(input.AvatarURL),
		Nickname:         trimStringPtr(input.Nickname),
	}

	return s.repo.UpdateProfile(ctx, userID, update)
}

func (s UserService) ScanLogin(ctx context.Context, userID int64, qrCode string) error {
	if s.cache == nil {
		return ErrUserQRCodeExpired
	}

	qrCode = strings.TrimSpace(qrCode)
	if qrCode == "" {
		return ErrUserInvalidInput
	}

	qrCodeStatusKey := "qr_code:" + qrCode
	currentStatus, ok := s.cache.Get(qrCodeStatusKey)
	if !ok || strings.TrimSpace(currentStatus) != "pending" {
		return ErrUserQRCodeExpired
	}

	s.cache.Set("qr_code:user:"+qrCode, strconv.FormatInt(userID, 10), 5*time.Minute)
	s.cache.Set(qrCodeStatusKey, "success", 5*time.Minute)
	return nil
}

func trimStringPtr(in *string) *string {
	if in == nil {
		return nil
	}
	v := strings.TrimSpace(*in)
	return &v
}
