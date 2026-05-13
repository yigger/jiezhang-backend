package dto

type AssetWriteRequest struct {
	Wallet struct {
		Name     string `json:"name"`
		Amount   string `json:"amount"`
		ParentID int64  `json:"parent_id"`
		IconPath string `json:"icon_path"`
		Remark   string `json:"remark"`
		Type     string `json:"type"`
	} `json:"wallet"`
}

type AssetSurplusRequest struct {
	AssetID int64  `json:"asset_id"`
	Amount  string `json:"amount"`
}
