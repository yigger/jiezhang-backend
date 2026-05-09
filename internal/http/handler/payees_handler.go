package handler

import "github.com/gin-gonic/gin"

type PayeesHandler struct{}

func NewPayeesHandler() PayeesHandler {
	return PayeesHandler{}
}

func (h PayeesHandler) List(c *gin.Context) {
	notImplemented(c, "GET /api/payees")
}

func (h PayeesHandler) Create(c *gin.Context) {
	notImplemented(c, "POST /api/payees")
}

func (h PayeesHandler) Update(c *gin.Context) {
	notImplemented(c, "PUT /api/payees/:id")
}

func (h PayeesHandler) Delete(c *gin.Context) {
	notImplemented(c, "DELETE /api/payees/:id")
}
