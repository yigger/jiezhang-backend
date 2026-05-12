package dto

type StatementWritePayload struct {
	Type         string `json:"type"`
	Amount       string `json:"amount"`
	Description  string `json:"description"`
	Mood         string `json:"mood"`
	CategoryID   int64  `json:"category_id"`
	AssetID      int64  `json:"asset_id"`
	FromAssetID  int64  `json:"from_asset_id"`
	ToAssetID    int64  `json:"to_asset_id"`
	PayeeID      int64  `json:"payee_id"`
	TargetObject string `json:"target_object"`
	Location     string `json:"location"`
	Nation       string `json:"nation"`
	Province     string `json:"province"`
	City         string `json:"city"`
	District     string `json:"district"`
	Street       string `json:"street"`
	Date         string `json:"date"`
	Time         string `json:"time"`
}

type StatementPatchPayload struct {
	Type         *string `json:"type"`
	Amount       *string `json:"amount"`
	Description  *string `json:"description"`
	Mood         *string `json:"mood"`
	CategoryID   *int64  `json:"category_id"`
	AssetID      *int64  `json:"asset_id"`
	FromAssetID  *int64  `json:"from_asset_id"`
	ToAssetID    *int64  `json:"to_asset_id"`
	PayeeID      *int64  `json:"payee_id"`
	TargetObject *string `json:"target_object"`
	Location     *string `json:"location"`
	Nation       *string `json:"nation"`
	Province     *string `json:"province"`
	City         *string `json:"city"`
	District     *string `json:"district"`
	Street       *string `json:"street"`
	Date         *string `json:"date"`
	Time         *string `json:"time"`
}

type StatementWriteRequest struct {
	Statement StatementWritePayload `json:"statement"`
}

type StatementPatchRequest struct {
	Statement StatementPatchPayload `json:"statement"`
}
