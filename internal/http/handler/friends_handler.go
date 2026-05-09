package handler

import "github.com/gin-gonic/gin"

type FriendsHandler struct{}

func NewFriendsHandler() FriendsHandler {
	return FriendsHandler{}
}

func (h FriendsHandler) List(c *gin.Context) {
	notImplemented(c, "GET /api/friends")
}

func (h FriendsHandler) Invite(c *gin.Context) {
	notImplemented(c, "POST /api/friends/invite")
}

func (h FriendsHandler) InviteInformation(c *gin.Context) {
	notImplemented(c, "GET /api/friends/invite_information")
}

func (h FriendsHandler) AcceptApply(c *gin.Context) {
	notImplemented(c, "POST /api/friends/accept_apply")
}

func (h FriendsHandler) Remove(c *gin.Context) {
	notImplemented(c, "DELETE /api/friends/:collaboratorId")
}

func (h FriendsHandler) Update(c *gin.Context) {
	notImplemented(c, "PUT /api/friends/:collaboratorId")
}
