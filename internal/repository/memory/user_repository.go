package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

// UserRepository is an in-memory implementation useful for local development.
type UserRepository struct {
	mu     sync.RWMutex
	nextID int64
	users  map[int64]domain.User
}

func NewUserRepository() *UserRepository {
	now := time.Now()
	seed := map[int64]domain.User{
		1: {
			ID:        1,
			Name:      "Alice",
			Email:     "alice@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	return &UserRepository{
		nextID: 2,
		users:  seed,
	}
}

func (r *UserRepository) FindByID(_ context.Context, id int64) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return domain.User{}, repository.ErrUserNotFound
	}

	return user, nil
}

func (r *UserRepository) List(_ context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]int64, 0, len(r.users))
	for id := range r.users {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	users := make([]domain.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, r.users[id])
	}

	return users, nil
}

func (r *UserRepository) Create(_ context.Context, user domain.User) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	user.ID = r.nextID
	user.CreatedAt = now
	user.UpdatedAt = now

	r.users[user.ID] = user
	r.nextID++

	return user, nil
}
