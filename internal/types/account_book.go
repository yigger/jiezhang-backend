package types

type AccountBookCreateRequest struct {
	Name        string                               `json:"name"`
	Description string                               `json:"description"`
	AccountType string                               `json:"account_type"`
	Categories  map[string][]AccountBookCategoryItem `json:"categories"`
	Assets      []AccountBookAssetItem               `json:"assets"`
}

type AccountBookUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AccountType struct {
		ID string `json:"id"`
	} `json:"account_type"`
}

type AccountBookCategoryItem struct {
	Name     string                 `json:"name"`
	IconPath string                 `json:"icon_path"`
	Childs   []AccountBookChildItem `json:"childs"`
}

type AccountBookAssetItem struct {
	Name     string                 `json:"name"`
	IconPath string                 `json:"icon_path"`
	Type     string                 `json:"type"`
	Childs   []AccountBookChildItem `json:"childs"`
}

type AccountBookChildItem struct {
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type AccountBookPresetChild struct {
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type AccountBookPresetAsset struct {
	Name     string                   `json:"name"`
	IconPath string                   `json:"icon_path"`
	Type     string                   `json:"type"`
	Childs   []AccountBookPresetChild `json:"childs"`
}

type AccountBookPresetCategory struct {
	Name     string                   `json:"name"`
	IconPath string                   `json:"icon_path"`
	Childs   []AccountBookPresetChild `json:"childs"`
}

type AccountBookPreset struct {
	Name       string                                 `json:"name"`
	Categories map[string][]AccountBookPresetCategory `json:"categories"`
	Assets     []AccountBookPresetAsset               `json:"assets"`
}

type AccountBookDetailItem struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	AccountType AccountBookTypeInfo `json:"account_type"`
}

type AccountBookTypeInfo struct {
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

type AccountBookTypeItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
