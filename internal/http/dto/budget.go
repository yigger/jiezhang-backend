package dto

type BudgetUpdateRequest struct {
	Type       string `json:"type"`
	Amount     string `json:"amount"`
	CategoryID int64  `json:"category_id"`
}
