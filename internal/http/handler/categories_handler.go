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

type CategoriesHandler struct {
	service service.CategoryService
}

func NewCategoriesHandler(service service.CategoryService) CategoriesHandler {
	return CategoriesHandler{service: service}
}

func (h CategoriesHandler) List(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h CategoriesHandler) Show(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	id, err := parseCategoryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid category id"})
		return
	}

	res, err := h.service.Show(c.Request.Context(), accountBook.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load category"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h CategoriesHandler) Parent(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
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

func (h CategoriesHandler) CategoryChilds(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	parentID, err := parseInt64Query(c, "parent_id")
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

func (h CategoriesHandler) CategoryStatements(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	categoryID, err := parseInt64Query(c, "category_id")
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

func (h CategoriesHandler) Delete(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseCategoryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid category id"})
		return
	}

	err = h.service.Delete(c.Request.Context(), id, accountBook.ID, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCategoryPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, repository.ErrCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to delete category"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h CategoriesHandler) Icons(c *gin.Context) {
	items, err := h.service.ListCategoryIcons()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to load category icons"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h CategoriesHandler) Update(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	id, err := parseCategoryID(c.Param("id"))
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
		case errors.Is(err, service.ErrCategoryPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, service.ErrCategoryInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "分类名不能为空哦~"})
		case errors.Is(err, repository.ErrCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to update category"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h CategoriesHandler) Create(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	currentUser, ok := requireCurrentUser(c)
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
		case errors.Is(err, service.ErrCategoryPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 401, "msg": "您无权限进行此操作~"})
		case errors.Is(err, service.ErrCategoryInvalidInput):
			c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "分类名不能为空哦~"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to create category"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func parseCategoryID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidParam("id")
	}
	return id, nil
}

func parseCategoryWriteInput(c *gin.Context) (service.CategoryWriteInput, error) {
	var req httpdto.CategoryWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return service.CategoryWriteInput{}, err
	}
	return service.CategoryWriteInput{
		Name:     req.Category.Name,
		ParentID: req.Category.ParentID,
		IconPath: req.Category.IconPath,
		Type:     req.Category.Type,
	}, nil
}
