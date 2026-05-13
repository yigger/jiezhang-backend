package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	httpdto "github.com/yigger/jiezhang-backend/internal/http/dto"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type AccountBookHandler struct {
	service service.AccountBookService
}

func NewAccountBookHandler(service service.AccountBookService) AccountBookHandler {
	return AccountBookHandler{service: service}
}

func (h AccountBookHandler) List(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
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

func (h AccountBookHandler) Show(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
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
		if errors.Is(err, service.ErrAccountBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的账簿"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load account book"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": item})
}

func (h AccountBookHandler) Types(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": h.service.Types()})
}

func (h AccountBookHandler) PresetCategories(c *gin.Context) {
	preset, err := h.service.PresetCategories(c.Query("account_type"))
	if err != nil {
		if errors.Is(err, service.ErrAccountBookInvalidType) {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "无效的账户类型"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load preset categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": preset})
}

func (h AccountBookHandler) Switch(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseAccountBookID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid account book id"})
		return
	}

	if err := h.service.Switch(c.Request.Context(), currentUser.ID, id); err != nil {
		if errors.Is(err, service.ErrAccountBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的账簿"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to switch account book"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "msg": "ok"})
}

func (h AccountBookHandler) Create(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	var req httpdto.AccountBookCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}

	item, err := h.service.Create(c.Request.Context(), service.AccountBookCreateInput{
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
		case errors.Is(err, service.ErrAccountBookInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "名称不能为空"})
		case errors.Is(err, service.ErrAccountBookInvalidType):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "无效的账户类型"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "创建账簿失败"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": item})
}

func (h AccountBookHandler) Update(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseAccountBookID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid account book id"})
		return
	}

	var req httpdto.AccountBookUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}

	item, err := h.service.Update(c.Request.Context(), service.AccountBookUpdateInput{
		UserID:      currentUser.ID,
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		AccountType: req.AccountType.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAccountBookNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的账簿"})
		case errors.Is(err, service.ErrAccountBookInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "名称不能为空"})
		case errors.Is(err, service.ErrAccountBookInvalidType):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "无效的账户类型"})
		case errors.Is(err, service.ErrAccountBookPermissionDenied):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "仅账簿创建者可变更"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "更新账簿失败"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": item})
}

func (h AccountBookHandler) Delete(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
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
		case errors.Is(err, service.ErrAccountBookNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "无效的账簿"})
		case errors.Is(err, service.ErrAccountBookPermissionDenied):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "仅账簿创建者可变更"})
		case errors.Is(err, service.ErrAccountBookInUse):
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
		return 0, errInvalidParam("id")
	}
	return id, nil
}

func toServiceAccountBookCategories(src map[string][]httpdto.AccountBookCategoryItem) map[string][]service.AccountBookCategoryInput {
	if len(src) == 0 {
		return nil
	}
	res := make(map[string][]service.AccountBookCategoryInput, len(src))
	for key, parents := range src {
		items := make([]service.AccountBookCategoryInput, 0, len(parents))
		for _, p := range parents {
			children := make([]service.AccountBookChildInput, 0, len(p.Childs))
			for _, child := range p.Childs {
				children = append(children, service.AccountBookChildInput{Name: child.Name, IconPath: child.IconPath})
			}
			items = append(items, service.AccountBookCategoryInput{Name: p.Name, IconPath: p.IconPath, Childs: children})
		}
		res[key] = items
	}
	return res
}

func toServiceAccountBookAssets(src []httpdto.AccountBookAssetItem) []service.AccountBookAssetInput {
	if len(src) == 0 {
		return nil
	}
	res := make([]service.AccountBookAssetInput, 0, len(src))
	for _, a := range src {
		children := make([]service.AccountBookChildInput, 0, len(a.Childs))
		for _, child := range a.Childs {
			children = append(children, service.AccountBookChildInput{Name: child.Name, IconPath: child.IconPath})
		}
		res = append(res, service.AccountBookAssetInput{Name: a.Name, IconPath: a.IconPath, Type: a.Type, Childs: children})
	}
	return res
}
