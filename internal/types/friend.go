package types

import (
	"time"
)

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

type InviteInformationBook struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type InviteInformationItem struct {
	InviteUser  FriendUserItem        `json:"invite_user"`
	AccountBook InviteInformationBook `json:"account_book"`
	RoleName    string                `json:"role_name"`
}

type FriendAuthority struct {
	ChangeRole bool `json:"change_role"`
	Remove     bool `json:"remove"`
}

type FriendUserItem struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar_path"`
}

type FriendCollaboratorItem struct {
	ID        int64          `json:"id"`
	Role      string         `json:"role"`
	RoleName  string         `json:"role_name"`
	Remark    string         `json:"remark"`
	User      FriendUserItem `json:"user"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type FriendListResponse struct {
	Collaborators []FriendCollaboratorItem `json:"collaborators"`
	Owner         FriendUserItem           `json:"owner"`
	Authority     FriendAuthority          `json:"authority"`
}
