package types

type StatementExportResult struct {
	Rows []ExportRow `json:"rows"`
}

type StatementExportCheckResult struct {
	TodayCount int `json:"today_count"`
}

type ExportRow struct {
	Category       string  `json:"category"`
	ParentCategory string  `json:"parent_category"`
	Type           string  `json:"type"`
	TypeName       string  `json:"type_name"`
	Asset          string  `json:"asset"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}
