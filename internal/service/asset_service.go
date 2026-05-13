package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

var (
	ErrAssetPermissionDenied = errors.New("asset permission denied")
	ErrAssetInvalidInput     = errors.New("asset invalid input")
)

type AssetService struct {
	repo       repository.AssetRepository
	urlBuilder AssetURLBuilder
}

type AssetURLBuilder interface {
	BuildPublicURL(raw string) string
}

func NewAssetService(repo repository.AssetRepository, urlBuilder AssetURLBuilder) AssetService {
	return AssetService{repo: repo, urlBuilder: urlBuilder}
}

type AssetItem struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	Order    int         `json:"order"`
	IconPath string      `json:"icon_path"`
	IconURL  string      `json:"icon_url,omitempty"`
	ParentID int64       `json:"parent_id"`
	Type     string      `json:"type"`
	Amount   float64     `json:"amount"`
	Remark   string      `json:"remark,omitempty"`
	Childs   []AssetItem `json:"childs,omitempty"`
}

type AssetShowResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Order    int     `json:"order"`
	IconPath string  `json:"icon_path"`
	ParentID int64   `json:"parent_id"`
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Remark   string  `json:"remark"`
	IconURL  string  `json:"icon_url"`
}

type AssetWriteInput struct {
	CreatorID     int64
	AccountBookID int64
	Name          string
	Amount        string
	ParentID      int64
	IconPath      string
	Remark        string
	Type          string
}

type AssetSurplusInput struct {
	UserID        int64
	AccountBookID int64
	AssetID       int64
	Amount        string
}

func (s AssetService) ListByParent(ctx context.Context, accountBookID int64, parentID int64) ([]AssetItem, error) {
	rows, err := s.repo.ListByParent(ctx, accountBookID, parentID)
	if err != nil {
		return nil, err
	}
	items := make([]AssetItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, AssetItem{
			ID:       row.ID,
			Name:     row.Name,
			Order:    row.Order,
			IconPath: row.IconPath,
			IconURL:  s.buildPublicURL(row.IconPath),
			ParentID: row.ParentID,
			Type:     row.Type,
			Amount:   row.Amount,
			Remark:   row.Remark,
		})
	}
	return items, nil
}

func (s AssetService) ListTree(ctx context.Context, accountBookID int64) ([]AssetItem, error) {
	parents, err := s.repo.ListByParent(ctx, accountBookID, 0)
	if err != nil {
		return nil, err
	}
	parentIDs := make([]int64, 0, len(parents))
	for _, parent := range parents {
		parentIDs = append(parentIDs, parent.ID)
	}

	childrenRows, err := s.repo.ListChildrenByParentIDs(ctx, repository.AssetFilter{AccountBookID: accountBookID}, parentIDs)
	if err != nil {
		return nil, err
	}

	childrenByParent := make(map[int64][]AssetItem, len(parentIDs))
	for _, child := range childrenRows {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], AssetItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: child.IconPath,
			IconURL:  s.buildPublicURL(child.IconPath),
			ParentID: child.ParentID,
			Amount:   0,
		})
	}

	items := make([]AssetItem, 0, len(parents))
	for _, parent := range parents {
		childs := childrenByParent[parent.ID]
		if childs == nil {
			childs = []AssetItem{}
		}
		items = append(items, AssetItem{
			ID:       parent.ID,
			Name:     parent.Name,
			Order:    parent.Order,
			IconPath: parent.IconPath,
			IconURL:  s.buildPublicURL(parent.IconPath),
			ParentID: parent.ParentID,
			Type:     parent.Type,
			Amount:   parent.Amount,
			Remark:   parent.Remark,
			Childs:   childs,
		})
	}
	return items, nil
}

func (s AssetService) Show(ctx context.Context, accountBookID int64, id int64) (AssetShowResponse, error) {
	row, err := s.repo.FindByID(ctx, accountBookID, id)
	if err != nil {
		return AssetShowResponse{}, err
	}
	return AssetShowResponse{
		ID:       row.ID,
		Name:     row.Name,
		Order:    row.Order,
		IconPath: row.IconPath,
		ParentID: row.ParentID,
		Type:     row.Type,
		Amount:   row.Amount,
		Remark:   row.Remark,
		IconURL:  s.buildPublicURL(row.IconPath),
	}, nil
}

func (s AssetService) Create(ctx context.Context, input AssetWriteInput) error {
	record, err := s.normalizeWriteInput(input)
	if err != nil {
		return err
	}
	ok, err := s.repo.CanAdmin(ctx, input.AccountBookID, input.CreatorID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAssetPermissionDenied
	}
	_, err = s.repo.Create(ctx, record)
	return err
}

func (s AssetService) Update(ctx context.Context, id int64, input AssetWriteInput) error {
	current, err := s.repo.FindByID(ctx, input.AccountBookID, id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(input.Type) == "" {
		input.Type = current.Type
	}
	record, err := s.normalizeWriteInput(input)
	if err != nil {
		return err
	}
	ok, err := s.repo.CanAdmin(ctx, input.AccountBookID, input.CreatorID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAssetPermissionDenied
	}
	return s.repo.UpdateByID(ctx, id, input.AccountBookID, record)
}

func (s AssetService) Delete(ctx context.Context, id int64, accountBookID int64, userID int64) error {
	ok, err := s.repo.CanAdmin(ctx, accountBookID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAssetPermissionDenied
	}
	return s.repo.DeleteByID(ctx, id, accountBookID)
}

func (s AssetService) UpdateSurplus(ctx context.Context, input AssetSurplusInput) error {
	amount, err := strconv.ParseFloat(strings.TrimSpace(input.Amount), 64)
	if err != nil {
		return ErrAssetInvalidInput
	}
	ok, err := s.repo.CanAdmin(ctx, input.AccountBookID, input.UserID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAssetPermissionDenied
	}
	return s.repo.UpdateAmountByID(ctx, input.AssetID, input.AccountBookID, amount)
}

func (s AssetService) ListAssetIcons() ([]map[string]string, error) {
	dir := filepath.Join("public", "images", "asset")
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return []map[string]string{}, nil
	}
	entries, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)
	items := make([]map[string]string, 0, len(entries))
	for _, path := range entries {
		name := filepath.Base(path)
		if strings.TrimSpace(name) == "" {
			continue
		}
		raw := "/images/asset/" + name
		items = append(items, map[string]string{
			"id":  raw,
			"url": s.buildPublicURL(raw),
		})
	}
	return items, nil
}

func (s AssetService) normalizeWriteInput(input AssetWriteInput) (repository.AssetWriteRecord, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return repository.AssetWriteRecord{}, ErrAssetInvalidInput
	}
	amount := 0.0
	if strings.TrimSpace(input.Amount) != "" {
		v, err := strconv.ParseFloat(strings.TrimSpace(input.Amount), 64)
		if err != nil {
			return repository.AssetWriteRecord{}, ErrAssetInvalidInput
		}
		amount = v
	}

	assetType := strings.TrimSpace(input.Type)
	if assetType == "" {
		assetType = "deposit"
	}

	return repository.AssetWriteRecord{
		CreatorID:     input.CreatorID,
		AccountBookID: input.AccountBookID,
		Name:          name,
		Amount:        amount,
		ParentID:      input.ParentID,
		IconPath:      strings.TrimSpace(input.IconPath),
		Remark:        strings.TrimSpace(input.Remark),
		Type:          assetType,
	}, nil
}

func (s AssetService) buildPublicURL(raw string) string {
	if s.urlBuilder == nil {
		return raw
	}
	return s.urlBuilder.BuildPublicURL(raw)
}
