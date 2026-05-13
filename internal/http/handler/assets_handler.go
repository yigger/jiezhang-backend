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

type AssetsHandler struct {
	service service.AssetService
}

func NewAssetsHandler(service service.AssetService) AssetsHandler {
	return AssetsHandler{service: service}
}

func (h AssetsHandler) List(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	parentID := int64(0)
	if raw := strings.TrimSpace(c.Query("parent_id")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid parent_id"})
			return
		}
		parentID = v
	}

	if parentID == 0 {
		items, err := h.service.ListTree(c.Request.Context(), accountBook.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load assets"})
			return
		}
		c.JSON(http.StatusOK, items)
		return
	}

	items, err := h.service.ListByParent(c.Request.Context(), accountBook.ID, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load assets"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h AssetsHandler) Show(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	id, err := parseAssetID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid asset id"})
		return
	}

	res, err := h.service.Show(c.Request.Context(), accountBook.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrAssetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load asset"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h AssetsHandler) Delete(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseAssetID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid asset id"})
		return
	}

	err = h.service.Delete(c.Request.Context(), id, accountBook.ID, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAssetPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, repository.ErrAssetNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to delete asset"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h AssetsHandler) Icons(c *gin.Context) {
	items, err := h.service.ListAssetIcons()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load asset icons"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h AssetsHandler) Update(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseAssetID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid asset id"})
		return
	}

	input, err := parseAssetWriteInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}
	input.CreatorID = currentUser.ID
	input.AccountBookID = accountBook.ID

	err = h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAssetPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, service.ErrAssetInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "分类名不能为空哦~"})
		case errors.Is(err, repository.ErrAssetNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update asset"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h AssetsHandler) Create(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}

	input, err := parseAssetWriteInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}
	input.CreatorID = currentUser.ID
	input.AccountBookID = accountBook.ID

	err = h.service.Create(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAssetPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, service.ErrAssetInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "分类名不能为空哦~"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to create asset"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h AssetsHandler) UpdateSurplus(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	var req httpdto.AssetSurplusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 422, "msg": "金额必须为数字"})
		return
	}

	err := h.service.UpdateSurplus(c.Request.Context(), service.AssetSurplusInput{
		UserID:        currentUser.ID,
		AccountBookID: accountBook.ID,
		AssetID:       req.AssetID,
		Amount:        req.Amount,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAssetInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 422, "msg": "金额必须为数字"})
		case errors.Is(err, service.ErrAssetPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "无权限修改此属性"})
		case errors.Is(err, repository.ErrAssetNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update surplus"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func parseAssetID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidParam("id")
	}
	return id, nil
}

func parseAssetWriteInput(c *gin.Context) (service.AssetWriteInput, error) {
	var req httpdto.AssetWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return service.AssetWriteInput{}, err
	}
	return service.AssetWriteInput{
		Name:     req.Wallet.Name,
		Amount:   req.Wallet.Amount,
		ParentID: req.Wallet.ParentID,
		IconPath: req.Wallet.IconPath,
		Remark:   req.Wallet.Remark,
		Type:     req.Wallet.Type,
	}, nil
}
