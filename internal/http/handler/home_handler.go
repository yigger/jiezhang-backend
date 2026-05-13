package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/service"
)

type HomeHandler struct {
	service service.HomeService
}

func NewHomeHandler(homeService service.HomeService) HomeHandler {
	return HomeHandler{service: homeService}
}

func (h HomeHandler) Header(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.service.GetHeader(c.Request.Context(), currentUser.ID, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get header"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h HomeHandler) Index(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	items, err := h.service.GetIndex(c.Request.Context(), currentUser.ID, accountBook.ID, c.Query("range"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get index list"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h HomeHandler) GetSettings(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.service.GetSettings(c.Request.Context(), currentUser, accountBook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get settings"})
		return
	}
	c.JSON(http.StatusOK, res)
}
