package statement

import "time"

type ListInput struct {
	UserID            int64
	AccountBookID     int64
	StartDate         *time.Time
	EndDate           *time.Time
	ParentCategoryIDs []int64
	ExceptIDs         []int64
	OrderBy           string
	Limit             int
	Offset            int
}

type WriteInput struct {
	StatementID   int64
	UserID        int64
	AccountBookID int64

	Type         string
	Amount       float64
	Description  string
	Mood         string
	CategoryID   int64
	AssetID      int64
	FromAssetID  int64
	ToAssetID    int64
	PayeeID      int64
	TargetObject string

	Location string
	Nation   string
	Province string
	City     string
	District string
	Street   string

	Date string
	Time string
}

type PatchInput struct {
	Type         *string
	Amount       *float64
	Description  *string
	Mood         *string
	CategoryID   *int64
	AssetID      *int64
	FromAssetID  *int64
	ToAssetID    *int64
	PayeeID      *int64
	TargetObject *string
	Location     *string
	Nation       *string
	Province     *string
	City         *string
	District     *string
	Street       *string
	Date         *string
	Time         *string
}

type UpdateInput struct {
	StatementID   int64
	UserID        int64
	AccountBookID int64
	Patch         PatchInput
}

type Payee struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type BaseItem struct {
	ID           int64   `json:"id"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	Title        string  `json:"title"`
	TargetObject string  `json:"target_object"`
	Mood         string  `json:"mood"`
	Money        string  `json:"money"`
	Category     string  `json:"category"`
	IconPath     string  `json:"icon_path"`
	Asset        string  `json:"asset"`
	Date         string  `json:"date"`
	Time         string  `json:"time"`
	TimeStr      string  `json:"timeStr"`
	Week         string  `json:"week"`
	Payee        Payee   `json:"payee"`
	Remark       string  `json:"remark"`
	CategoryID   int64   `json:"category_id"`
	AssetID      int64   `json:"asset_id"`
}

type ListItem struct {
	BaseItem
	Location  string    `json:"location"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	Street    string    `json:"street"`
	MonthDay  string    `json:"month_day"`
	HasPic    bool      `json:"has_pic"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DetailItem struct {
	BaseItem
	Location    string        `json:"location"`
	Province    string        `json:"province"`
	City        string        `json:"city"`
	Street      string        `json:"street"`
	MonthDay    string        `json:"month_day"`
	HasPic      bool          `json:"has_pic"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	UploadFiles []interface{} `json:"upload_files"`
}
