package friend

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	"github.com/yigger/jiezhang-backend/internal/types"
)

var (
	ErrFriendInvalidInput      = errors.New("friend invalid input")
	ErrFriendPermissionDenied  = errors.New("friend permission denied")
	ErrFriendInviteToken       = errors.New("friend invite token invalid")
	ErrFriendInviteExpired     = errors.New("friend invite expired")
	ErrFriendAlreadyMember     = errors.New("friend already member")
	ErrFriendCannotEditSelf    = errors.New("friend cannot edit self")
	ErrFriendCollaboratorNotIn = errors.New("friend collaborator not in account book")
)

var friendRoleNameMap = map[string]string{
	"viewer": "观察者",
	"member": "普通成员",
	"admin":  "管理员",
	"owner":  "拥有者",
}

type FriendService struct {
	repo       repo.FriendRepository
	urlBuilder repo.URLBuilder
	codec      repo.TokenCodec
}

func NewFriendService(repo repo.FriendRepository, urlBuilder repo.URLBuilder, codec repo.TokenCodec) FriendService {
	return FriendService{
		repo:       repo,
		urlBuilder: urlBuilder,
		codec:      codec,
	}
}

type FriendInviteInput struct {
	AccountBookID int64
	UserID        int64
	Role          string
}

type FriendUpdateInput struct {
	AccountBookID  int64
	OperatorUserID int64
	CollaboratorID int64
	Role           *string
	Remark         *string
}

type FriendRemoveInput struct {
	AccountBookID  int64
	OperatorUserID int64
	CollaboratorID int64
}

type FriendAcceptInput struct {
	UserID      int64
	InviteToken string
	Nickname    string
}

type FriendInviteTokenPayload struct {
	InviteUserID  int64  `json:"invite_user_id"`
	AccountBookID int64  `json:"account_book_id"`
	Role          string `json:"role"`
	ExpireAtUnix  int64  `json:"expire_at_unix"`
}

func (s FriendService) List(ctx context.Context, accountBookID int64, currentUserID int64) (types.FriendListResponse, error) {
	accountBook, err := s.repo.FindAccessibleAccountBookByID(ctx, currentUserID, accountBookID)
	if err != nil {
		return types.FriendListResponse{}, err
	}

	collaborators, err := s.repo.ListCollaborators(ctx, accountBook.ID)
	if err != nil {
		return types.FriendListResponse{}, err
	}
	operator, err := s.repo.FindCollaboratorByUserID(ctx, accountBook.ID, currentUserID)
	if err != nil {
		return types.FriendListResponse{}, err
	}

	ownerRow, err := s.repo.FindUserByID(ctx, accountBook.UserID)
	if err != nil {
		return types.FriendListResponse{}, err
	}
	owner := helperservice.ToUserContext(ownerRow)

	userIDs := make([]int64, 0, len(collaborators))
	for _, c := range collaborators {
		userIDs = append(userIDs, c.UserID)
	}
	users, err := s.repo.ListUsersByIDs(ctx, userIDs)
	if err != nil {
		return types.FriendListResponse{}, err
	}
	usersByID := make(map[int64]types.UserContext, len(users))
	for _, u := range users {
		usersByID[u.ID] = helperservice.ToUserContext(u)
	}
	items := make([]types.FriendCollaboratorItem, 0, len(collaborators))
	for _, c := range collaborators {
		items = append(items, types.FriendCollaboratorItem{
			ID:       c.ID,
			Role:     c.Role,
			RoleName: friendRoleName(c.Role),
			Remark:   c.Remark,
			User: types.FriendUserItem{
				ID:       c.UserID,
				Nickname: usersByID[c.UserID].Nickname,
				Avatar:   s.buildPublicURL(usersByID[c.UserID].AvatarUrl),
			},
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}

	canAdmin := isRoleCanAdmin(operator.Role)
	return types.FriendListResponse{
		Collaborators: items,
		Owner: types.FriendUserItem{
			ID:       owner.ID,
			Nickname: owner.Nickname,
			Avatar:   s.buildPublicURL(owner.AvatarUrl),
		},
		Authority: types.FriendAuthority{
			ChangeRole: canAdmin,
			Remove:     canAdmin,
		},
	}, nil
}

func (s FriendService) Invite(ctx context.Context, input FriendInviteInput) (string, error) {
	role := strings.TrimSpace(input.Role)
	if input.AccountBookID <= 0 || input.UserID <= 0 || role == "" {
		return "", ErrFriendInvalidInput
	}
	if !isSupportedCollaboratorRole(role) {
		return "", ErrFriendInvalidInput
	}

	accountBook, err := s.repo.FindAccessibleAccountBookByID(ctx, input.UserID, input.AccountBookID)
	if err != nil {
		return "", err
	}
	canAdmin, err := s.repo.CanAdmin(ctx, accountBook.ID, input.UserID)
	if err != nil {
		return "", err
	}
	if !canAdmin {
		return "", ErrFriendPermissionDenied
	}

	payload := FriendInviteTokenPayload{
		InviteUserID:  input.UserID,
		AccountBookID: accountBook.ID,
		Role:          role,
		ExpireAtUnix:  time.Now().Add(24 * time.Hour).Unix(),
	}
	return s.encryptInvitePayload(payload)
}

func (s FriendService) InviteInformation(ctx context.Context, token string) (types.InviteInformationItem, error) {
	payload, err := s.decryptAndValidateInviteToken(ctx, strings.TrimSpace(token))
	if err != nil {
		return types.InviteInformationItem{}, err
	}

	inviteUserRow, err := s.repo.FindUserByID(ctx, payload.InviteUserID)
	if err != nil {
		return types.InviteInformationItem{}, err
	}
	inviteUser := helperservice.ToUserContext(inviteUserRow)

	accountBook, err := s.repo.FindAccountBookByID(ctx, payload.AccountBookID)
	if err != nil {
		return types.InviteInformationItem{}, err
	}

	return types.InviteInformationItem{
		InviteUser: types.FriendUserItem{
			ID:       inviteUser.ID,
			Nickname: inviteUser.Nickname,
			Avatar:   s.buildPublicURL(inviteUser.AvatarUrl),
		},
		AccountBook: types.InviteInformationBook{
			ID:   accountBook.ID,
			Name: accountBook.Name,
		},
		RoleName: friendRoleName(payload.Role),
	}, nil
}

func (s FriendService) AcceptApply(ctx context.Context, input FriendAcceptInput) error {
	if input.UserID <= 0 || strings.TrimSpace(input.InviteToken) == "" {
		return ErrFriendInvalidInput
	}
	payload, err := s.decryptAndValidateInviteToken(ctx, strings.TrimSpace(input.InviteToken))
	if err != nil {
		return err
	}

	if payload.InviteUserID == input.UserID {
		return ErrFriendAlreadyMember
	}
	if _, err := s.repo.FindCollaboratorByUserID(ctx, payload.AccountBookID, input.UserID); err == nil {
		return ErrFriendAlreadyMember
	} else if !errors.Is(err, repo.ErrFriendCollaboratorNotFound) {
		return err
	}

	if err := s.repo.CreateCollaborator(ctx, tablemodel.AccountBookCollaborator{
		AccountBookID: payload.AccountBookID,
		UserID:        input.UserID,
		Role:          payload.Role,
		Remark:        strings.TrimSpace(input.Nickname),
	}); err != nil {
		if errors.Is(err, repo.ErrFriendCollaboratorExists) {
			return ErrFriendAlreadyMember
		}
		return err
	}
	return nil
}

func (s FriendService) Update(ctx context.Context, input FriendUpdateInput) error {
	if input.AccountBookID <= 0 || input.OperatorUserID <= 0 || input.CollaboratorID <= 0 {
		return ErrFriendInvalidInput
	}
	collaborator, err := s.repo.FindCollaboratorByID(ctx, input.AccountBookID, input.CollaboratorID)
	if err != nil {
		return err
	}

	if input.Role != nil {
		if collaborator.UserID == input.OperatorUserID {
			return ErrFriendCannotEditSelf
		}
		role := strings.TrimSpace(*input.Role)
		if !isSupportedCollaboratorRole(role) {
			return ErrFriendInvalidInput
		}
		canAdmin, adminErr := s.repo.CanAdmin(ctx, input.AccountBookID, input.OperatorUserID)
		if adminErr != nil {
			return adminErr
		}
		if !canAdmin {
			return ErrFriendPermissionDenied
		}
	}

	update := repo.FriendCollaboratorUpdateRecord{
		Role:   input.Role,
		Remark: input.Remark,
	}
	return s.repo.UpdateCollaborator(ctx, input.AccountBookID, input.CollaboratorID, update)
}

func (s FriendService) Remove(ctx context.Context, input FriendRemoveInput) error {
	if input.AccountBookID <= 0 || input.OperatorUserID <= 0 || input.CollaboratorID <= 0 {
		return ErrFriendInvalidInput
	}
	collaborator, err := s.repo.FindCollaboratorByID(ctx, input.AccountBookID, input.CollaboratorID)
	if err != nil {
		return err
	}

	// self quit
	if collaborator.UserID != input.OperatorUserID {
		canAdmin, adminErr := s.repo.CanAdmin(ctx, input.AccountBookID, input.OperatorUserID)
		if adminErr != nil {
			return adminErr
		}
		if !canAdmin {
			return ErrFriendPermissionDenied
		}
	}

	removed, err := s.repo.DeleteCollaborator(ctx, input.AccountBookID, input.CollaboratorID)
	if err != nil {
		return err
	}
	firstAccountBookID, err := s.repo.FindFirstOwnedAccountBookID(ctx, removed.UserID)
	if err != nil {
		return err
	}
	return s.repo.UpdateUserDefaultAccountBook(ctx, removed.UserID, firstAccountBookID)
}

func (s FriendService) decryptAndValidateInviteToken(ctx context.Context, token string) (FriendInviteTokenPayload, error) {
	payload, err := s.decryptInvitePayload(token)
	if err != nil {
		return FriendInviteTokenPayload{}, ErrFriendInviteToken
	}

	if payload.InviteUserID <= 0 || payload.AccountBookID <= 0 || strings.TrimSpace(payload.Role) == "" {
		return FriendInviteTokenPayload{}, ErrFriendInviteToken
	}
	if !isSupportedCollaboratorRole(payload.Role) {
		return FriendInviteTokenPayload{}, ErrFriendInviteToken
	}
	if time.Now().Unix() > payload.ExpireAtUnix {
		return FriendInviteTokenPayload{}, ErrFriendInviteExpired
	}

	if _, err := s.repo.FindUserByID(ctx, payload.InviteUserID); err != nil {
		return FriendInviteTokenPayload{}, err
	}
	if _, err := s.repo.FindAccessibleAccountBookByID(ctx, payload.InviteUserID, payload.AccountBookID); err != nil {
		return FriendInviteTokenPayload{}, err
	}
	return payload, nil
}

func isRoleCanAdmin(role string) bool {
	role = strings.TrimSpace(strings.ToLower(role))
	return role == "owner" || role == "admin"
}

func isSupportedCollaboratorRole(role string) bool {
	_, ok := friendRoleNameMap[strings.TrimSpace(strings.ToLower(role))]
	return ok
}

func friendRoleName(role string) string {
	role = strings.TrimSpace(strings.ToLower(role))
	if v, ok := friendRoleNameMap[role]; ok {
		return v
	}
	return role
}

func (s FriendService) buildPublicURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return s.urlBuilder.BuildPublicURL(raw)
}

func (s FriendService) encryptInvitePayload(payload FriendInviteTokenPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return s.codec.Encrypt(raw)
}
func (s FriendService) decryptInvitePayload(token string) (FriendInviteTokenPayload, error) {
	raw, err := s.codec.Decrypt(token)
	if err != nil {
		return FriendInviteTokenPayload{}, ErrFriendInviteToken
	}
	var payload FriendInviteTokenPayload
	if err = json.Unmarshal(raw, &payload); err != nil {
		return FriendInviteTokenPayload{}, err
	}
	return payload, nil
}
