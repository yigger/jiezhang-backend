package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

var (
	ErrUserInvalidInput  = errors.New("user invalid input")
	ErrUserQRCodeExpired = errors.New("qr code expired")
)

type UserService struct {
	repo  repository.UserRepository
	cache sessioncache.Cache
}

func NewUserService(repo repository.UserRepository, cache sessioncache.Cache) UserService {
	return UserService{repo: repo, cache: cache}
}

func (s UserService) FindByID(ctx context.Context, id int64) (domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s UserService) Create(ctx context.Context, name, email string) (domain.User, error) {
	user := domain.User{
		Nickname: strings.TrimSpace(name),
		Email:    strings.TrimSpace(strings.ToLower(email)),
	}
	return s.repo.Create(ctx, user)
}

type UserProfile struct {
	ID               int64  `json:"id"`
	ThemeID          int64  `json:"theme_id"`
	AvatarURL        string `json:"avatar_url"`
	Nickname         string `json:"nickname"`
	Persist          int64  `json:"persist"`
	StatementsCount  int64  `json:"sts_count"`
	Email            string `json:"email"`
	Remind           bool   `json:"remind"`
	HiddenAssetMoney bool   `json:"hidden_asset_money"`
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

func (s UserService) GetProfile(ctx context.Context, userID int64) (UserProfile, error) {
	record, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return UserProfile{}, err
	}

	return UserProfile{
		ID:               record.ID,
		ThemeID:          record.ThemeID,
		AvatarURL:        strings.TrimSpace(record.AvatarURL),
		Nickname:         strings.TrimSpace(record.Nickname),
		Persist:          record.Persist,
		StatementsCount:  record.StatementsCount,
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

	update := repository.UserProfileUpdateRecord{
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
