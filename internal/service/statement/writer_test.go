package statement

import (
	"context"
	"errors"
	"testing"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

type writerStore struct {
	oldEffect, effect                repo.BalanceEffect
	current                          tablemodel.Statement
	written                          tablemodel.Statement
	locked, inTx, committed, deleted bool
	failure                          error
}

func (s *writerStore) WithinTransaction(ctx context.Context, fn func(repo.Mutation) error) error {
	s.inTx = true
	defer func() { s.inTx = false }()
	e := fn(s)
	s.committed = e == nil
	return e
}
func (s *writerStore) LockCurrent(context.Context, int64, int64) (tablemodel.Statement, error) {
	if !s.inTx {
		panic("read outside transaction")
	}
	s.locked = true
	return s.current, nil
}
func (s *writerStore) Create(_ context.Context, r tablemodel.Statement, effect repo.BalanceEffect) (int64, error) {
	if !s.inTx {
		panic("create outside transaction")
	}
	s.written, s.effect = r, effect
	return 1, s.failure
}
func (s *writerStore) UpdateByID(_ context.Context, _, _ int64, r tablemodel.Statement, oldEffect, effect repo.BalanceEffect) error {
	if !s.locked || !s.inTx {
		panic("write without lock")
	}
	s.written = r
	s.oldEffect, s.effect = oldEffect, effect
	return s.failure
}
func (s *writerStore) DeleteByID(_ context.Context, _, _ int64, effect repo.BalanceEffect) error {
	if !s.inTx || !s.locked {
		panic("delete without lock")
	}
	s.effect = effect
	s.deleted = true
	return s.failure
}
func (s *writerStore) GetSimpleRowByID(context.Context, int64, int64) (tablemodel.Statement, error) {
	if s.inTx {
		panic("response query inside tx")
	}
	return s.current, nil
}
func TestUpdateMergesLockedCurrentAndChecksOwner(t *testing.T) {
	for _, tc := range []struct {
		name    string
		user    int64
		failure error
	}{{"partial update", 2, nil}, {"non owner", 9, nil}, {"write failure", 2, errors.New("write failed")}} {
		t.Run(tc.name, func(t *testing.T) {
			store := &writerStore{current: tablemodel.Statement{ID: 1, UserID: 2, Type: "income", Amount: 10, AssetID: 3, CategoryID: 4, Description: "latest concurrent value", CreatedAt: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)}, failure: tc.failure}
			writer := NewWriter(store, nil, store, nil, NewRowMapper(nil))
			amount := float64(20)
			_, err := writer.UpdateStatement(context.Background(), UpdateInput{StatementID: 1, AccountBookID: 5, UserID: tc.user, Patch: PatchInput{Amount: &amount}})
			if tc.user != 2 {
				if !errors.Is(err, ErrStatementPermissionDenied) || store.committed || store.written.Amount != 0 {
					t.Fatalf("unauthorized write: %v %+v", err, store)
				}
				return
			}
			if store.oldEffect.Source != 10 || store.effect.Source != 20 {
				t.Fatalf("incorrect balance deltas: %+v %+v", store.oldEffect, store.effect)
			}
			if store.written.Year != 2026 || store.written.Month != 9 || store.written.Day != 1 || store.written.TimeText != "12:00" {
				t.Fatalf("occurrence fields lost: %+v", store.written)
			}
			if store.written.Description != "latest concurrent value" || store.written.Amount != 20 {
				t.Fatalf("merged=%+v", store.written)
			}
			if tc.failure != nil {
				if !errors.Is(err, tc.failure) || store.committed {
					t.Fatalf("commit after failure: %v", err)
				}
			} else if err != nil || !store.committed {
				t.Fatalf("update failed: %v", err)
			}
		})
	}
}
func TestDeleteRejectsNonOwner(t *testing.T) {
	s := &writerStore{current: tablemodel.Statement{UserID: 2}}
	w := NewWriter(s, nil, s, nil, NewRowMapper(nil))
	err := w.DeleteStatement(context.Background(), 1, 9, 5)
	if !errors.Is(err, ErrStatementPermissionDenied) || s.deleted || s.committed {
		t.Fatalf("unauthorized delete: %v", err)
	}
}

func (s *writerStore) BatchGetAssets(context.Context, []int64) ([]tablemodel.Asset, error) {
	return nil, nil
}

func (s *writerStore) BatchGetCategories(context.Context, []int64) ([]tablemodel.Category, error) {
	return nil, nil
}

func (s *writerStore) BatchGetPayees(context.Context, []int64) ([]tablemodel.Payee, error) {
	return nil, nil
}

func (s *writerStore) BatchGetCollaboratorRemarks(context.Context, int64, []int64) ([]tablemodel.AccountBookCollaborator, error) {
	return nil, nil
}

func (s *writerStore) BatchStatementAvatars(context.Context, []int64) ([]tablemodel.UserAsset, error) {
	return nil, nil
}

func TestCreateAndDeletePassBalanceEffects(t *testing.T) {
	store := &writerStore{current: tablemodel.Statement{ID: 1, UserID: 2, Type: "repayment", Amount: 25}}
	writer := NewWriter(store, nil, store, nil, NewRowMapper(nil))
	_, err := writer.CreateStatement(context.Background(), WriteInput{UserID: 2, AccountBookID: 5, Type: "income", Amount: 30, AssetID: 3, CategoryID: 4})
	if err != nil || !store.committed || store.effect != (repo.BalanceEffect{Source: 30}) {
		t.Fatalf("create: err=%v effect=%+v committed=%v", err, store.effect, store.committed)
	}
	err = writer.DeleteStatement(context.Background(), 1, 2, 5)
	if err != nil || !store.committed || !store.deleted || store.effect != (repo.BalanceEffect{Source: -25, Target: -25, HasTarget: true}) {
		t.Fatalf("delete: err=%v effect=%+v committed=%v", err, store.effect, store.committed)
	}
}
