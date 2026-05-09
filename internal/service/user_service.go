package service

import (
	"context"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return UserService{repo: repo}
}

func (s UserService) FindByID(ctx context.Context, id int64) (domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s UserService) Create(ctx context.Context, name, email string) (domain.User, error) {
	user := domain.User{
		Name:  strings.TrimSpace(name),
		Email: strings.TrimSpace(strings.ToLower(email)),
	}
	return s.repo.Create(ctx, user)
}
