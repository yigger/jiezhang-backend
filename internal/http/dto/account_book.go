package dto

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
