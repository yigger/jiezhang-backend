package types

// InsightAmounts uses integer cents; balance means income minus ordinary expenditure.
type InsightAmounts struct {
	IncomeCents  int64 `json:"income_cents"`
	ExpendCents  int64 `json:"expend_cents"`
	BalanceCents int64 `json:"balance_cents"`
	Count        int   `json:"count"`
}
type InsightMonth struct {
	InsightAmounts
	Month    int  `json:"month"`
	Started  bool `json:"started"`
	Complete bool `json:"complete"`
}
type InsightCategory struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	AmountCents int64  `json:"amount_cents"`
	Count       int    `json:"count"`
}
type InsightGroup struct {
	Key  string  `json:"key"`
	Name string  `json:"name"`
	IDs  []int64 `json:"ids"`
	InsightAmounts
	AverageCents int64             `json:"average_cents"`
	LastDate     string            `json:"last_date"`
	Categories   []InsightCategory `json:"categories"`
	StatementIDs []int64           `json:"statement_ids"`
}
type InsightLargeExpense struct {
	ID          int64  `json:"id"`
	AmountCents int64  `json:"amount_cents"`
	Date        string `json:"date"`
	Category    string `json:"category"`
	Description string `json:"description"`
}
type InsightRecurringCandidate struct {
	Key          string  `json:"key"`
	Name         string  `json:"name"`
	AmountCents  int64   `json:"amount_cents"`
	AnnualCents  int64   `json:"annual_cents"`
	Months       int     `json:"months"`
	DueDay       int     `json:"due_day"`
	CategoryID   int64   `json:"category_id"`
	AssetID      int64   `json:"asset_id"`
	StatementIDs []int64 `json:"statement_ids"`
}
type InsightReport struct {
	Year      int    `json:"year"`
	AsOf      string `json:"as_of"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	InsightAmounts
	Months              []InsightMonth              `json:"months"`
	Merchants           []InsightGroup              `json:"merchants"`
	Members             []InsightGroup              `json:"members"`
	LargeExpenses       []InsightLargeExpense       `json:"large_expenses"`
	RecurringCandidates []InsightRecurringCandidate `json:"recurring_candidates"`
}
