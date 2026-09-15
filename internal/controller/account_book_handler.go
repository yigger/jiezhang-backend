package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	accountbookservice "github.com/yigger/jiezhang-backend/internal/service/accountbook"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type AccountBookHandler struct {
	service accountbookservice.AccountBookService
}

func NewAccountBookHandler(service accountbookservice.AccountBookService) AccountBookHandler {
	return AccountBookHandler{service: service}
}

// List 用户可访问的账本列表
// @Summary 用户可访问的账本列表
// @ID AccountBookHandler_List
// @Tags AccountBook
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {array} types.AccountBookListItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /account_books [get]
func (h AccountBookHandler) List(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	items, err := h.service.List(c.Request.Context(), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to list account books"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Show 账本详情
// @Summary 账本详情
// @ID AccountBookHandler_Show
// @Tags AccountBook
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.APIResponse{data=types.AccountBookDetailItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /account_books/{id} [get]
func (h AccountBookHandler) Show(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseAccountBookID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid account book id"})
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id, currentUser.ID)
	if err != nil {
		if errors.Is(err, accountbookservice.ErrAccountBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的账簿"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load account book"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": item})
}

// Types 账本类型列表
// @Summary 账本类型列表
// @ID AccountBookHandler_Types
// @Tags AccountBook
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse{data=types.APIResponse} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /account_books/types [get]
func (h AccountBookHandler) Types(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": h.service.Types()})
}

// PresetCategories 预设分类和资产模板，Query: `account_type`
// @Summary 预设分类和资产模板，Query: `account_type`
// @ID AccountBookHandler_PresetCategories
// @Tags AccountBook
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param account_type query string false "account_type"
// @Success 200 {object} types.APIResponse{data=types.AccountBookPreset} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /account_books/preset_categories [get]
func (h AccountBookHandler) PresetCategories(c *gin.Context) {
	preset, err := h.service.PresetCategories(c.Query("account_type"))
	if err != nil {
		if errors.Is(err, accountbookservice.ErrAccountBookInvalidType) {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "无效的账户类型"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load preset categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": preset})
}

// Switch 切换默认账本
// @Summary 切换默认账本
// @ID AccountBookHandler_Switch
// @Tags AccountBook
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /account_books/{id}/switch [put]
func (h AccountBookHandler) Switch(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseAccountBookID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid account book id"})
		return
	}

	if err := h.service.Switch(c.Request.Context(), currentUser.ID, id); err != nil {
		if errors.Is(err, accountbookservice.ErrAccountBookNotFound) || errors.Is(err, accountbookservice.ErrRepositoryAccountBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的账簿"})
			return
		}
		if errors.Is(err, accountbookservice.ErrRepositoryUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "用户不存在"})
			return
		}
		log.Printf("[Switch] error switching account book %d for user %d: %v", id, currentUser.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to switch account book"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "msg": "ok"})
}

// Create 创建账本，Body: `{name, description, account_type, categories, assets}`
// @Summary 创建账本，Body: `{name, description, account_type, categories, assets}`
// @ID AccountBookHandler_Create
// @Tags AccountBook
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.AccountBookCreateRequest true "请求体"
// @Success 200 {object} types.APIResponse{data=types.AccountBookListItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /account_books [post]
func (h AccountBookHandler) Create(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var req types.AccountBookCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}

	item, err := h.service.Create(c.Request.Context(), accountbookservice.AccountBookCreateInput{
		UserID:       currentUser.ID,
		UserNickname: currentUser.Nickname,
		Name:         req.Name,
		Description:  req.Description,
		AccountType:  req.AccountType,
		Categories:   toServiceAccountBookCategories(req.Categories),
		Assets:       toServiceAccountBookAssets(req.Assets),
	})
	if err != nil {
		switch {
		case errors.Is(err, accountbookservice.ErrAccountBookInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "名称不能为空"})
		case errors.Is(err, accountbookservice.ErrAccountBookInvalidType):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "无效的账户类型"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "创建账簿失败"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": item})
}

// Update 更新账本，Body: `{name, description, account_type}`
// @Summary 更新账本，Body: `{name, description, account_type}`
// @ID AccountBookHandler_Update
// @Tags AccountBook
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Accept json
// @Param body body types.AccountBookUpdateRequest true "请求体"
// @Success 200 {object} types.APIResponse{data=types.AccountBookListItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /account_books/{id} [put]
func (h AccountBookHandler) Update(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseAccountBookID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid account book id"})
		return
	}

	var req types.AccountBookUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}

	item, err := h.service.Update(c.Request.Context(), accountbookservice.AccountBookUpdateInput{
		UserID:      currentUser.ID,
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		AccountType: req.AccountType.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, accountbookservice.ErrAccountBookNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的账簿"})
		case errors.Is(err, accountbookservice.ErrAccountBookInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "名称不能为空"})
		case errors.Is(err, accountbookservice.ErrAccountBookInvalidType):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "无效的账户类型"})
		case errors.Is(err, accountbookservice.ErrAccountBookPermissionDenied):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "仅账簿创建者可变更"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "更新账簿失败"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": item})
}

// Delete 删除账本
// @Summary 删除账本
// @ID AccountBookHandler_Delete
// @Tags AccountBook
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /account_books/{id} [delete]
func (h AccountBookHandler) Delete(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseAccountBookID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid account book id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), currentUser.ID, id, currentUser.AccountBookId); err != nil {
		switch {
		case errors.Is(err, accountbookservice.ErrAccountBookNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的账簿"})
		case errors.Is(err, accountbookservice.ErrAccountBookPermissionDenied):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "仅账簿创建者可变更"})
		case errors.Is(err, accountbookservice.ErrAccountBookInUse):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "不能删除正在使用的账簿，请先切换到其它账簿"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "删除失败"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "msg": "ok"})
}

func parseAccountBookID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidParam("id")
	}
	return id, nil
}

func toServiceAccountBookCategories(src map[string][]types.AccountBookCategoryItem) map[string][]accountbookservice.AccountBookCategoryInput {
	if len(src) == 0 {
		return nil
	}
	res := make(map[string][]accountbookservice.AccountBookCategoryInput, len(src))
	for key, parents := range src {
		items := make([]accountbookservice.AccountBookCategoryInput, 0, len(parents))
		for _, p := range parents {
			children := make([]accountbookservice.AccountBookChildInput, 0, len(p.Childs))
			for _, child := range p.Childs {
				children = append(children, accountbookservice.AccountBookChildInput{Name: child.Name, IconPath: child.IconPath})
			}
			items = append(items, accountbookservice.AccountBookCategoryInput{Name: p.Name, IconPath: p.IconPath, Childs: children})
		}
		res[key] = items
	}
	return res
}

func toServiceAccountBookAssets(src []types.AccountBookAssetItem) []accountbookservice.AccountBookAssetInput {
	if len(src) == 0 {
		return nil
	}
	res := make([]accountbookservice.AccountBookAssetInput, 0, len(src))
	for _, a := range src {
		children := make([]accountbookservice.AccountBookChildInput, 0, len(a.Childs))
		for _, child := range a.Childs {
			children = append(children, accountbookservice.AccountBookChildInput{Name: child.Name, IconPath: child.IconPath})
		}
		res = append(res, accountbookservice.AccountBookAssetInput{Name: a.Name, IconPath: a.IconPath, Type: a.Type, Childs: children})
	}
	return res
}
