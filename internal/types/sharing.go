package types

type StatementListByTokenResult struct {
	Data       []StatementListItem     `json:"data"`
	DateRange  StatementDateRangeItem  `json:"date_range"`
	SharedUser StatementSharedUserItem `json:"shared_user"`
}

type StatementSharedUserItem struct {
	Nickname   string `json:"nickname"`
	AvatarPath string `json:"avatar_path"`
}

type StatementDateRangeItem struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
