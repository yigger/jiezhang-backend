package types

type BudgetUpdateRequest struct {
	Type       string         `json:"type"`
	Amount     FlexibleAmount `json:"amount"`
	CategoryID *int64         `json:"category_id"`
}

type BudgetCategoryDetail struct {
	Root   BudgetRoot        `json:"root"`
	Childs []BudgetChildItem `json:"childs"`
}

type BudgetRoot struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	IconPath       string  `json:"icon_path"`
	SourceAmount   float64 `json:"source_amount"`
	UsedAmount     string  `json:"used_amount"`
	Amount         string  `json:"amount"`
	Surplus        string  `json:"surplus"`
	UsePercent     int     `json:"use_percent"`
	SurplusPercent int     `json:"surplus_percent"`
}

type BudgetChildItem struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	IconPath       string  `json:"icon_path"`
	SourceAmount   float64 `json:"source_amount"`
	Amount         string  `json:"amount"`
	Surplus        string  `json:"surplus"`
	UsePercent     int     `json:"use_percent"`
	UsedAmount     float64 `json:"used_amount"`
	SurplusPercent int     `json:"surplus_percent"`
}

type BudgetParentItem struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	IconPath       string  `json:"icon_path"`
	SourceAmount   float64 `json:"source_amount"`
	Amount         string  `json:"amount"`
	Surplus        string  `json:"surplus"`
	UsedAmount     float64 `json:"used_amount"`
	UsePercent     int     `json:"use_percent"`
	SurplusPercent int     `json:"surplus_percent"`
}

type BudgetSummary struct {
	SourceAmount float64 `json:"source_amount"`
	Amount       string  `json:"amount"`
	Used         float64 `json:"used"`
	Surplus      float64 `json:"surplus"`
}
