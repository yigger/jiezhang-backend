package dto

type UserUpdateRequest struct {
	User UserUpdatePayload `json:"user"`
}

type UserUpdatePayload struct {
	ThemeID          *int64  `json:"theme_id"`
	Country          *string `json:"country"`
	City             *string `json:"city"`
	Gender           *int    `json:"gender"`
	Language         *string `json:"language"`
	Province         *string `json:"province"`
	BGAvatarID       *int64  `json:"bg_avatar_id"`
	HiddenAssetMoney *bool   `json:"hidden_asset_money"`
	AvatarURL        *string `json:"avatar_url"`
	Nickname         *string `json:"nickname"`
	BGAvatar         *string `json:"bg_avatar"`
}

type UserScanLoginRequest struct {
	QRCode string `json:"qr_code"`
}
