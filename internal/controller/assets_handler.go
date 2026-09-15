package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	assetservice "github.com/yigger/jiezhang-backend/internal/service/asset"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type AssetsHandler struct {
	service assetservice.AssetService
}

func NewAssetsHandler(service assetservice.AssetService) AssetsHandler {
	return AssetsHandler{service: service}
}

// List 资产列表/树，Query: `parent_id`(选填)
// @Summary 资产列表/树，Query: `parent_id`(选填)
// @ID AssetsHandler_List
// @Tags Assets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param parent_id query string false "parent_id"
// @Success 200 {array} types.AssetItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /assets [get]
func (h AssetsHandler) List(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
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

// Show 资产详情
// @Summary 资产详情
// @ID AssetsHandler_Show
// @Tags Assets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.AssetShowResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /assets/{id} [get]
func (h AssetsHandler) Show(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
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
		if errors.Is(err, assetservice.ErrRepositoryAssetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load asset"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Delete 删除资产（含子资产和关联账单）
// @Summary 删除资产（含子资产和关联账单）
// @ID AssetsHandler_Delete
// @Tags Assets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /assets/{id} [delete]
func (h AssetsHandler) Delete(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := RequireCurrentUser(c)
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
		case errors.Is(err, assetservice.ErrAssetPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, assetservice.ErrRepositoryAssetNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to delete asset"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// Icons 获取资产图标列表
// @Summary 获取资产图标列表
// @ID AssetsHandler_Icons
// @Tags Assets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {array} map[string]string "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /icons/assets_with_url [get]
func (h AssetsHandler) Icons(c *gin.Context) {
	items, err := h.service.ListAssetIcons()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load asset icons"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Update 更新资产，Body 同上
// @Summary 更新资产，Body 同上
// @ID AssetsHandler_Update
// @Tags Assets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Accept json
// @Param body body types.AssetWriteRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /assets/{id} [put]
func (h AssetsHandler) Update(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := RequireCurrentUser(c)
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
		case errors.Is(err, assetservice.ErrAssetPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, assetservice.ErrAssetInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "分类名不能为空哦~"})
		case errors.Is(err, assetservice.ErrRepositoryAssetNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update asset"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// Create 创建资产，Body: `{wallet: {name, amount, parent_id, icon_path, remark, type}}`
// @Summary 创建资产，Body: `{wallet: {name, amount, parent_id, icon_path, remark, type}}`
// @ID AssetsHandler_Create
// @Tags Assets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.AssetWriteRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /assets [post]
func (h AssetsHandler) Create(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := RequireCurrentUser(c)
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
		case errors.Is(err, assetservice.ErrAssetPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, assetservice.ErrAssetInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "分类名不能为空哦~"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to create asset"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// UpdateSurplus 更新资产余额，Body: `{asset_id, amount}`，amount 支持数字或字符串
// @Summary 更新资产余额，Body: `{asset_id, amount}`，amount 支持数字或字符串
// @ID AssetsHandler_UpdateSurplus
// @Tags Assets
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.AssetSurplusRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /wallet/surplus [put]
func (h AssetsHandler) UpdateSurplus(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var req types.AssetSurplusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 422, "msg": "金额必须为数字"})
		return
	}

	err := h.service.UpdateSurplus(c.Request.Context(), assetservice.AssetSurplusInput{
		UserID:        currentUser.ID,
		AccountBookID: accountBook.ID,
		AssetID:       req.AssetID,
		Amount:        string(req.Amount),
	})
	if err != nil {
		switch {
		case errors.Is(err, assetservice.ErrAssetInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 422, "msg": "金额必须为数字"})
		case errors.Is(err, assetservice.ErrAssetPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "无权限修改此属性"})
		case errors.Is(err, assetservice.ErrRepositoryAssetNotFound):
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
		return 0, ErrInvalidParam("id")
	}
	return id, nil
}

func parseAssetWriteInput(c *gin.Context) (assetservice.AssetWriteInput, error) {
	var req types.AssetWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return assetservice.AssetWriteInput{}, err
	}
	return assetservice.AssetWriteInput{
		Name:     req.Wallet.Name,
		Amount:   req.Wallet.Amount,
		ParentID: req.Wallet.ParentID,
		IconPath: req.Wallet.IconPath,
		Remark:   req.Wallet.Remark,
		Type:     req.Wallet.Type,
	}, nil
}
