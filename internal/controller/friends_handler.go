package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	friendservice "github.com/yigger/jiezhang-backend/internal/service/friend"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type FriendsHandler struct {
	service friendservice.FriendService
}

func NewFriendsHandler(service friendservice.FriendService) FriendsHandler {
	return FriendsHandler{service: service}
}

// List 协作者列表
// @Summary 协作者列表
// @ID FriendsHandler_List
// @Tags Friends
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse{data=types.FriendListResponse} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /friends [get]
func (h FriendsHandler) List(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.service.List(c.Request.Context(), accountBook.ID, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, friendservice.ErrRepositoryFriendAccountBookNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "账本不存在"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load collaborators"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

// Invite 邀请，Body: `{account_book_id, role}`
// @Summary 邀请，Body: `{account_book_id, role}`
// @ID FriendsHandler_Invite
// @Tags Friends
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.FriendInviteRequest true "请求体"
// @Success 200 {object} types.APIResponse{data=string} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /friends/invite [post]
func (h FriendsHandler) Invite(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}

	var req types.FriendInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}

	token, err := h.service.Invite(c.Request.Context(), friendservice.FriendInviteInput{
		AccountBookID: req.AccountBookID,
		UserID:        currentUser.ID,
		Role:          req.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, friendservice.ErrFriendInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		case errors.Is(err, friendservice.ErrFriendPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "管理员才可以邀请他人"})
		case errors.Is(err, friendservice.ErrRepositoryFriendAccountBookNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "账本不存在"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to invite collaborator"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": token})
}

// InviteInformation 查看邀请信息，Query: `invite_token`
// @Summary 查看邀请信息，Query: `invite_token`
// @ID FriendsHandler_InviteInformation
// @Tags Friends
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse{data=types.InviteInformationItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /friends/invite_information [get]
func (h FriendsHandler) InviteInformation(c *gin.Context) {
	var req types.FriendInviteInformationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "not found"})
		return
	}

	res, err := h.service.InviteInformation(c.Request.Context(), req.InviteToken)
	if err != nil {
		switch {
		case errors.Is(err, friendservice.ErrFriendInviteToken):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "not found"})
		case errors.Is(err, friendservice.ErrFriendInviteExpired):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "邀请已过期"})
		case errors.Is(err, friendservice.ErrRepositoryUserNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "用户不存在"})
		case errors.Is(err, friendservice.ErrRepositoryFriendAccountBookNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "账本不存在"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load invite information"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

// AcceptApply 接受邀请，Body: `{invite_token, nickname}`
// @Summary 接受邀请，Body: `{invite_token, nickname}`
// @ID FriendsHandler_AcceptApply
// @Tags Friends
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.FriendAcceptApplyRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /friends/accept_apply [post]
func (h FriendsHandler) AcceptApply(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var req types.FriendAcceptApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}

	err := h.service.AcceptApply(c.Request.Context(), friendservice.FriendAcceptInput{
		UserID:      currentUser.ID,
		InviteToken: req.InviteToken,
		Nickname:    req.Nickname,
	})
	if err != nil {
		switch {
		case errors.Is(err, friendservice.ErrFriendAlreadyMember):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "你已经是成员啦，马上为您切换..."})
		case errors.Is(err, friendservice.ErrFriendInviteExpired):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "邀请已过期"})
		case errors.Is(err, friendservice.ErrFriendInviteToken), errors.Is(err, friendservice.ErrFriendInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to accept invite"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "msg": "ok"})
}

// Remove 移除协作者
// @Summary 移除协作者
// @ID FriendsHandler_Remove
// @Tags Friends
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param collaboratorId path integer true "collaboratorId"
// @Accept json
// @Param body body types.FriendRemoveRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /friends/{collaboratorId} [delete]
func (h FriendsHandler) Remove(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	collaboratorID, err := parseCollaboratorID(c.Param("collaboratorId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}
	var req types.FriendRemoveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.ShouldBindJSON(&req)
	}

	err = h.service.Remove(c.Request.Context(), friendservice.FriendRemoveInput{
		AccountBookID:  req.AccountBookID,
		OperatorUserID: currentUser.ID,
		CollaboratorID: collaboratorID,
	})
	if err != nil {
		switch {
		case errors.Is(err, friendservice.ErrFriendInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		case errors.Is(err, friendservice.ErrFriendPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "只有账簿拥有者可以删除成员"})
		case errors.Is(err, friendservice.ErrRepositoryFriendCollaboratorNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "该用户不是成员"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to remove collaborator"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// Update 更新协作者，Body: `{account_book_id, role?, remark?}`
// @Summary 更新协作者，Body: `{account_book_id, role?, remark?}`
// @ID FriendsHandler_Update
// @Tags Friends
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param collaboratorId path integer true "collaboratorId"
// @Accept json
// @Param body body types.FriendUpdateRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /friends/{collaboratorId} [put]
func (h FriendsHandler) Update(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	collaboratorID, err := parseCollaboratorID(c.Param("collaboratorId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}

	var req types.FriendUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}

	err = h.service.Update(c.Request.Context(), friendservice.FriendUpdateInput{
		AccountBookID:  req.AccountBookID,
		OperatorUserID: currentUser.ID,
		CollaboratorID: collaboratorID,
		Role:           req.Role,
		Remark:         req.Remark,
	})
	if err != nil {
		switch {
		case errors.Is(err, friendservice.ErrFriendInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		case errors.Is(err, friendservice.ErrFriendCannotEditSelf):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您不能编辑自己的权限"})
		case errors.Is(err, friendservice.ErrFriendPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "管理员才可以邀请他人"})
		case errors.Is(err, friendservice.ErrRepositoryFriendCollaboratorNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "该用户不是成员"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update collaborator"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func parseCollaboratorID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidParam("collaborator_id")
	}
	return id, nil
}
