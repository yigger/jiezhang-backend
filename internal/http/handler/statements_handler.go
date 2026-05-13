package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	httpdto "github.com/yigger/jiezhang-backend/internal/http/dto"
	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

type StatementsHandler struct {
	service service.StatementService
}

func NewStatementsHandler(service service.StatementService) StatementsHandler {
	return StatementsHandler{service: service}
}

func (h StatementsHandler) Categories(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	statementType := c.Query("type")
	if statementType == "" {
		statementType = "expend"
	}

	input := service.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          statementType,
	}
	categories, err := h.service.GetCategories(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h StatementsHandler) Assets(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	filter := service.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	}

	assets, err := h.service.GetAssets(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assets"})
		return
	}
	c.JSON(http.StatusOK, assets)
}

func (h StatementsHandler) CategoryFrequent(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	filter := service.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	}

	categories, err := h.service.CategoriesGuess(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get frequent categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h StatementsHandler) AssetFrequent(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	assets, err := h.service.AssetsGuess(c.Request.Context(), service.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get frequent assets"})
		return
	}
	c.JSON(http.StatusOK, assets)
}

func (h StatementsHandler) Search(c *gin.Context) {
	accountBook, _ := requireAccountBook(c)
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	statements, err := h.service.SearchStatements(c.Request.Context(), accountBook.ID, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statements)

}

func (h StatementsHandler) List(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	input, err := buildStatementListInput(c, currentUser.ID, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statements, err := h.service.GetStatements(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statements"})
		return
	}

	c.JSON(http.StatusOK, statements)
}

func (h StatementsHandler) ListByToken(c *gin.Context) {
	res, err := h.service.ListByToken(c.Request.Context(), service.StatementListByTokenInput{
		Token:   c.Query("token"),
		OrderBy: c.Query("order_by"),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatementDecodeFailed):
			c.JSON(http.StatusOK, gin.H{"status": 501, "msg": "解码失败"})
		case errors.Is(err, service.ErrStatementInvalidToken):
			c.JSON(http.StatusOK, gin.H{"status": 502, "msg": "token 无效"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to list by token"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

func (h StatementsHandler) Create(c *gin.Context) {
	currentUser, _ := requireCurrentUser(c)
	accountBook, _ := requireAccountBook(c)

	input, err := buildStatementWriteInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.UserID = currentUser.ID
	input.AccountBookID = accountBook.ID

	statement, err := h.service.CreateStatement(c.Request.Context(), input)
	if err != nil {
		var ve service.ValidateError
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

func (h StatementsHandler) Update(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
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

	statement, err := h.service.UpdateStatement(c.Request.Context(), statementdto.UpdateInput{
		StatementID:   statementID,
		UserID:        currentUser.ID,
		AccountBookID: accountBook.ID,
		Patch:         patch,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatementPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 500, "msg": "不能更改他人账单哦"})
		case errors.Is(err, service.ErrStatementInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "error": "invalid statement"})
		case errors.Is(err, repository.ErrStatementNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "error": "statement not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to update statement"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": statement})
}

func (h StatementsHandler) Show(c *gin.Context) {
	accountBook, _ := requireAccountBook(c)
	statementID, err := parseStatementID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid statementId"})
		return
	}

	statement, err := h.service.GetStatementByID(c.Request.Context(), statementID, accountBook.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500})
		return
	}

	c.JSON(http.StatusOK, statement)
}

func (h StatementsHandler) Delete(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	statementID, err := parseStatementID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid statementId"})
		return
	}

	err = h.service.DeleteStatement(c.Request.Context(), statementID, currentUser.ID, accountBook.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatementPermissionDenied):
			c.JSON(http.StatusOK, gin.H{"status": 500, "msg": "只能删除自己创建的账单"})
		case errors.Is(err, repository.ErrStatementNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "error": "statement not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to delete statement"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h StatementsHandler) Images(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	res, err := h.service.GetImages(c.Request.Context(), accountBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to get statement images"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": res})
}

func (h StatementsHandler) GenerateShareKey(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	var req struct {
		StartDate            string `json:"start_date"`
		EndDate              string `json:"end_date"`
		CategoryIDs          string `json:"category_ids"`
		ExceptedStatementIDs string `json:"exceptedStatementIds"`
		ExceptStatementIDs   string `json:"except_statement_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "参数错误"})
		return
	}
	exceptIDs := strings.TrimSpace(req.ExceptStatementIDs)
	if exceptIDs == "" {
		exceptIDs = strings.TrimSpace(req.ExceptedStatementIDs)
	}

	token, err := h.service.GenerateShareKey(c.Request.Context(), service.StatementGenerateShareKeyInput{
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

func (h StatementsHandler) ExportCheck(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}

	_, err := h.service.ExportCheck(c.Request.Context(), currentUser.ID)
	if err != nil {
		if errors.Is(err, service.ErrStatementExportLimited) {
			c.JSON(http.StatusOK, gin.H{"status": 503, "msg": "今日导出次数已达上限，请明天再试"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to check export limit"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h StatementsHandler) TargetObjects(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	targetObjects, err := h.service.GetTargetObjects(c.Request.Context(), service.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get target objects"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": targetObjects})
}

func (h StatementsHandler) RemoveAvatar(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	statementID, err := parseStatementID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "invalid statementId"})
		return
	}

	var req struct {
		AvatarID int64 `json:"avatar_id"`
	}
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

	err = h.service.RemoveAvatar(c.Request.Context(), accountBook.ID, statementID, req.AvatarID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrStatementNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "账单不存在或已删除"})
		case errors.Is(err, repository.ErrStatementAvatarNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "图片不存在或已删除"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "msg": "failed to remove avatar"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
}

func (h StatementsHandler) DefaultCategoryAsset(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	item, err := h.service.GetDefaultCategoryAsset(c.Request.Context(), service.GetCategoriesInput{
		AccountBookID: accountBook.ID,
		Type:          c.Query("type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get default category asset"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h StatementsHandler) ExportExcel(c *gin.Context) {
	currentUser, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	content, err := h.service.ExportExcelFile(c.Request.Context(), service.StatementExportInput{
		AccountBookID: accountBook.ID,
		UserID:        currentUser.ID,
		Range:         strings.TrimSpace(c.Query("range")),
	})
	if err != nil {
		if errors.Is(err, service.ErrStatementExportLimited) {
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

func buildStatementListInput(c *gin.Context, userID int64, accountBookID int64) (statementdto.ListInput, error) {
	var startDate *time.Time
	var endDate *time.Time
	var err error

	if v := strings.TrimSpace(c.Query("start_date")); v != "" {
		t, parseErr := parseFlexibleDateTime(v)
		if parseErr != nil {
			return statementdto.ListInput{}, parseErr
		}
		startDate = &t
	}
	if v := strings.TrimSpace(c.Query("end_date")); v != "" {
		t, parseErr := parseFlexibleDateTime(v)
		if parseErr != nil {
			return statementdto.ListInput{}, parseErr
		}
		endDate = &t
	}

	limit := 50
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil || limit <= 0 {
			return statementdto.ListInput{}, errInvalidParam("limit")
		}
	}

	offset := 0
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		offset, err = strconv.Atoi(v)
		if err != nil || offset < 0 {
			return statementdto.ListInput{}, errInvalidParam("offset")
		}
	}

	parentCategoryIDs, err := parseCSVInt64(c.Query("category_ids"))
	if err != nil {
		return statementdto.ListInput{}, errInvalidParam("category_ids")
	}

	exceptIDs, err := parseCSVInt64(c.Query("except_ids"))
	if err != nil {
		return statementdto.ListInput{}, errInvalidParam("except_ids")
	}

	return statementdto.ListInput{
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
	return time.Time{}, errInvalidParam("date")
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
			return nil, errInvalidParam("csv int ids")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

type invalidParamError struct {
	field string
}

func (e invalidParamError) Error() string {
	return "invalid parameter: " + e.field
}

func errInvalidParam(field string) error {
	return invalidParamError{field: field}
}

func buildStatementWriteInput(c *gin.Context) (statementdto.WriteInput, error) {
	var req httpdto.StatementWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return statementdto.WriteInput{}, err
	}

	p := req.Statement

	amount, err := strconv.ParseFloat(p.Amount, 64)
	if err != nil {
		return statementdto.WriteInput{}, errInvalidParam("amount")
	}

	return statementdto.WriteInput{
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

func buildStatementPatchInput(c *gin.Context) (statementdto.PatchInput, error) {
	var req httpdto.StatementPatchRequest
	// 不用指针时，ShouldBindJSON 后 1 和 2 会混在一起，Go 看起来都像零值，没法判断“用户是没传，还是故意要改成零值”。
	if err := c.ShouldBindJSON(&req); err != nil {
		return statementdto.PatchInput{}, err
	}

	p := req.Statement
	input := statementdto.PatchInput{
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
		amount, err := strconv.ParseFloat(strings.TrimSpace(*p.Amount), 64)
		if err != nil {
			return statementdto.PatchInput{}, errInvalidParam("amount")
		}
		input.Amount = &amount
	}
	return input, nil
}

func parseStatementID(c *gin.Context) (int64, error) {
	v := strings.TrimSpace(c.Param("statementId"))
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidParam("statementId")
	}
	return id, nil
}
