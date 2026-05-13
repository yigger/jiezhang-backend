package dto

type FriendInviteRequest struct {
	AccountBookID int64  `json:"account_book_id"`
	Role          string `json:"role"`
}

type FriendInviteInformationRequest struct {
	InviteToken string `form:"invite_token" json:"invite_token"`
}

type FriendAcceptApplyRequest struct {
	InviteToken string `json:"invite_token"`
	Nickname    string `json:"nickname"`
}

type FriendUpdateRequest struct {
	AccountBookID int64   `json:"account_book_id"`
	Role          *string `json:"role"`
	Remark        *string `json:"remark"`
}

type FriendRemoveRequest struct {
	AccountBookID int64 `form:"account_book_id" json:"account_book_id"`
}
