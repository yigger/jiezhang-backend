package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/service"
)

type StatementsHandler struct {
	service service.StatementService
}

func NewStatementsHandler(service service.StatementService) StatementsHandler {
	return StatementsHandler{service: service}
}

func (h StatementsHandler) Categories(c *gin.Context) {
	var _ domain.User
	var accountBook domain.AccountBook
	var ok bool
	if _, accountBook, ok = fetchCurrentUser(c); !ok {
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

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func (h StatementsHandler) Assets(c *gin.Context) {
	notImplemented(c, "GET /api/statements/assets")
}

func (h StatementsHandler) CategoryFrequent(c *gin.Context) {
	notImplemented(c, "GET /api/statements/category_frequent")
}

func (h StatementsHandler) AssetFrequent(c *gin.Context) {
	notImplemented(c, "GET /api/statements/asset_frequent")
}

func (h StatementsHandler) List(c *gin.Context) {
	var currentUser domain.User
	var ok bool
	if currentUser, _, ok = fetchCurrentUser(c); !ok {
		return
	}

	input, err := buildStatementListInput(c, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statements, err := h.service.GetStatements(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": statements})
}

func (h StatementsHandler) ListByToken(c *gin.Context) {
	notImplemented(c, "GET /api/statements/list_by_token")
}

func (h StatementsHandler) Create(c *gin.Context) {
	notImplemented(c, "POST /api/statements")
}

func (h StatementsHandler) Update(c *gin.Context) {
	notImplemented(c, "PUT /api/statements/:statementId")
}

func (h StatementsHandler) Show(c *gin.Context) {
	notImplemented(c, "GET /api/statements/:statementId")
}

func (h StatementsHandler) Delete(c *gin.Context) {
	notImplemented(c, "DELETE /api/statements/:statementId")
}

func (h StatementsHandler) Search(c *gin.Context) {
	notImplemented(c, "GET /api/search")
}

func (h StatementsHandler) Images(c *gin.Context) {
	notImplemented(c, "GET /api/statements/images")
}

func (h StatementsHandler) GenerateShareKey(c *gin.Context) {
	notImplemented(c, "POST /api/statements/generate_share_key")
}

func (h StatementsHandler) ExportCheck(c *gin.Context) {
	notImplemented(c, "POST /api/statements/export_check")
}

func (h StatementsHandler) TargetObjects(c *gin.Context) {
	notImplemented(c, "GET /api/statements/target_objects")
}

func (h StatementsHandler) RemoveAvatar(c *gin.Context) {
	notImplemented(c, "DELETE /api/statements/:statementId/avatar")
}

func (h StatementsHandler) DefaultCategoryAsset(c *gin.Context) {
	notImplemented(c, "GET /api/statements/default_category_asset")
}

func (h StatementsHandler) ExportExcel(c *gin.Context) {
	notImplemented(c, "GET /api/statements/export_excel")
}

func buildStatementListInput(c *gin.Context, userID int64) (service.StatementListInput, error) {
	var startDate *time.Time
	var endDate *time.Time
	var err error

	if v := strings.TrimSpace(c.Query("start_date")); v != "" {
		t, parseErr := parseFlexibleDateTime(v)
		if parseErr != nil {
			return service.StatementListInput{}, parseErr
		}
		startDate = &t
	}
	if v := strings.TrimSpace(c.Query("end_date")); v != "" {
		t, parseErr := parseFlexibleDateTime(v)
		if parseErr != nil {
			return service.StatementListInput{}, parseErr
		}
		endDate = &t
	}

	accountBookID := int64(0)
	if v := strings.TrimSpace(c.Query("account_book_id")); v != "" {
		accountBookID, err = strconv.ParseInt(v, 10, 64)
		if err != nil || accountBookID < 0 {
			return service.StatementListInput{}, errInvalidParam("account_book_id")
		}
	}

	limit := 50
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil || limit <= 0 {
			return service.StatementListInput{}, errInvalidParam("limit")
		}
	}

	offset := 0
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		offset, err = strconv.Atoi(v)
		if err != nil || offset < 0 {
			return service.StatementListInput{}, errInvalidParam("offset")
		}
	}

	parentCategoryIDs, err := parseCSVInt64(c.Query("category_ids"))
	if err != nil {
		return service.StatementListInput{}, errInvalidParam("category_ids")
	}

	exceptIDs, err := parseCSVInt64(c.Query("except_ids"))
	if err != nil {
		return service.StatementListInput{}, errInvalidParam("except_ids")
	}

	return service.StatementListInput{
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
