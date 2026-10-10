package types

import (
	"time"
)

type ShareKeyRequest struct {
	StartDate            string `json:"start_date"`
	EndDate              string `json:"end_date"`
	CategoryIDs          string `json:"category_ids"`
	ExceptedStatementIDs string `json:"exceptedStatementIds"`
	ExceptStatementIDs   string `json:"except_statement_ids"`
}

type AvatarDeleteRequest struct {
	AvatarID int64 `json:"avatar_id"`
}

type ShareKeyData struct {
	ShareKey string `json:"share_key"`
}

type StatementWritePayload struct {
	ProjectID    int64          `json:"project_id"`
	ConsumerID   int64          `json:"consumer_id"`
	Type         string         `json:"type"`
	Amount       FlexibleAmount `json:"amount"`
	Description  string         `json:"description"`
	Mood         string         `json:"mood"`
	CategoryID   int64          `json:"category_id"`
	AssetID      int64          `json:"asset_id"`
	FromAssetID  int64          `json:"from_asset_id"`
	ToAssetID    int64          `json:"to_asset_id"`
	PayeeID      int64          `json:"payee_id"`
	TargetObject string         `json:"target_object"`
	Location     string         `json:"location"`
	Nation       string         `json:"nation"`
	Province     string         `json:"province"`
	City         string         `json:"city"`
	District     string         `json:"district"`
	Street       string         `json:"street"`
	Date         string         `json:"date"`
	Time         string         `json:"time"`
}

type StatementPatchPayload struct {
	Type         *string         `json:"type"`
	Amount       *FlexibleAmount `json:"amount"`
	Description  *string         `json:"description"`
	Mood         *string         `json:"mood"`
	CategoryID   *int64          `json:"category_id"`
	AssetID      *int64          `json:"asset_id"`
	FromAssetID  *int64          `json:"from_asset_id"`
	ToAssetID    *int64          `json:"to_asset_id"`
	PayeeID      *int64          `json:"payee_id"`
	TargetObject *string         `json:"target_object"`
	Location     *string         `json:"location"`
	Nation       *string         `json:"nation"`
	Province     *string         `json:"province"`
	City         *string         `json:"city"`
	District     *string         `json:"district"`
	Street       *string         `json:"street"`
	Date         *string         `json:"date"`
	Time         *string         `json:"time"`
}

type StatementWriteRequest struct {
	Statement StatementWritePayload `json:"statement"`
}

type StatementPatchRequest struct {
	Statement StatementPatchPayload `json:"statement"`
}

type StatementDetailItem struct {
	StatementBaseItem
	AmountNumber  float64                   `json:"amount_number"`
	Location      string                    `json:"location"`
	Province      string                    `json:"province"`
	City          string                    `json:"city"`
	Street        string                    `json:"street"`
	MonthDay      string                    `json:"month_day"`
	HasPic        bool                      `json:"has_pic"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
	UploadFiles   []StatementUploadFileItem `json:"upload_files"`
	TargetAssetID int64                     `json:"target_asset_id"`
	Residue       string                    `json:"residue"`
	TargetAsset   *StatementTargetAssetInfo `json:"target_asset,omitempty"`
	CanEdit       bool                      `json:"can_edit"`
}

type StatementUploadFileItem struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

type StatementTargetAssetInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementListItem struct {
	StatementBaseItem
	Location  string    `json:"location"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	Street    string    `json:"street"`
	MonthDay  string    `json:"month_day"`
	HasPic    bool      `json:"has_pic"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StatementBaseItem struct {
	ID           int64          `json:"id"`
	Type         string         `json:"type"`
	Amount       float64        `json:"amount"`
	Description  string         `json:"description"`
	Title        string         `json:"title"`
	TargetObject string         `json:"target_object"`
	Mood         string         `json:"mood"`
	Money        string         `json:"money"`
	Category     string         `json:"category"`
	IconPath     string         `json:"icon_path"`
	Asset        string         `json:"asset"`
	Date         string         `json:"date"`
	Time         string         `json:"time"`
	TimeStr      string         `json:"timeStr"`
	Week         string         `json:"week"`
	Payee        StatementPayee `json:"payee"`
	Remark       string         `json:"remark"`
	CategoryID   int64          `json:"category_id"`
	AssetID      int64          `json:"asset_id"`
}

type StatementPayee struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementDefaultCategoryAssetItem struct {
	CategoryName string `json:"category_name"`
	AssetName    string `json:"asset_name"`
	CategoryID   int64  `json:"category_id"`
	AssetID      int64  `json:"asset_id"`
}

type StatementAssetsResult struct {
	Frequent   []StatementFrequentAssetItem `json:"frequent"`
	Categories []StatementAssetTreeItem     `json:"categories"`
}

type StatementAssetTreeItem struct {
	ID       int64                     `json:"id"`
	Name     string                    `json:"name"`
	IconPath string                    `json:"icon_path"`
	Childs   []StatementAssetChildItem `json:"childs"`
}

type StatementAssetChildItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type StatementAssetParentItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementFrequentAssetItem struct {
	ID       int64                     `json:"id"`
	Name     string                    `json:"name"`
	IconPath string                    `json:"icon_path"`
	Parent   *StatementAssetParentItem `json:"parent"`
}

type StatementCategoriesResult struct {
	Frequent   []StatementFrequentCategoryItem `json:"frequent"`
	Categories []StatementCategoryTreeItem     `json:"categories"`
}

type StatementCategoryTreeItem struct {
	ID       int64                        `json:"id"`
	Name     string                       `json:"name"`
	IconPath string                       `json:"icon_path"`
	Childs   []StatementCategoryChildItem `json:"childs"`
}

type StatementCategoryChildItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type StatementCategoryParentItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementFrequentCategoryItem struct {
	ID       int64                        `json:"id"`
	Name     string                       `json:"name"`
	IconPath string                       `json:"icon_path"`
	Parent   *StatementCategoryParentItem `json:"parent"`
}

type StatementImagesResult struct {
	AvatarTimeline []StatementImageYearGroup `json:"avatar_timeline"`
	Avatars        []string                  `json:"avatars"`
}

type StatementImageYearGroup struct {
	Year int                        `json:"year"`
	Data []StatementImageMonthGroup `json:"data"`
}

type StatementImageMonthGroup struct {
	Month int                  `json:"month"`
	Data  []StatementImageItem `json:"data"`
}

type StatementImageItem struct {
	StatementID int64  `json:"statement_id"`
	AvatarID    int64  `json:"avatar_id"`
	Path        string `json:"path"`
}
