package types

type CategoryWriteRequest struct {
	Category struct {
		Name     string `json:"name"`
		ParentID int64  `json:"parent_id"`
		IconPath string `json:"icon_path"`
		Type     string `json:"type"`
	} `json:"category"`
}

type CategoryStatementChild struct {
	ID          int64  `json:"id"`
	Day         int    `json:"day"`
	Week        string `json:"week"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	IconPath    string `json:"icon_path"`
	Description string `json:"description"`
	Money       string `json:"money"`
	TimeStr     string `json:"timeStr"`
	Asset       string `json:"asset"`
}

type CategoryStatementsMonthItem struct {
	Year   int                      `json:"year"`
	Month  int                      `json:"month"`
	Childs []CategoryStatementChild `json:"childs"`
}

type CategoryShowResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Order      int    `json:"order"`
	IconPath   string `json:"icon_path"`
	ParentID   int64  `json:"parent_id"`
	Type       string `json:"type"`
	ParentName string `json:"parent_name"`
	IconURL    string `json:"icon_url"`
}

type CategoryListResponse struct {
	Header     CategoryHeader `json:"header"`
	Categories []CategoryItem `json:"categories"`
}

type CategoryHeader struct {
	Month      string `json:"month"`
	Year       string `json:"year"`
	All        string `json:"all"`
	ParentName string `json:"parent_name,omitempty"`
}

type CategoryItem struct {
	ID       int64          `json:"id"`
	Name     string         `json:"name"`
	Order    int            `json:"order"`
	IconPath string         `json:"icon_path"`
	ParentID int64          `json:"parent_id"`
	Type     string         `json:"type"`
	Amount   string         `json:"amount,omitempty"`
	IconURL  string         `json:"icon_url,omitempty"`
	Childs   []CategoryItem `json:"childs,omitempty"`
}
