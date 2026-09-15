package types

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

type UserProfile struct {
	ID               int64  `json:"id"`
	ThemeID          int64  `json:"theme_id"`
	AvatarURL        string `json:"avatar_url"`
	Nickname         string `json:"nickname"`
	Persist          int64  `json:"persist"`
	StatementsCount  int64  `json:"sts_count"`
	Email            string `json:"email"`
	Remind           bool   `json:"remind"`
	HiddenAssetMoney bool   `json:"hidden_asset_money"`
}

type UserCreateRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Email string `json:"email" binding:"required,email"`
}
