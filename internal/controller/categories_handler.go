package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	categoryservice "github.com/yigger/jiezhang-backend/internal/service/category"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type CategoriesHandler struct {
	service categoryservice.CategoryService
}

func NewCategoriesHandler(service categoryservice.CategoryService) CategoriesHandler {
	return CategoriesHandler{service: service}
}

// List 分类列表（按 parent_id 展开），Query: `type`(默认 expend)、`parent_id`(默认 0)
// @Summary 分类列表（按 parent_id 展开），Query: `type`(默认 expend)、`parent_id`(默认 0)
// @ID CategoriesHandler_List
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param parent_id query string false "parent_id"
// @Param type query string false "type"
// @Success 200 {object} types.CategoryListResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /categories/category_list [get]
func (h CategoriesHandler) List(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	statementType := strings.TrimSpace(c.Query("type"))
	if statementType == "" {
		statementType = "expend"
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

	res, err := h.service.ListByParent(c.Request.Context(), accountBook.ID, statementType, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load categories"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Show 分类详情
// @Summary 分类详情
// @ID CategoriesHandler_Show
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.CategoryShowResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /categories/{id} [get]
func (h CategoriesHandler) Show(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	id, err := ParseCategoryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid category id"})
		return
	}

	res, err := h.service.Show(c.Request.Context(), accountBook.ID, id)
	if err != nil {
		if errors.Is(err, categoryservice.ErrRepositoryCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load category"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Parent 父级分类树，Query: `type`(默认 expend)
// @Summary 父级分类树，Query: `type`(默认 expend)
// @ID CategoriesHandler_Parent
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {array} types.CategoryItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /categories/parent [get]
func (h CategoriesHandler) Parent(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	statementType := strings.TrimSpace(c.Query("type"))
	if statementType == "" {
		statementType = "expend"
	}

	res, err := h.service.ListTree(c.Request.Context(), accountBook.ID, statementType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load categories"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// CategoryChilds 子分类列表，Query: `parent_id`(必填)
// @Summary 子分类列表，Query: `parent_id`(必填)
// @ID CategoriesHandler_CategoryChilds
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.CategoryListResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /categories/category_childs [get]
func (h CategoriesHandler) CategoryChilds(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	parentID, err := ParseInt64Query(c, "parent_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid parent_id"})
		return
	}
	res, err := h.service.ListByParent(c.Request.Context(), accountBook.ID, "", parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load categories"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// CategoryStatements 分类下的账单列表，Query: `category_id`(必填)
// @Summary 分类下的账单列表，Query: `category_id`(必填)
// @ID CategoriesHandler_CategoryStatements
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {array} types.CategoryStatementsMonthItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /categories/category_statements [get]
func (h CategoriesHandler) CategoryStatements(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	categoryID, err := ParseInt64Query(c, "category_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid category_id"})
		return
	}
	res, err := h.service.ListStatementsByCategory(c.Request.Context(), accountBook.ID, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load category statements"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Delete 删除分类（含子分类和关联账单）
// @Summary 删除分类（含子分类和关联账单）
// @ID CategoriesHandler_Delete
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /categories/{id} [delete]
func (h CategoriesHandler) Delete(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	id, err := ParseCategoryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid category id"})
		return
	}

	err = h.service.Delete(c.Request.Context(), id, accountBook.ID, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, categoryservice.ErrCategoryPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, categoryservice.ErrRepositoryCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to delete category"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// Icons 获取分类图标列表
// @Summary 获取分类图标列表
// @ID CategoriesHandler_Icons
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {array} map[string]string "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /icons/categories_with_url [get]
func (h CategoriesHandler) Icons(c *gin.Context) {
	items, err := h.service.ListCategoryIcons()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load category icons"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Update 更新分类，Body 同上
// @Summary 更新分类，Body 同上
// @ID CategoriesHandler_Update
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param id path integer true "id"
// @Accept json
// @Param body body types.CategoryWriteRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /categories/{id} [put]
func (h CategoriesHandler) Update(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	id, err := ParseCategoryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid category id"})
		return
	}

	input, err := parseCategoryWriteInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}
	input.UserID = currentUser.ID
	input.AccountBookID = accountBook.ID

	err = h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, categoryservice.ErrCategoryPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, categoryservice.ErrCategoryInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "分类名不能为空哦~"})
		case errors.Is(err, categoryservice.ErrRepositoryCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update category"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// Create 创建分类，Body: `{category: {name, parent_id, icon_path, type}}`
// @Summary 创建分类，Body: `{category: {name, parent_id, icon_path, type}}`
// @ID CategoriesHandler_Create
// @Tags Categories
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.CategoryWriteRequest true "请求体"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /categories [post]
func (h CategoriesHandler) Create(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}

	input, err := parseCategoryWriteInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid request body"})
		return
	}
	input.UserID = currentUser.ID
	input.AccountBookID = accountBook.ID

	err = h.service.Create(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, categoryservice.ErrCategoryPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, categoryservice.ErrCategoryInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "分类名不能为空哦~"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to create category"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func ParseCategoryID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidParam("id")
	}
	return id, nil
}

func parseCategoryWriteInput(c *gin.Context) (categoryservice.CategoryWriteInput, error) {
	var req types.CategoryWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return categoryservice.CategoryWriteInput{}, err
	}
	return categoryservice.CategoryWriteInput{
		Name:     req.Category.Name,
		ParentID: req.Category.ParentID,
		IconPath: req.Category.IconPath,
		Type:     req.Category.Type,
	}, nil
}
