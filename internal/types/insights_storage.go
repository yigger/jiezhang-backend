package types

type InsightProjectInput struct {
	Icon           *string `json:"icon"`
	Color          *string `json:"color"`
	ParticipantIDs []int64 `json:"participant_ids"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	BudgetCents    int64   `json:"budget_cents"`
	Archived       bool    `json:"archived"`
}
type InsightFixedCostInput struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	AmountCents    int64  `json:"amount_cents"`
	CategoryID     int64  `json:"category_id"`
	AssetID        int64  `json:"asset_id"`
	IntervalMonths int    `json:"interval_months"`
	DueDay         int    `json:"due_day"`
	CandidateKey   string `json:"candidate_key"`
	Active         bool   `json:"active"`
}
type InsightAllocation struct {
	MemberID    int64 `json:"member_id"`
	AmountCents int64 `json:"amount_cents"`
}
type InsightAnnotationInput struct {
	ConsumerID  *int64              `json:"consumer_id"`
	StatementID int64               `json:"statement_id"`
	ProjectID   *int64              `json:"project_id"`
	PayerID     *int64              `json:"payer_id"`
	FixedCostID *int64              `json:"fixed_cost_id"`
	Allocations []InsightAllocation `json:"allocations"`
}
type InsightAssetBalance struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	AmountCents int64  `json:"amount_cents"`
}

type InsightProjectSummary struct {
	InsightProjectInput
	CreatorID int64          `json:"creator_id"`
	CanEdit   bool           `json:"can_edit"`
	Summary   InsightGroup   `json:"summary"`
	Payers    []InsightGroup `json:"payers"`
}
type InsightFixedCostSummary struct {
	NextRunDate *string `json:"next_run_date"`
	InsightFixedCostInput
	CreatorID    int64 `json:"creator_id"`
	CanEdit      bool  `json:"can_edit"`
	MonthlyCents int64 `json:"monthly_cents"`
	AnnualCents  int64 `json:"annual_cents"`
}
type InsightMemberOption struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type InsightAnnotation struct {
	InsightAnnotationInput
	CanEdit    bool `json:"can_edit"`
	SplitValid bool `json:"split_valid"`
}
type InsightAssetChange struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	BeforeCents   *int64 `json:"before_cents"`
	AfterCents    *int64 `json:"after_cents"`
	NetDeltaCents int64  `json:"net_delta_cents"`
	Reason        string `json:"reason"`
}
type InsightPortfolioPoint struct {
	ID               int64                 `json:"id"`
	Date             string                `json:"date"`
	Note             string                `json:"note"`
	AssetsCents      int64                 `json:"assets_cents"`
	LiabilitiesCents int64                 `json:"liabilities_cents"`
	NetCents         int64                 `json:"net_cents"`
	DeltaCents       *int64                `json:"delta_cents"`
	Balances         []InsightAssetBalance `json:"balances"`
	Changes          []InsightAssetChange  `json:"changes"`
}
type InsightWorkspace struct {
	EditableStatementIDs []int64                   `json:"editable_statement_ids"`
	Projects             []InsightProjectSummary   `json:"projects"`
	FixedCosts           []InsightFixedCostSummary `json:"fixed_costs"`
	Annotations          []InsightAnnotation       `json:"annotations"`
	Members              []InsightMemberOption     `json:"members"`
	Payers               []InsightGroup            `json:"payers"`
	Burdens              []InsightGroup            `json:"burdens"`
	UnknownPayerCents    int64                     `json:"unknown_payer_cents"`
	UnallocatedCents     int64                     `json:"unallocated_cents"`
	InvalidSplitCount    int                       `json:"invalid_split_count"`
	Portfolio            []InsightPortfolioPoint   `json:"portfolio"`
	MonthlyFixedCents    int64                     `json:"monthly_fixed_cents"`
	AnnualFixedCents     int64                     `json:"annual_fixed_cents"`
}
type InsightSnapshotInput struct {
	Note string `json:"note"`
}
type InsightSavedResponse struct {
	ID int64 `json:"id"`
}
type InsightMerchantAliasInput struct {
	PayeeIDs []int64 `json:"payee_ids"`
	Name     string  `json:"name"`
}
