package insights

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
	"testing"
)

type storageFake struct {
	repo.InsightsStorage
	tx *storageTxFake
}

func (f storageFake) Transact(_ context.Context, fn func(repo.InsightsStorageTx) error) error {
	return fn(f.tx)
}

type storageTxFake struct {
	repo.InsightsStorageTx
	statement    model.Statement
	project      model.InsightProject
	assets       []model.Asset
	annotation   *model.InsightStatementAnnotation
	snapshot     *model.InsightPortfolioSnapshot
	projectSaved bool
}

func (f *storageTxFake) LockBook(id int64) (model.AccountBook, error) {
	return model.AccountBook{ID: id, UserID: 1}, nil
}
func (f *storageTxFake) Members(_ int64) ([]model.AccountBookCollaborator, error) {
	return []model.AccountBookCollaborator{{UserID: 2, Role: "member"}}, nil
}
func (f *storageTxFake) LockStatement(book, id int64) (model.Statement, error) {
	if f.statement.ID != id || f.statement.AccountBookID != book {
		return model.Statement{}, errors.New("missing statement")
	}
	return f.statement, nil
}
func (f *storageTxFake) Project(book, id int64) (model.InsightProject, error) {
	if f.project.ID != id || f.project.AccountBookID != book {
		return model.InsightProject{}, errors.New("missing project")
	}
	return f.project, nil
}
func (f *storageTxFake) SaveProject(p *model.InsightProject) error { f.projectSaved = true; return nil }
func (f *storageTxFake) SaveAnnotation(a *model.InsightStatementAnnotation) error {
	f.annotation = a
	return nil
}
func (f *storageTxFake) Assets(_ int64) ([]model.Asset, error) { return f.assets, nil }
func (f *storageTxFake) CreateSnapshot(a *model.InsightPortfolioSnapshot) error {
	f.snapshot = a
	return nil
}
func TestAnnotationPermissionsAndExactSplit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		user  int64
		payer int64
		split int64
		want  error
	}{{"author", 2, 1, 1234, nil}, {"owner", 1, 2, 1234, nil}, {"outsider", 9, 1, 1234, ErrForbidden}, {"unknown payer", 2, 9, 1234, ErrInvalidInput}, {"split mismatch", 2, 1, 1233, ErrInvalidInput}} {
		t.Run(tc.name, func(t *testing.T) {
			tx := &storageTxFake{statement: model.Statement{ID: 5, AccountBookID: 8, UserID: 2, Type: "expend", Amount: 12.34}}
			s := NewStorage(storageFake{tx: tx}, nil)
			in := types.InsightAnnotationInput{StatementID: 5, PayerID: &tc.payer, Allocations: []types.InsightAllocation{{MemberID: 2, AmountCents: tc.split}}}
			err := s.SaveAnnotation(context.Background(), 8, tc.user, in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
			if (tx.annotation != nil) != (tc.want == nil) {
				t.Fatal("invalid metadata was saved")
			}
		})
	}
}
func TestAnnotationRejectsArchivedProjectAndDuplicateSplit(t *testing.T) {
	tx := &storageTxFake{statement: model.Statement{ID: 5, AccountBookID: 8, UserID: 2, Type: "expend", Amount: 10}, project: model.InsightProject{ID: 3, AccountBookID: 8, Archived: true}}
	s := NewStorage(storageFake{tx: tx}, nil)
	id := int64(3)
	if err := s.SaveAnnotation(context.Background(), 8, 2, types.InsightAnnotationInput{StatementID: 5, ProjectID: &id}); !errors.Is(err, ErrInvalidInput) {
		t.Fatal(err)
	}
	if err := s.SaveAnnotation(context.Background(), 8, 2, types.InsightAnnotationInput{StatementID: 5, Allocations: []types.InsightAllocation{{MemberID: 2, AmountCents: 500}, {MemberID: 2, AmountCents: 500}}}); !errors.Is(err, ErrInvalidInput) {
		t.Fatal(err)
	}
}
func TestProjectOnlyCreatorOrOwnerCanModify(t *testing.T) {
	tx := &storageTxFake{project: model.InsightProject{ID: 3, AccountBookID: 8, CreatorID: 1}}
	s := NewStorage(storageFake{tx: tx}, nil)
	_, err := s.SaveProject(context.Background(), 8, 2, types.InsightProjectInput{ID: 3, Name: "旅行", BudgetCents: 10000})
	if !errors.Is(err, ErrForbidden) || tx.projectSaved {
		t.Fatal("member changed another member's project")
	}
	_, err = s.SaveProject(context.Background(), 8, 1, types.InsightProjectInput{ID: 3, Name: "旅行", BudgetCents: 10000})
	if err != nil || !tx.projectSaved {
		t.Fatal(err)
	}
}
func TestSnapshotPreservesNegativeBalancesAndLiabilities(t *testing.T) {
	tx := &storageTxFake{assets: []model.Asset{{ID: 1, Name: "存款", Type: "deposit", Amount: 100.1}, {ID: 2, Name: "透支", Type: "deposit", Amount: -10}, {ID: 3, Name: "信用卡", Type: "debt", Amount: 30.2}}}
	s := NewStorage(storageFake{tx: tx}, nil)
	r, err := s.CapturePortfolio(context.Background(), 8, 2, "月末")
	if err != nil {
		t.Fatal(err)
	}
	if r.AssetsCents != 9010 || r.LiabilitiesCents != 3020 || r.AssetsCents-r.LiabilitiesCents != 5990 {
		t.Fatalf("snapshot totals %+v", r)
	}
	var balances []types.InsightAssetBalance
	if err = json.Unmarshal(r.Balances, &balances); err != nil || len(balances) != 3 || balances[1].AmountCents != -1000 {
		t.Fatal("snapshot details lost")
	}
	if tx.snapshot == nil {
		t.Fatal("snapshot not persisted")
	}
}

func (f *storageTxFake) Annotation(book, statement int64) (model.InsightStatementAnnotation, bool, error) {
	if f.annotation == nil {
		return model.InsightStatementAnnotation{}, false, nil
	}
	return *f.annotation, true, nil
}

func TestExistingArchivedProjectDoesNotBlockPaymentUpdates(t *testing.T) {
	id := int64(3)
	payer := int64(1)
	tx := &storageTxFake{statement: model.Statement{ID: 5, AccountBookID: 8, UserID: 2, Type: "expend", Amount: 10}, project: model.InsightProject{ID: 3, AccountBookID: 8, Archived: true}, annotation: &model.InsightStatementAnnotation{StatementID: 5, ProjectID: &id}}
	service := NewStorage(storageFake{tx: tx}, nil)
	if err := service.SaveAnnotation(context.Background(), 8, 2, types.InsightAnnotationInput{StatementID: 5, ProjectID: &id, PayerID: &payer}); err != nil {
		t.Fatal(err)
	}
	if tx.annotation.PayerID == nil || *tx.annotation.PayerID != 1 {
		t.Fatal("payment metadata not saved")
	}
}

func TestProjectRejectsUnknownParticipantsAndReversedDates(t *testing.T) {
	tx := &storageTxFake{}
	s := NewStorage(storageFake{tx: tx}, nil)
	start, end := "2026-10-10", "2026-10-01"
	for _, in := range []types.InsightProjectInput{{Name: "旅行", ParticipantIDs: []int64{9}}, {Name: "旅行", ParticipantIDs: []int64{1, 1}}, {Name: "旅行", StartDate: &start, EndDate: &end}} {
		if _, err := s.SaveProject(context.Background(), 8, 1, in); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("invalid project accepted %+v %v", in, err)
		}
	}
}

func TestProjectAppearanceValidationAndCompatibility(t *testing.T) {
	oldIcon, oldColor := "jcon-car", "#123456"
	for _, tc := range []struct {
		name                string
		icon, color         *string
		wantIcon, wantColor string
		invalid             bool
	}{
		{"omitted preserves", nil, nil, oldIcon, oldColor, false},
		{"normalizes", strptr(" jcon-home "), strptr("#aBcDeF"), "jcon-home", "#ABCDEF", false},
		{"empty resets", strptr(""), strptr(""), "", "", false},
		{"invalid icon", strptr("jcon-car extra"), nil, "", "", true},
		{"invalid color", nil, strptr("red;"), "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := &storageTxFake{project: model.InsightProject{ID: 3, AccountBookID: 8, CreatorID: 1, Icon: &oldIcon, Color: &oldColor}}
			result, err := NewStorage(storageFake{tx: tx}, nil).SaveProject(context.Background(), 8, 1, types.InsightProjectInput{ID: 3, Name: "旅行", Icon: tc.icon, Color: tc.color})
			if tc.invalid {
				if !errors.Is(err, ErrInvalidInput) || tx.projectSaved {
					t.Fatal("invalid appearance saved", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			value := func(p *string) string {
				if p == nil {
					return ""
				}
				return *p
			}
			if value(result.Icon) != tc.wantIcon || value(result.Color) != tc.wantColor {
				t.Fatal("appearance lost")
			}
		})
	}
}

func strptr(v string) *string { return &v }
