package repo

import (
	"context"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

// Transactor commits all operations in the callback, or rolls all of them back.
type Transactor interface {
	WithinTransaction(context.Context, func(Mutation) error) error
}

// Mutation is only valid inside a transaction. LockCurrent serializes patch
// merging with competing writers; adapters must not expose database handles.
type Mutation interface {
	LockCurrent(context.Context, int64, int64) (tablemodel.Statement, error)
	Create(context.Context, tablemodel.Statement, BalanceEffect) (int64, error)
	UpdateByID(context.Context, int64, int64, tablemodel.Statement, BalanceEffect, BalanceEffect) error
	DeleteByID(context.Context, int64, int64, BalanceEffect) error
}
type DetailQuery interface {
	BatchLookup
	GetSimpleRowByID(context.Context, int64, int64) (tablemodel.Statement, error)
}
type OwnerQuery interface {
	GetOwnerID(context.Context, int64, int64) (int64, error)
}
type AvatarWriter interface {
	DeleteAvatarByID(context.Context, int64, int64, int64) error
}

// BalanceEffect carries service-calculated asset deltas for an atomic write.
type BalanceEffect struct {
	Source, Target float64
	HasTarget      bool
}
