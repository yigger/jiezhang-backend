package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
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
	repo       repository.FriendRepository
	urlBuilder FriendURLBuilder
	secret     string
}

type FriendURLBuilder interface {
	BuildPublicURL(raw string) string
}

func NewFriendService(repo repository.FriendRepository, urlBuilder FriendURLBuilder, secret string) FriendService {
	return FriendService{
		repo:       repo,
		urlBuilder: urlBuilder,
		secret:     strings.TrimSpace(secret),
	}
}

type FriendListResponse struct {
	Collaborators []FriendCollaboratorItem `json:"collaborators"`
	Owner         FriendUserItem           `json:"owner"`
	Authority     FriendAuthority          `json:"authority"`
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

type FriendUserItem struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar_path"`
}

type FriendAuthority struct {
	ChangeRole bool `json:"change_role"`
	Remove     bool `json:"remove"`
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

type InviteInformationItem struct {
	InviteUser  FriendUserItem        `json:"invite_user"`
	AccountBook InviteInformationBook `json:"account_book"`
	RoleName    string                `json:"role_name"`
}

type InviteInformationBook struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type FriendInviteTokenPayload struct {
	InviteUserID  int64  `json:"invite_user_id"`
	AccountBookID int64  `json:"account_book_id"`
	Role          string `json:"role"`
	ExpireAtUnix  int64  `json:"expire_at_unix"`
}

func (s FriendService) List(ctx context.Context, accountBookID int64, currentUserID int64) (FriendListResponse, error) {
	accountBook, err := s.repo.FindAccessibleAccountBookByID(ctx, currentUserID, accountBookID)
	if err != nil {
		return FriendListResponse{}, err
	}

	collaborators, err := s.repo.ListCollaborators(ctx, accountBook.ID)
	if err != nil {
		return FriendListResponse{}, err
	}
	operator, err := s.repo.FindCollaboratorByUserID(ctx, accountBook.ID, currentUserID)
	if err != nil {
		return FriendListResponse{}, err
	}

	owner, err := s.repo.FindUserByID(ctx, accountBook.UserID)
	if err != nil {
		return FriendListResponse{}, err
	}

	items := make([]FriendCollaboratorItem, 0, len(collaborators))
	for _, c := range collaborators {
		items = append(items, FriendCollaboratorItem{
			ID:       c.ID,
			Role:     c.Role,
			RoleName: friendRoleName(c.Role),
			Remark:   c.Remark,
			User: FriendUserItem{
				ID:       c.UserID,
				Nickname: c.UserNickname,
				Avatar:   s.buildPublicURL(c.UserAvatarURL),
			},
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}

	canAdmin := isRoleCanAdmin(operator.Role)
	return FriendListResponse{
		Collaborators: items,
		Owner: FriendUserItem{
			ID:       owner.ID,
			Nickname: owner.Nickname,
			Avatar:   s.buildPublicURL(owner.AvatarURL),
		},
		Authority: FriendAuthority{
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

func (s FriendService) InviteInformation(ctx context.Context, token string) (InviteInformationItem, error) {
	payload, err := s.decryptAndValidateInviteToken(ctx, strings.TrimSpace(token))
	if err != nil {
		return InviteInformationItem{}, err
	}

	inviteUser, err := s.repo.FindUserByID(ctx, payload.InviteUserID)
	if err != nil {
		return InviteInformationItem{}, err
	}
	accountBook, err := s.repo.FindAccountBookByID(ctx, payload.AccountBookID)
	if err != nil {
		return InviteInformationItem{}, err
	}

	return InviteInformationItem{
		InviteUser: FriendUserItem{
			ID:       inviteUser.ID,
			Nickname: inviteUser.Nickname,
			Avatar:   s.buildPublicURL(inviteUser.AvatarURL),
		},
		AccountBook: InviteInformationBook{
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
	} else if !errors.Is(err, repository.ErrFriendCollaboratorNotFound) {
		return err
	}

	if err := s.repo.CreateCollaborator(ctx, repository.FriendCollaboratorCreateRecord{
		AccountBookID: payload.AccountBookID,
		UserID:        input.UserID,
		Role:          payload.Role,
		Remark:        strings.TrimSpace(input.Nickname),
	}); err != nil {
		if errors.Is(err, repository.ErrFriendCollaboratorExists) {
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

	update := repository.FriendCollaboratorUpdateRecord{
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

func (s FriendService) encryptInvitePayload(payload FriendInviteTokenPayload) (string, error) {
	if strings.TrimSpace(s.secret) == "" {
		return "", ErrFriendInvalidInput
	}
	plain, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	key := sha256.Sum256([]byte(s.secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	plain = pkcs7Pad(plain, aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plain))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plain)

	combined := append(iv, ciphertext...)
	first := base64.StdEncoding.EncodeToString(combined)
	second := base64.StdEncoding.EncodeToString([]byte(first))
	return second, nil
}

func (s FriendService) decryptInvitePayload(token string) (FriendInviteTokenPayload, error) {
	if strings.TrimSpace(s.secret) == "" {
		return FriendInviteTokenPayload{}, ErrFriendInviteToken
	}
	firstBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return FriendInviteTokenPayload{}, err
	}
	combined, err := base64.StdEncoding.DecodeString(string(firstBytes))
	if err != nil {
		return FriendInviteTokenPayload{}, err
	}
	if len(combined) < aes.BlockSize || len(combined)%aes.BlockSize != 0 {
		return FriendInviteTokenPayload{}, ErrFriendInviteToken
	}

	key := sha256.Sum256([]byte(s.secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return FriendInviteTokenPayload{}, err
	}
	iv := combined[:aes.BlockSize]
	ciphertext := combined[aes.BlockSize:]
	plain := make([]byte, len(ciphertext))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plain, ciphertext)

	plain, err = pkcs7Unpad(plain, aes.BlockSize)
	if err != nil {
		return FriendInviteTokenPayload{}, err
	}

	var payload FriendInviteTokenPayload
	if err := json.Unmarshal(plain, &payload); err != nil {
		return FriendInviteTokenPayload{}, err
	}
	return payload, nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := bytesRepeat(byte(padding), padding)
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, ErrFriendInviteToken
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, ErrFriendInviteToken
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, ErrFriendInviteToken
		}
	}
	return data[:len(data)-padding], nil
}

func bytesRepeat(b byte, count int) []byte {
	if count <= 0 {
		return []byte{}
	}
	buf := make([]byte, count)
	for i := 0; i < count; i++ {
		buf[i] = b
	}
	return buf
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
	if s.urlBuilder == nil {
		return raw
	}
	return s.urlBuilder.BuildPublicURL(raw)
}
