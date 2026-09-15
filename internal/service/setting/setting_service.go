package setting

import (
	"context"
	"errors"
	"strings"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

var ErrSettingInvalidInput = errors.New("setting invalid input")

type SettingService struct {
	repo repo.SettingRepository
}

func NewSettingService(repo repo.SettingRepository) SettingService {
	return SettingService{repo: repo}
}

type SettingFeedbackInput struct {
	UserID  int64
	Content string
	Type    int
}

func (s SettingService) SubmitFeedback(ctx context.Context, input SettingFeedbackInput) error {
	content := strings.TrimSpace(input.Content)
	if input.UserID <= 0 || content == "" {
		return ErrSettingInvalidInput
	}
	now := time.Now()
	return s.repo.CreateFeedback(ctx, tablemodel.Feedback{
		UserID:    input.UserID,
		Content:   content,
		Type:      input.Type,
		CreatedAt: now,
		UpdatedAt: now,
	})
}
