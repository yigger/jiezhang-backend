package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

//go:embed account_book_preset_data.json
var accountBookPresetRaw []byte

var (
	ErrAccountBookNotFound         = errors.New("account book not found")
	ErrAccountBookInvalidInput     = errors.New("account book invalid input")
	ErrAccountBookInvalidType      = errors.New("account book invalid type")
	ErrAccountBookPermissionDenied = errors.New("account book permission denied")
	ErrAccountBookInUse            = errors.New("account book in use")
)

var accountTypeMap = map[string]int{
	"daily":    0,
	"family":   1,
	"travel":   2,
	"business": 3,
}

var accountTypeNameMap = map[string]string{
	"daily":    "基础账簿",
	"family":   "家庭账簿",
	"travel":   "旅行账簿",
	"business": "生意账簿",
}

type AccountBookService struct {
	repo        repository.AccountBookRepository
	presetByKey map[string]accountBookPreset
}

func NewAccountBookService(repo repository.AccountBookRepository) AccountBookService {
	presetByKey := make(map[string]accountBookPreset)
	_ = json.Unmarshal(accountBookPresetRaw, &presetByKey)
	return AccountBookService{repo: repo, presetByKey: presetByKey}
}

type AccountBookTypeItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccountBookListItem struct {
	ID              int64       `json:"id"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	AccountType     int         `json:"account_type"`
	AccountTypeName string      `json:"account_type_name"`
	UserID          int64       `json:"user_id"`
	Budget          float64     `json:"budget"`
	CreatedAt       interface{} `json:"created_at"`
	UpdatedAt       interface{} `json:"updated_at"`
}

type AccountBookTypeInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccountBookDetailItem struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	AccountType AccountBookTypeInfo `json:"account_type"`
}

type AccountBookCreateInput struct {
	UserID       int64
	UserNickname string
	Name         string
	Description  string
	AccountType  string
	Categories   map[string][]AccountBookCategoryInput
	Assets       []AccountBookAssetInput
}

type AccountBookUpdateInput struct {
	UserID      int64
	ID          int64
	Name        string
	Description string
	AccountType string
}

type AccountBookCategoryInput struct {
	Name     string
	IconPath string
	Childs   []AccountBookChildInput
}

type AccountBookAssetInput struct {
	Name     string
	IconPath string
	Type     string
	Childs   []AccountBookChildInput
}

type AccountBookChildInput struct {
	Name     string
	IconPath string
}

type accountBookPreset struct {
	Name       string                                 `json:"name"`
	Categories map[string][]accountBookPresetCategory `json:"categories"`
	Assets     []accountBookPresetAsset               `json:"assets"`
}

type accountBookPresetCategory struct {
	Name     string                   `json:"name"`
	IconPath string                   `json:"icon_path"`
	Childs   []accountBookPresetChild `json:"childs"`
}

type accountBookPresetAsset struct {
	Name     string                   `json:"name"`
	IconPath string                   `json:"icon_path"`
	Type     string                   `json:"type"`
	Childs   []accountBookPresetChild `json:"childs"`
}

type accountBookPresetChild struct {
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

func (s AccountBookService) List(ctx context.Context, userID int64) ([]AccountBookListItem, error) {
	rows, err := s.repo.ListAccessible(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]AccountBookListItem, 0, len(rows))
	for _, row := range rows {
		_, typeName := accountTypeByValue(row.AccountType)
		items = append(items, AccountBookListItem{
			ID:              row.ID,
			Name:            row.Name,
			Description:     row.Description,
			AccountType:     row.AccountType,
			AccountTypeName: typeName,
			UserID:          row.UserID,
			Budget:          row.Budget,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return items, nil
}

func (s AccountBookService) GetByID(ctx context.Context, id int64, userID int64) (AccountBookDetailItem, error) {
	row, err := s.repo.FindAccessibleByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountBookNotFound) {
			return AccountBookDetailItem{}, ErrAccountBookNotFound
		}
		return AccountBookDetailItem{}, err
	}
	typeID, typeName := accountTypeByValue(row.AccountType)
	return AccountBookDetailItem{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		AccountType: AccountBookTypeInfo{ID: typeID, Name: typeName},
	}, nil
}

func (s AccountBookService) Types() []AccountBookTypeItem {
	items := make([]AccountBookTypeItem, 0, len(accountTypeNameMap))
	for id, name := range accountTypeNameMap {
		items = append(items, AccountBookTypeItem{ID: id, Name: name})
	}
	return items
}

func (s AccountBookService) PresetCategories(accountTypeID string) (accountBookPreset, error) {
	accountTypeID = strings.TrimSpace(accountTypeID)
	if _, ok := accountTypeMap[accountTypeID]; !ok {
		return accountBookPreset{}, ErrAccountBookInvalidType
	}
	preset, ok := s.presetByKey[accountTypeID]
	if !ok {
		return accountBookPreset{}, ErrAccountBookInvalidType
	}
	return preset, nil
}

func (s AccountBookService) Switch(ctx context.Context, userID int64, accountBookID int64) error {
	if _, err := s.repo.FindAccessibleByID(ctx, accountBookID, userID); err != nil {
		if errors.Is(err, repository.ErrAccountBookNotFound) {
			return ErrAccountBookNotFound
		}
		return err
	}
	return s.repo.SwitchDefaultByUserID(ctx, userID, accountBookID)
}

func (s AccountBookService) Create(ctx context.Context, input AccountBookCreateInput) (AccountBookListItem, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return AccountBookListItem{}, ErrAccountBookInvalidInput
	}
	accountTypeID := strings.TrimSpace(input.AccountType)
	accountTypeValue, ok := accountTypeMap[accountTypeID]
	if !ok {
		return AccountBookListItem{}, ErrAccountBookInvalidType
	}

	categories := normalizeCategories(input.Categories)
	assets := normalizeAssets(input.Assets)
	if len(categories) == 0 && len(assets) == 0 {
		preset, presetOK := s.presetByKey[accountTypeID]
		if presetOK {
			categories = fromPresetCategories(preset.Categories)
			assets = fromPresetAssets(preset.Assets)
		}
	}

	record, err := s.repo.Create(ctx, repository.AccountBookCreateInput{
		UserID:       input.UserID,
		UserNickname: strings.TrimSpace(input.UserNickname),
		Name:         name,
		Description:  strings.TrimSpace(input.Description),
		AccountType:  accountTypeValue,
		Categories:   categories,
		Assets:       assets,
	})
	if err != nil {
		return AccountBookListItem{}, err
	}

	_, typeName := accountTypeByValue(record.AccountType)
	return AccountBookListItem{
		ID:              record.ID,
		Name:            record.Name,
		Description:     record.Description,
		AccountType:     record.AccountType,
		AccountTypeName: typeName,
		UserID:          record.UserID,
		Budget:          record.Budget,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}, nil
}

func (s AccountBookService) Update(ctx context.Context, input AccountBookUpdateInput) (AccountBookListItem, error) {
	row, err := s.repo.FindAccessibleByID(ctx, input.ID, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountBookNotFound) {
			return AccountBookListItem{}, ErrAccountBookNotFound
		}
		return AccountBookListItem{}, err
	}
	if row.UserID != input.UserID {
		return AccountBookListItem{}, ErrAccountBookPermissionDenied
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return AccountBookListItem{}, ErrAccountBookInvalidInput
	}
	accountTypeValue, ok := accountTypeMap[strings.TrimSpace(input.AccountType)]
	if !ok {
		return AccountBookListItem{}, ErrAccountBookInvalidType
	}
	if err := s.repo.UpdateByID(ctx, input.ID, repository.AccountBookUpdateInput{
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		AccountType: accountTypeValue,
	}); err != nil {
		if errors.Is(err, repository.ErrAccountBookNotFound) {
			return AccountBookListItem{}, ErrAccountBookNotFound
		}
		return AccountBookListItem{}, err
	}

	updated, err := s.repo.FindAccessibleByID(ctx, input.ID, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountBookNotFound) {
			return AccountBookListItem{}, ErrAccountBookNotFound
		}
		return AccountBookListItem{}, err
	}
	_, typeName := accountTypeByValue(updated.AccountType)
	return AccountBookListItem{
		ID:              updated.ID,
		Name:            updated.Name,
		Description:     updated.Description,
		AccountType:     updated.AccountType,
		AccountTypeName: typeName,
		UserID:          updated.UserID,
		Budget:          updated.Budget,
		CreatedAt:       updated.CreatedAt,
		UpdatedAt:       updated.UpdatedAt,
	}, nil
}

func (s AccountBookService) Delete(ctx context.Context, userID int64, accountBookID int64, currentUsingAccountBookID int64) error {
	row, err := s.repo.FindAccessibleByID(ctx, accountBookID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountBookNotFound) {
			return ErrAccountBookNotFound
		}
		return err
	}
	if row.UserID != userID {
		return ErrAccountBookPermissionDenied
	}
	if currentUsingAccountBookID == accountBookID {
		return ErrAccountBookInUse
	}
	if err := s.repo.DeleteByID(ctx, accountBookID); err != nil {
		if errors.Is(err, repository.ErrAccountBookNotFound) {
			return ErrAccountBookNotFound
		}
		return err
	}
	return nil
}

func accountTypeByValue(accountType int) (string, string) {
	for id, value := range accountTypeMap {
		if value == accountType {
			return id, accountTypeNameMap[id]
		}
	}
	return "daily", accountTypeNameMap["daily"]
}

func normalizeCategories(src map[string][]AccountBookCategoryInput) map[string][]repository.AccountBookCategoryTemplate {
	if len(src) == 0 {
		return nil
	}
	res := make(map[string][]repository.AccountBookCategoryTemplate, len(src))
	for statementType, parents := range src {
		cleanType := strings.TrimSpace(statementType)
		if cleanType == "" {
			continue
		}
		items := make([]repository.AccountBookCategoryTemplate, 0, len(parents))
		for _, p := range parents {
			children := make([]repository.AccountBookCategoryChildTemplate, 0, len(p.Childs))
			for _, child := range p.Childs {
				children = append(children, repository.AccountBookCategoryChildTemplate{
					Name:     strings.TrimSpace(child.Name),
					IconPath: strings.TrimSpace(child.IconPath),
				})
			}
			items = append(items, repository.AccountBookCategoryTemplate{
				Name:     strings.TrimSpace(p.Name),
				IconPath: strings.TrimSpace(p.IconPath),
				Childs:   children,
			})
		}
		res[cleanType] = items
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func normalizeAssets(src []AccountBookAssetInput) []repository.AccountBookAssetTemplate {
	if len(src) == 0 {
		return nil
	}
	res := make([]repository.AccountBookAssetTemplate, 0, len(src))
	for _, asset := range src {
		children := make([]repository.AccountBookAssetChildTemplate, 0, len(asset.Childs))
		for _, child := range asset.Childs {
			children = append(children, repository.AccountBookAssetChildTemplate{
				Name:     strings.TrimSpace(child.Name),
				IconPath: strings.TrimSpace(child.IconPath),
			})
		}
		res = append(res, repository.AccountBookAssetTemplate{
			Name:     strings.TrimSpace(asset.Name),
			IconPath: strings.TrimSpace(asset.IconPath),
			Type:     strings.TrimSpace(asset.Type),
			Childs:   children,
		})
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func fromPresetCategories(src map[string][]accountBookPresetCategory) map[string][]repository.AccountBookCategoryTemplate {
	if len(src) == 0 {
		return nil
	}
	res := make(map[string][]repository.AccountBookCategoryTemplate, len(src))
	for statementType, parents := range src {
		items := make([]repository.AccountBookCategoryTemplate, 0, len(parents))
		for _, p := range parents {
			children := make([]repository.AccountBookCategoryChildTemplate, 0, len(p.Childs))
			for _, child := range p.Childs {
				children = append(children, repository.AccountBookCategoryChildTemplate{
					Name:     child.Name,
					IconPath: child.IconPath,
				})
			}
			items = append(items, repository.AccountBookCategoryTemplate{
				Name:     p.Name,
				IconPath: p.IconPath,
				Childs:   children,
			})
		}
		res[statementType] = items
	}
	return res
}

func fromPresetAssets(src []accountBookPresetAsset) []repository.AccountBookAssetTemplate {
	if len(src) == 0 {
		return nil
	}
	res := make([]repository.AccountBookAssetTemplate, 0, len(src))
	for _, asset := range src {
		children := make([]repository.AccountBookAssetChildTemplate, 0, len(asset.Childs))
		for _, child := range asset.Childs {
			children = append(children, repository.AccountBookAssetChildTemplate{
				Name:     child.Name,
				IconPath: child.IconPath,
			})
		}
		res = append(res, repository.AccountBookAssetTemplate{
			Name:     asset.Name,
			IconPath: asset.IconPath,
			Type:     asset.Type,
			Childs:   children,
		})
	}
	return res
}
