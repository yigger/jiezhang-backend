package handler

import "github.com/gin-gonic/gin"

type StatementsHandler struct{}

func NewStatementsHandler() StatementsHandler {
	return StatementsHandler{}
}

func (h StatementsHandler) Categories(c *gin.Context) {
	notImplemented(c, "GET /api/statements/categories")
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
	notImplemented(c, "GET /api/statements")
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
