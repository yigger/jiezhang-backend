package helper

import (
	"context"
	"errors"
	"testing"

	model "github.com/yigger/jiezhang-backend/internal/model"
)

type relationStore struct {
	categories                map[int64]model.Category
	assets                    map[int64]model.Asset
	categoryCalls, assetCalls int
	failure                   error
}

func (s *relationStore) BatchGetCategories(_ context.Context, ids []int64) ([]model.Category, error) {
	s.categoryCalls++
	var rows []model.Category
	for _, id := range ids {
		if row, ok := s.categories[id]; ok {
			rows = append(rows, row)
		}
	}
	return rows, nil
}
func (s *relationStore) BatchGetAssets(_ context.Context, ids []int64) ([]model.Asset, error) {
	s.assetCalls++
	if s.failure != nil {
		return nil, s.failure
	}
	var rows []model.Asset
	for _, id := range ids {
		if row, ok := s.assets[id]; ok {
			rows = append(rows, row)
		}
	}
	return rows, nil
}
func (*relationStore) BatchGetPayees(context.Context, []int64) ([]model.Payee, error) {
	return []model.Payee{{ID: 7, Name: "商家"}}, nil
}
func (*relationStore) BatchGetCollaboratorRemarks(_ context.Context, bookID int64, _ []int64) ([]model.AccountBookCollaborator, error) {
	if bookID != 9 {
		return nil, errors.New("wrong book")
	}
	return []model.AccountBookCollaborator{{UserID: 5, Remark: "成员备注"}}, nil
}
func (*relationStore) BatchStatementAvatars(context.Context, []int64) ([]model.UserAsset, error) {
	return []model.UserAsset{{ImageableID: 2}}, nil
}
func TestAssembleRowsUsesBatchModelsAndPreservesMissingRelations(t *testing.T) {
	target, payee := int64(4), int64(7)
	store := &relationStore{
		categories: map[int64]model.Category{1: {ID: 1, Name: "餐饮", ParentID: 10, IconPath: "food.png"}, 10: {ID: 10, Name: "支出"}},
		assets:     map[int64]model.Asset{3: {ID: 3, Name: "现金", ParentID: 30}, 30: {ID: 30, Name: "钱包"}, 4: {ID: 4, Name: "银行卡"}},
	}
	input := []model.Statement{
		{ID: 2, UserID: 5, CategoryID: 1, AssetID: 3, TargetAssetID: &target, PayeeID: &payee, Amount: 12.5, Description: "午饭"},
		{ID: 1, UserID: 6, CategoryID: 999, AssetID: 999},
	}
	rows, err := AssembleRows(context.Background(), store, 9, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].ID != 2 || rows[1].ID != 1 {
		t.Fatalf("order changed: %+v", rows)
	}
	first := rows[0]
	if first.CategoryName != "餐饮" || first.CategoryParentName != "支出" || first.AssetName != "现金" || first.AssetParentName != "钱包" || first.TargetAssetName != "银行卡" || first.PayeeName != "商家" || first.Remark != "成员备注" || !first.HasPic {
		t.Fatalf("relations: %+v", first)
	}
	if first.Amount != 12.5 || first.Description != "午饭" || first.CategoryParentID != 10 {
		t.Fatalf("fields lost: %+v", first)
	}
	second := rows[1]
	if second.PayeeID != 0 || second.TargetAssetID != 0 || second.CategoryName != "" || second.AssetName != "" || second.HasPic {
		t.Fatalf("missing relations: %+v", second)
	}
	if store.categoryCalls != 2 || store.assetCalls != 2 {
		t.Fatalf("expected batched child/parent queries: %+v", store)
	}
	if input[0].TargetAssetID != &target || input[1].PayeeID != nil {
		t.Fatal("input models mutated")
	}
}
func TestAssembleRowsPropagatesLookupFailure(t *testing.T) {
	failure := errors.New("asset lookup failed")
	_, err := AssembleRows(context.Background(), &relationStore{failure: failure}, 9, []model.Statement{{ID: 1}})
	if !errors.Is(err, failure) {
		t.Fatalf("got %v", err)
	}
}
