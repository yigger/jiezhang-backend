package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	httpdto "github.com/yigger/jiezhang-backend/internal/http/dto"
	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type FriendsHandler struct {
	service service.FriendService
}

func NewFriendsHandler(service service.FriendService) FriendsHandler {
	return FriendsHandler{service: service}
}

func (h FriendsHandler) List(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.service.List(c.Request.Context(), accountBook.ID, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrFriendAccountBookNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "账本不存在"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load collaborators"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

func (h FriendsHandler) Invite(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}

	var req httpdto.FriendInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}

	token, err := h.service.Invite(c.Request.Context(), service.FriendInviteInput{
		AccountBookID: req.AccountBookID,
		UserID:        currentUser.ID,
		Role:          req.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFriendInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		case errors.Is(err, service.ErrFriendPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "管理员才可以邀请他人"})
		case errors.Is(err, repository.ErrFriendAccountBookNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "账本不存在"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to invite collaborator"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": token})
}

func (h FriendsHandler) InviteInformation(c *gin.Context) {
	var req httpdto.FriendInviteInformationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "not found"})
		return
	}

	res, err := h.service.InviteInformation(c.Request.Context(), req.InviteToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFriendInviteToken):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "not found"})
		case errors.Is(err, service.ErrFriendInviteExpired):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "邀请已过期"})
		case errors.Is(err, repository.ErrUserNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "用户不存在"})
		case errors.Is(err, repository.ErrFriendAccountBookNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "账本不存在"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load invite information"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

func (h FriendsHandler) AcceptApply(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	var req httpdto.FriendAcceptApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}

	err := h.service.AcceptApply(c.Request.Context(), service.FriendAcceptInput{
		UserID:      currentUser.ID,
		InviteToken: req.InviteToken,
		Nickname:    req.Nickname,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFriendAlreadyMember):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "你已经是成员啦，马上为您切换..."})
		case errors.Is(err, service.ErrFriendInviteExpired):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "邀请已过期"})
		case errors.Is(err, service.ErrFriendInviteToken), errors.Is(err, service.ErrFriendInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to accept invite"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "msg": "ok"})
}

func (h FriendsHandler) Remove(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	collaboratorID, err := parseCollaboratorID(c.Param("collaboratorId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}
	var req httpdto.FriendRemoveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.ShouldBindJSON(&req)
	}

	err = h.service.Remove(c.Request.Context(), service.FriendRemoveInput{
		AccountBookID:  req.AccountBookID,
		OperatorUserID: currentUser.ID,
		CollaboratorID: collaboratorID,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFriendInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		case errors.Is(err, service.ErrFriendPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "只有账簿拥有者可以删除成员"})
		case errors.Is(err, repository.ErrFriendCollaboratorNotFound):
			c.JSON(http.StatusOK, gin.H{"status": 404, "msg": "该用户不是成员"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to remove collaborator"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h FriendsHandler) Update(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	collaboratorID, err := parseCollaboratorID(c.Param("collaboratorId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}

	var req httpdto.FriendUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}

	err = h.service.Update(c.Request.Context(), service.FriendUpdateInput{
		AccountBookID:  req.AccountBookID,
		OperatorUserID: currentUser.ID,
		CollaboratorID: collaboratorID,
		Role:           req.Role,
		Remark:         req.Remark,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFriendInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		case errors.Is(err, service.ErrFriendCannotEditSelf):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您不能编辑自己的权限"})
		case errors.Is(err, service.ErrFriendPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "管理员才可以邀请他人"})
		case errors.Is(err, repository.ErrFriendCollaboratorNotFound):
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
		return 0, errInvalidParam("collaborator_id")
	}
	return id, nil
}
