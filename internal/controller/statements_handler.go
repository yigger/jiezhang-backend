package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	exportservice "github.com/yigger/jiezhang-backend/internal/service/export"
	sharingservice "github.com/yigger/jiezhang-backend/internal/service/sharing"
	statementservice "github.com/yigger/jiezhang-backend/internal/service/statement"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type StatementsHandler struct {
	reader   *statementservice.Reader
	writer   *statementservice.Writer
	sharing  *sharingservice.Service
	exporter *exportservice.Service
}

func NewStatementsHandler(reader *statementservice.Reader, writer *statementservice.Writer, sharing *sharingservice.Service, exporter *exportservice.Service) StatementsHandler {
	return StatementsHandler{reader: reader, writer: writer, sharing: sharing, exporter: exporter}
}

// Categories 获取分类列表，Query: `type`（默认 expend）
// @Summary 获取分类列表，Query: `type`（默认 expend）
// @ID StatementsHandler_Categories
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {object} types.StatementCategoriesResult "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/categories [get]
func (h StatementsHandler) Categories(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	statementType := c.Query("type")
	if statementType == "" {
		statementType = "expend"
	}

	input := statementservice.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          statementType,
	}
	categories, err := h.reader.GetCategories(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// Assets 获取资产列表，Query: `type`
// @Summary 获取资产列表，Query: `type`
// @ID StatementsHandler_Assets
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {object} types.StatementAssetsResult "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/assets [get]
func (h StatementsHandler) Assets(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	filter := statementservice.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	}

	assets, err := h.reader.GetAssets(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assets"})
		return
	}
	c.JSON(http.StatusOK, assets)
}

// CategoryFrequent 常用分类，Query: `type`
// @Summary 常用分类，Query: `type`
// @ID StatementsHandler_CategoryFrequent
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {array} types.StatementFrequentCategoryItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/category_frequent [get]
func (h StatementsHandler) CategoryFrequent(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	filter := statementservice.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	}

	categories, err := h.reader.CategoriesGuess(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get frequent categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// AssetFrequent 常用资产，Query: `type`
// @Summary 常用资产，Query: `type`
// @ID StatementsHandler_AssetFrequent
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {array} types.StatementFrequentAssetItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/asset_frequent [get]
func (h StatementsHandler) AssetFrequent(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	assets, err := h.reader.AssetsGuess(c.Request.Context(), statementservice.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get frequent assets"})
		return
	}
	c.JSON(http.StatusOK, assets)
}

// Search 搜索账单，Query: `keyword`
// @Summary 搜索账单，Query: `keyword`
// @ID StatementsHandler_Search
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param keyword query string false "keyword"
// @Success 200 {array} types.StatementListItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /search [get]
func (h StatementsHandler) Search(c *gin.Context) {
	accountBook, _ := RequireAccountBook(c)
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	statements, err := h.reader.SearchStatements(c.Request.Context(), accountBook.ID, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statements)

}

// List 账单列表，Query: `start_date`, `end_date`, `limit`, `offset`, `category_ids`, `except_ids`, `order_by`
// @Summary 账单列表，Query: `start_date`, `end_date`, `limit`, `offset`, `category_ids`, `except_ids`, `order_by`
// @ID StatementsHandler_List
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param category_ids query string false "category_ids"
// @Param end_date query string false "end_date"
// @Param except_ids query string false "except_ids"
// @Param limit query string false "limit"
// @Param offset query string false "offset"
// @Param order_by query string false "order_by"
// @Param start_date query string false "start_date"
// @Success 200 {array} types.StatementListItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /statements [get]
func (h StatementsHandler) List(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	input, err := buildStatementListInput(c, currentUser.ID, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statements, err := h.reader.GetStatements(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statements"})
		return
	}

	c.JSON(http.StatusOK, statements)
}

// ListByToken 通过分享 token 查看账单，Query: `token`, `order_by`
// @Summary 通过分享 token 查看账单，Query: `token`, `order_by`
// @ID StatementsHandler_ListByToken
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param order_by query string false "order_by"
// @Param token query string false "token"
// @Success 200 {object} types.APIResponse{data=types.StatementListByTokenResult} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/list_by_token [get]
func (h StatementsHandler) ListByToken(c *gin.Context) {
	res, err := h.sharing.ListByToken(c.Request.Context(), sharingservice.StatementListByTokenInput{
		Token:   c.Query("token"),
		OrderBy: c.Query("order_by"),
	})
	if err != nil {
		switch {
		case errors.Is(err, sharingservice.ErrStatementDecodeFailed):
			c.JSON(http.StatusOK, gin.H{"status": 501, "msg": "解码失败"})
		case errors.Is(err, sharingservice.ErrStatementInvalidToken):
			c.JSON(http.StatusOK, gin.H{"status": 502, "msg": "token 无效"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to list by token"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

// Create 创建账单，Body: `{statement: {type, amount, description, mood, category_id, asset_id, from_asset_id, to_asset_id, payee_id, target_object, location, nation, province, city, district, street, date, time}}`，amount 支持数字或字符串
// @Summary 创建账单，Body: `{statement: {type, amount, description, mood, category_id, asset_id, from_asset_id, to_asset_id, payee_id, target_object, location, nation, province, city, district, street, date, time}}`，amount 支持数字或字符串
// @ID StatementsHandler_Create
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Accept json
// @Param body body types.StatementWriteRequest true "请求体"
// @Success 200 {object} types.APIResponse{data=types.StatementListItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /statements [post]
func (h StatementsHandler) Create(c *gin.Context) {
	currentUser, _ := RequireCurrentUser(c)
	accountBook, _ := RequireAccountBook(c)

	input, err := buildStatementWriteInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.UserID = currentUser.ID
	input.AccountBookID = accountBook.ID

	statement, err := h.writer.CreateStatement(c.Request.Context(), input)
	if err != nil {
		var ve statementservice.ValidateError
		if errors.Is(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "error": ve.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"data":   statement,
	})
}

// Update 更新账单，Body 同上但所有字段可选（指针）
// @Summary 更新账单，Body 同上但所有字段可选（指针）
// @ID StatementsHandler_Update
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param statementId path integer true "statementId"
// @Accept json
// @Param body body types.StatementPatchRequest true "请求体"
// @Success 200 {object} types.APIResponse{data=types.StatementListItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /statements/{statementId} [put]
func (h StatementsHandler) Update(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	statementID, err := parseStatementID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid statementId"})
		return
	}

	patch, err := buildStatementPatchInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statement, err := h.writer.UpdateStatement(c.Request.Context(), statementservice.UpdateInput{
		StatementID:   statementID,
		UserID:        currentUser.ID,
		AccountBookID: accountBook.ID,
		Patch:         patch,
	})
	if err != nil {
		switch {
		case errors.Is(err, statementservice.ErrStatementPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 500, "msg": "不能更改他人账单哦"})
		case errors.Is(err, statementservice.ErrStatementInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "error": "invalid statement"})
		case errors.Is(err, statementservice.ErrRepositoryStatementNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "error": "statement not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to update statement"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": statement})
}

// Show 账单详情（含 upload_files、target_asset、can_edit 等）
// @Summary 账单详情（含 upload_files、target_asset、can_edit 等）
// @ID StatementsHandler_Show
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param statementId path integer true "statementId"
// @Success 200 {object} types.StatementDetailItem "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Router /statements/{statementId} [get]
func (h StatementsHandler) Show(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, _ := RequireAccountBook(c)
	statementID, err := parseStatementID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid statementId"})
		return
	}

	statement, err := h.reader.GetStatementByID(c.Request.Context(), statementID, accountBook.ID, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500})
		return
	}

	c.JSON(http.StatusOK, statement)
}

// Delete 删除账单
// @Summary 删除账单
// @ID StatementsHandler_Delete
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param statementId path integer true "statementId"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /statements/{statementId} [delete]
func (h StatementsHandler) Delete(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	statementID, err := parseStatementID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid statementId"})
		return
	}

	err = h.writer.DeleteStatement(c.Request.Context(), statementID, currentUser.ID, accountBook.ID)
	if err != nil {
		switch {
		case errors.Is(err, statementservice.ErrStatementPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 500, "msg": "只能删除自己创建的账单"})
		case errors.Is(err, statementservice.ErrRepositoryStatementNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "error": "statement not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to delete statement"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// Images 账单图片时间轴
// @Summary 账单图片时间轴
// @ID StatementsHandler_Images
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse{data=types.StatementImagesResult} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/images [get]
func (h StatementsHandler) Images(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.reader.GetImages(c.Request.Context(), accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to get statement images"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

// GenerateShareKey 生成分享链接，Body: `{start_date, end_date, category_ids, except_statement_ids}`
// @Summary 生成分享链接，Body: `{start_date, end_date, category_ids, except_statement_ids}`
// @ID StatementsHandler_GenerateShareKey
// @Tags Statements
// @Produce json
// @Accept json
// @Param body body types.ShareKeyRequest true "请求体"
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse{data=types.ShareKeyData} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/generate_share_key [post]
func (h StatementsHandler) GenerateShareKey(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	var req types.ShareKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}
	exceptIDs := strings.TrimSpace(req.ExceptStatementIDs)
	if exceptIDs == "" {
		exceptIDs = strings.TrimSpace(req.ExceptedStatementIDs)
	}

	token, err := h.sharing.GenerateShareKey(c.Request.Context(), sharingservice.StatementGenerateShareKeyInput{
		AccountBookID:      accountBook.ID,
		UserID:             currentUser.ID,
		StartDate:          strings.TrimSpace(req.StartDate),
		EndDate:            strings.TrimSpace(req.EndDate),
		CategoryIDs:        strings.TrimSpace(req.CategoryIDs),
		ExceptStatementIDs: exceptIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to generate share key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": gin.H{"share_key": token}})
}

// ExportCheck 导出次数检查
// @Summary 导出次数检查
// @ID StatementsHandler_ExportCheck
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/export_check [post]
func (h StatementsHandler) ExportCheck(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}

	_, err := h.exporter.ExportCheck(c.Request.Context(), currentUser.ID)
	if err != nil {
		if errors.Is(err, exportservice.ErrStatementExportLimited) {
			c.JSON(http.StatusOK, gin.H{"status": 503, "msg": "今日导出次数已达上限，请明天再试"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to check export limit"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// TargetObjects 获取目标对象列表，Query: `type`
// @Summary 获取目标对象列表，Query: `type`
// @ID StatementsHandler_TargetObjects
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {object} types.APIResponse{data=[]string} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/target_objects [get]
func (h StatementsHandler) TargetObjects(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	targetObjects, err := h.reader.GetTargetObjects(c.Request.Context(), statementservice.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get target objects"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": targetObjects})
}

// RemoveAvatar 删除账单附件图片，Body: `{avatar_id}`
// @Summary 删除账单附件图片，Body: `{avatar_id}`
// @ID StatementsHandler_RemoveAvatar
// @Tags Statements
// @Produce json
// @Accept json
// @Param body body types.AvatarDeleteRequest false "请求体"
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param statementId path integer true "statementId"
// @Param avatar_id query string false "avatar_id"
// @Success 200 {object} types.APIResponse "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Failure 400 {object} types.APIResponse "请求失败"
// @Failure 404 {object} types.APIResponse "请求失败"
// @Router /statements/{statementId}/avatar [delete]
func (h StatementsHandler) RemoveAvatar(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	statementID, err := parseStatementID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid statementId"})
		return
	}

	var req types.AvatarDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AvatarID <= 0 {
		avatarIDRaw := strings.TrimSpace(c.Query("avatar_id"))
		if avatarIDRaw == "" {
			avatarIDRaw = strings.TrimSpace(c.PostForm("avatar_id"))
		}
		if avatarIDRaw != "" {
			if v, parseErr := strconv.ParseInt(avatarIDRaw, 10, 64); parseErr == nil && v > 0 {
				req.AvatarID = v
			}
		}
		if req.AvatarID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid avatar_id"})
			return
		}
	}

	err = h.writer.RemoveAvatar(c.Request.Context(), accountBook.ID, statementID, req.AvatarID)
	if err != nil {
		switch {
		case errors.Is(err, statementservice.ErrRepositoryStatementNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "账单不存在或已删除"})
		case errors.Is(err, statementservice.ErrRepositoryStatementAvatarNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "图片不存在或已删除"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to remove avatar"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

// DefaultCategoryAsset 上次使用的分类和资产，Query: `type`
// @Summary 上次使用的分类和资产，Query: `type`
// @ID StatementsHandler_DefaultCategoryAsset
// @Tags Statements
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param type query string false "type"
// @Success 200 {object} types.APIResponse{data=types.StatementDefaultCategoryAssetItem} "成功；兼容接口也可能在 HTTP 200 的 status/msg 中返回业务错误"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/default_category_asset [get]
func (h StatementsHandler) DefaultCategoryAsset(c *gin.Context) {
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	item, err := h.reader.GetDefaultCategoryAsset(c.Request.Context(), statementservice.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get default category asset"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// ExportExcel 导出 Excel，Query: `range`，返回 .xlsx 文件
// @Summary 导出 Excel，Query: `range`，返回 .xlsx 文件
// @ID StatementsHandler_ExportExcel
// @Tags Statements
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账本 ID；未提供时使用用户默认账本"
// @Param range query string false "range"
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success 200 {file} file "Excel 文件"
// @Failure 500 {object} types.APIResponse "服务错误"
// @Router /statements/export_excel [get]
func (h StatementsHandler) ExportExcel(c *gin.Context) {
	currentUser, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := RequireAccountBook(c)
	if !ok {
		return
	}

	content, err := h.exporter.ExportExcelFile(c.Request.Context(), exportservice.StatementExportInput{
		AccountBookID: accountBook.ID,
		UserID:        currentUser.ID,
		Range:         strings.TrimSpace(c.Query("range")),
	})
	if err != nil {
		if errors.Is(err, exportservice.ErrStatementExportLimited) {
			c.JSON(http.StatusOK, gin.H{"status": 503, "msg": "今日导出次数已达上限，请明天再试"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to export excel"})
		return
	}

	filename := "statements_" + time.Now().Format("20060102_150405") + ".xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", content)
}

func buildStatementListInput(c *gin.Context, userID int64, accountBookID int64) (statementservice.ListInput, error) {
	var startDate *time.Time
	var endDate *time.Time
	var err error

	if v := strings.TrimSpace(c.Query("start_date")); v != "" {
		t, parseErr := parseFlexibleDateTime(v)
		if parseErr != nil {
			return statementservice.ListInput{}, parseErr
		}
		startDate = &t
	}
	if v := strings.TrimSpace(c.Query("end_date")); v != "" {
		t, parseErr := parseFlexibleDateTime(v)
		if parseErr != nil {
			return statementservice.ListInput{}, parseErr
		}
		endDate = &t
	}

	limit := 50
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil || limit <= 0 {
			return statementservice.ListInput{}, ErrInvalidParam("limit")
		}
	}

	offset := 0
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		offset, err = strconv.Atoi(v)
		if err != nil || offset < 0 {
			return statementservice.ListInput{}, ErrInvalidParam("offset")
		}
	}

	parentCategoryIDs, err := parseCSVInt64(c.Query("category_ids"))
	if err != nil {
		return statementservice.ListInput{}, ErrInvalidParam("category_ids")
	}

	exceptIDs, err := parseCSVInt64(c.Query("except_ids"))
	if err != nil {
		return statementservice.ListInput{}, ErrInvalidParam("except_ids")
	}

	return statementservice.ListInput{
		UserID:            userID,
		AccountBookID:     accountBookID,
		StartDate:         startDate,
		EndDate:           endDate,
		ParentCategoryIDs: parentCategoryIDs,
		ExceptIDs:         exceptIDs,
		OrderBy:           strings.TrimSpace(c.Query("order_by")),
		Limit:             limit,
		Offset:            offset,
	}, nil
}

func parseFlexibleDateTime(v string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrInvalidParam("date")
}

func parseCSVInt64(v string) ([]int64, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}

	parts := strings.Split(v, ",")
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			return nil, ErrInvalidParam("csv int ids")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func buildStatementWriteInput(c *gin.Context) (statementservice.WriteInput, error) {
	var req types.StatementWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return statementservice.WriteInput{}, err
	}

	p := req.Statement

	amount, err := strconv.ParseFloat(string(p.Amount), 64)
	if err != nil {
		return statementservice.WriteInput{}, ErrInvalidParam("amount")
	}

	return statementservice.WriteInput{
		ProjectID: p.ProjectID, ConsumerID: p.ConsumerID,
		Type:         p.Type,
		Amount:       amount,
		Description:  p.Description,
		Mood:         p.Mood,
		CategoryID:   p.CategoryID,
		AssetID:      p.AssetID,
		FromAssetID:  p.FromAssetID,
		ToAssetID:    p.ToAssetID,
		PayeeID:      p.PayeeID,
		TargetObject: p.TargetObject,
		Location:     p.Location,
		Nation:       p.Nation,
		Province:     p.Province,
		City:         p.City,
		District:     p.District,
		Street:       p.Street,
		Date:         p.Date,
		Time:         p.Time,
	}, nil
}

func buildStatementPatchInput(c *gin.Context) (statementservice.PatchInput, error) {
	var req types.StatementPatchRequest
	// 不用指针时，ShouldBindJSON 后 1 和 2 会混在一起，Go 看起来都像零值，没法判断“用户是没传，还是故意要改成零值”。
	if err := c.ShouldBindJSON(&req); err != nil {
		return statementservice.PatchInput{}, err
	}

	p := req.Statement
	input := statementservice.PatchInput{
		Type:         p.Type,
		Description:  p.Description,
		Mood:         p.Mood,
		CategoryID:   p.CategoryID,
		AssetID:      p.AssetID,
		FromAssetID:  p.FromAssetID,
		ToAssetID:    p.ToAssetID,
		PayeeID:      p.PayeeID,
		TargetObject: p.TargetObject,
		Location:     p.Location,
		Nation:       p.Nation,
		Province:     p.Province,
		City:         p.City,
		District:     p.District,
		Street:       p.Street,
		Date:         p.Date,
		Time:         p.Time,
	}

	if p.Amount != nil {
		amount, err := strconv.ParseFloat(strings.TrimSpace(string(*p.Amount)), 64)
		if err != nil {
			return statementservice.PatchInput{}, ErrInvalidParam("amount")
		}
		input.Amount = &amount
	}
	return input, nil
}

func parseStatementID(c *gin.Context) (int64, error) {
	v := strings.TrimSpace(c.Param("statementId"))
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidParam("statementId")
	}
	return id, nil
}
