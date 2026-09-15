package types

type WalletTimelineItem struct {
	ExpendAmount float64 `json:"expend_amount"`
	IncomeAmount float64 `json:"income_amount"`
	Surplus      float64 `json:"surplus"`
	Year         int     `json:"year"`
	Month        int     `json:"month"`
	Hidden       int     `json:"hidden"`
}

type WalletTimelineResponse struct {
	Status int                  `json:"status"`
	Data   []WalletTimelineItem `json:"data"`
}

type WalletInformation struct {
	Name          string  `json:"name"`
	Income        string  `json:"income"`
	Expend        string  `json:"expend"`
	Surplus       string  `json:"surplus"`
	SourceSurplus float64 `json:"source_surplus"`
}

type WalletTypeAmountItem struct {
	CategoryID int64  `json:"category_id"`
	Name       string `json:"name"`
	Amount     string `json:"amount"`
}

type WalletTypeSummary struct {
	Name   string                 `json:"name"`
	Amount string                 `json:"amount"`
	Childs []WalletTypeAmountItem `json:"childs"`
}

type WalletChildAsset struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	IconPath string `json:"icon_path"`
}

type WalletParent struct {
	Name   string             `json:"name"`
	Amount string             `json:"amount"`
	Childs []WalletChildAsset `json:"childs"`
}

type WalletHeader struct {
	TotalAsset     string `json:"total_asset"`
	NetWorth       string `json:"net_worth"`
	TotalLiability string `json:"total_liability"`
}

type WalletResponse struct {
	Header        WalletHeader      `json:"header"`
	List          []WalletParent    `json:"list"`
	AmountVisible bool              `json:"amount_visible"`
	Receivables   WalletTypeSummary `json:"receivables"`
	Payables      WalletTypeSummary `json:"payables"`
}
