package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/repository"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

type ValidateError struct {
	Message string
}

func (e ValidateError) Error() string {
	return fmt.Sprintf("validate error: %s", e.Message)
}

type StatementService struct {
	statementRepo repository.StatementRepository
	queryRepo     repository.StatementQueryRepository
	categoryRepo  repository.CategoryRepository
	assetRepo     repository.AssetRepository
	rowMapper     statementdto.RowMapper
	userRepo      repository.UserRepository
	cache         sessioncache.Cache
	tokenSecret   string
}

func NewStatementService(statementRepo repository.StatementRepository, queryRepo repository.StatementQueryRepository, categoryRepo repository.CategoryRepository, assetRepo repository.AssetRepository, rowMapper statementdto.RowMapper) StatementService {
	return StatementService{
		statementRepo: statementRepo,
		queryRepo:     queryRepo,
		categoryRepo:  categoryRepo,
		assetRepo:     assetRepo,
		rowMapper:     rowMapper,
	}
}

func NewStatementServiceWithSession(
	statementRepo repository.StatementRepository,
	queryRepo repository.StatementQueryRepository,
	categoryRepo repository.CategoryRepository,
	assetRepo repository.AssetRepository,
	userRepo repository.UserRepository,
	cache sessioncache.Cache,
	tokenSecret string,
	rowMapper statementdto.RowMapper,
) StatementService {
	svc := NewStatementService(statementRepo, queryRepo, categoryRepo, assetRepo, rowMapper)
	svc.userRepo = userRepo
	svc.cache = cache
	svc.tokenSecret = strings.TrimSpace(tokenSecret)
	return svc
}

var (
	ErrStatementPermissionDenied = errors.New("statement permission denied")
	ErrStatementInvalidInput     = errors.New("statement invalid input")
	ErrStatementInvalidToken     = errors.New("statement share token invalid")
	ErrStatementDecodeFailed     = errors.New("statement share token decode failed")
	ErrStatementExportLimited    = errors.New("statement export daily limit reached")
)

type StatementShareTokenPayload struct {
	AccountBookID      int64  `json:"account_book_id"`
	UserID             int64  `json:"user_id"`
	StartDate          string `json:"start_date"`
	EndDate            string `json:"end_date"`
	CategoryIDs        string `json:"category_ids"`
	ExceptStatementIDs string `json:"except_statement_ids"`
}

type StatementGenerateShareKeyInput struct {
	AccountBookID      int64
	UserID             int64
	StartDate          string
	EndDate            string
	CategoryIDs        string
	ExceptStatementIDs string
}

type StatementListByTokenInput struct {
	Token   string
	OrderBy string
}

type StatementDateRangeItem struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type StatementSharedUserItem struct {
	Nickname   string `json:"nickname"`
	AvatarPath string `json:"avatar_path"`
}

type StatementListByTokenResult struct {
	Data       []statementdto.ListItem `json:"data"`
	DateRange  StatementDateRangeItem  `json:"date_range"`
	SharedUser StatementSharedUserItem `json:"shared_user"`
}

type StatementImageItem struct {
	StatementID int64  `json:"statement_id"`
	AvatarID    int64  `json:"avatar_id"`
	Path        string `json:"path"`
}

type StatementImageMonthGroup struct {
	Month int                  `json:"month"`
	Data  []StatementImageItem `json:"data"`
}

type StatementImageYearGroup struct {
	Year int                        `json:"year"`
	Data []StatementImageMonthGroup `json:"data"`
}

type StatementImagesResult struct {
	AvatarTimeline []StatementImageYearGroup `json:"avatar_timeline"`
	Avatars        []string                  `json:"avatars"`
}

type StatementExportCheckResult struct {
	TodayCount int `json:"today_count"`
}

type StatementExportInput struct {
	AccountBookID int64
	UserID        int64
	Range         string
}

type StatementExportRowItem struct {
	Category       string  `json:"category"`
	ParentCategory string  `json:"parent_category"`
	Type           string  `json:"type"`
	TypeName       string  `json:"type_name"`
	Asset          string  `json:"asset"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type StatementExportResult struct {
	Rows []StatementExportRowItem `json:"rows"`
}

// 账单列表
func (s StatementService) GetStatements(ctx context.Context, input statementdto.ListInput) ([]statementdto.ListItem, error) {
	filter := repository.StatementListFilter{
		UserID:            input.UserID,
		AccountBookID:     input.AccountBookID,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		ParentCategoryIDs: input.ParentCategoryIDs,
		ExceptIDs:         input.ExceptIDs,
		OrderBy:           input.OrderBy,
		Limit:             input.Limit,
		Offset:            input.Offset,
	}
	rows, err := s.queryRepo.ListRowsWithRelations(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]statementdto.ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s StatementService) SearchStatements(ctx context.Context, accountBookID int64, keyword string) ([]statementdto.ListItem, error) {
	filter := repository.StatementListFilter{
		AccountBookID: accountBookID,
		Keyword:       keyword,
		OrderBy:       "created_at desc",
	}
	rows, err := s.queryRepo.ListRowsWithRelations(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]statementdto.ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s StatementService) CreateStatement(ctx context.Context, input statementdto.WriteInput) (statementdto.ListItem, error) {
	record, err := normalizeStatementWriteInput(input)
	if err != nil {
		return statementdto.ListItem{}, err
	}

	statementID, err := s.statementRepo.Create(ctx, record)
	if err != nil {
		return statementdto.ListItem{}, err
	}

	row, err := s.queryRepo.GetRowByIDWithRelations(ctx, statementID, input.AccountBookID)
	if err != nil {
		return statementdto.ListItem{}, err
	}
	return s.rowMapper.ToListItem(row), nil
}

func (s StatementService) UpdateStatement(ctx context.Context, input statementdto.UpdateInput) (statementdto.ListItem, error) {
	ownerID, err := s.statementRepo.GetOwnerID(ctx, input.StatementID, input.AccountBookID)
	if err != nil {
		return statementdto.ListItem{}, err
	}
	if ownerID != input.UserID {
		return statementdto.ListItem{}, ErrStatementPermissionDenied
	}

	currentRow, err := s.queryRepo.GetRowByIDWithRelations(ctx, input.StatementID, input.AccountBookID)
	if err != nil {
		return statementdto.ListItem{}, err
	}

	merged := s.mergeStatementPatch(currentRow, input)
	record, err := normalizeStatementWriteInput(merged)
	if err != nil {
		return statementdto.ListItem{}, err
	}
	if err := s.statementRepo.UpdateByID(ctx, input.StatementID, input.AccountBookID, record); err != nil {
		return statementdto.ListItem{}, err
	}
	row, err := s.queryRepo.GetRowByIDWithRelations(ctx, input.StatementID, input.AccountBookID)
	if err != nil {
		return statementdto.ListItem{}, err
	}
	return s.rowMapper.ToListItem(row), nil
}

func (s StatementService) DeleteStatement(ctx context.Context, statementID int64, userID int64, accountBookID int64) error {
	ownerID, err := s.statementRepo.GetOwnerID(ctx, statementID, accountBookID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return ErrStatementPermissionDenied
	}
	return s.statementRepo.DeleteByID(ctx, statementID, accountBookID)
}

func normalizeStatementWriteInput(input statementdto.WriteInput) (repository.StatementWriteRecord, error) {
	statementType := strings.TrimSpace(input.Type)
	if statementType == "" {
		return repository.StatementWriteRecord{}, ValidateError{Message: "invalid statement type"}
	}
	if input.Amount <= 0 {
		return repository.StatementWriteRecord{}, ValidateError{Message: "invalid amount"}
	}

	occurredAt, err := parseStatementDateTime(input.Date, input.Time)
	if err != nil {
		return repository.StatementWriteRecord{}, ValidateError{Message: "invalid date or time"}
	}

	assetID := input.AssetID
	targetAssetID := int64PtrOrNil(input.ToAssetID)
	if statementType == "transfer" || statementType == "repayment" {
		assetID = input.FromAssetID
		if assetID <= 0 || input.ToAssetID <= 0 {
			return repository.StatementWriteRecord{}, ValidateError{Message: "invalid asset ID"}
		}
		targetAssetID = int64PtrOrNil(input.ToAssetID)
	}

	switch statementType {
	case "expend", "income":
		// 只有收入和支出需要检验分类和资产ID
		if assetID <= 0 || input.CategoryID <= 0 {
			return repository.StatementWriteRecord{}, ValidateError{Message: "invalid asset or category ID"}
		}
	case "transfer", "repayment":
		// 转账和还款需要检验转入和转出资产的 ID
		fromAssetID := input.FromAssetID
		toAssetID := input.ToAssetID
		if fromAssetID <= 0 || toAssetID <= 0 {
			return repository.StatementWriteRecord{}, ValidateError{Message: "invalid from or to asset ID"}
		}
	case "loan_in", "loan_out", "reimburse", "payment_proxy":
		// 借贷、报销、代付需要检验资产ID，分类ID可以为空
		if assetID <= 0 {
			return repository.StatementWriteRecord{}, ValidateError{Message: "invalid asset ID"}
		}
	default:
		return repository.StatementWriteRecord{}, ValidateError{Message: "invalid statement type"}
	}

	return repository.StatementWriteRecord{
		UserID:        input.UserID,
		AccountBookID: input.AccountBookID,
		Type:          statementType,
		Amount:        input.Amount,
		Description:   input.Description,
		Mood:          input.Mood,
		CategoryID:    input.CategoryID,
		AssetID:       assetID,
		TargetAssetID: targetAssetID,
		PayeeID:       int64PtrOrNil(input.PayeeID),
		TargetObject:  input.TargetObject,
		Location:      input.Location,
		Nation:        input.Nation,
		Province:      input.Province,
		City:          input.City,
		District:      input.District,
		Street:        input.Street,
		OccurredAt:    occurredAt,
	}, nil
}

func parseStatementDateTime(dateStr string, timeStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(timeStr)
	if dateStr == "" {
		now := time.Now()
		dateStr = now.Format("2006-01-02")
	}
	if timeStr == "" {
		timeStr = "00:00:00"
	}

	layouts := []string{"2006-01-02 15:04:05", "2006-01-02 15:04"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, dateStr+" "+timeStr, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ValidateError{Message: "invalid date or time"}
}

func int64PtrOrNil(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	n := v
	return &n
}

func (s StatementService) mergeStatementPatch(current repository.StatementListRowRecord, update statementdto.UpdateInput) statementdto.WriteInput {
	input := statementdto.WriteInput{
		StatementID:   update.StatementID,
		UserID:        update.UserID,
		AccountBookID: update.AccountBookID,
		Type:          current.Type,
		Amount:        current.Amount,
		Description:   current.Description,
		Mood:          current.Mood,
		CategoryID:    current.CategoryID,
		AssetID:       current.AssetID,
		FromAssetID:   current.AssetID,
		TargetObject:  current.TargetObject,
		Location:      current.Location,
		Nation:        current.Nation,
		Province:      current.Province,
		City:          current.City,
		District:      current.District,
		Street:        current.Street,
		Date:          current.CreatedAt.Format("2006-01-02"),
		Time:          current.CreatedAt.Format("15:04:05"),
	}
	if current.TargetAssetID > 0 {
		input.ToAssetID = current.TargetAssetID
	}
	if current.PayeeID > 0 {
		input.PayeeID = current.PayeeID
	}

	p := update.Patch
	if p.Type != nil {
		input.Type = strings.TrimSpace(*p.Type)
	}
	if p.Amount != nil {
		input.Amount = *p.Amount
	}
	if p.Description != nil {
		input.Description = *p.Description
	}
	if p.Mood != nil {
		input.Mood = *p.Mood
	}
	if p.CategoryID != nil {
		input.CategoryID = *p.CategoryID
	}
	if p.AssetID != nil {
		input.AssetID = *p.AssetID
	}
	if p.FromAssetID != nil {
		input.FromAssetID = *p.FromAssetID
	}
	if p.ToAssetID != nil {
		input.ToAssetID = *p.ToAssetID
	}
	if p.PayeeID != nil {
		input.PayeeID = *p.PayeeID
	}
	if p.TargetObject != nil {
		input.TargetObject = *p.TargetObject
	}
	if p.Location != nil {
		input.Location = *p.Location
	}
	if p.Nation != nil {
		input.Nation = *p.Nation
	}
	if p.Province != nil {
		input.Province = *p.Province
	}
	if p.City != nil {
		input.City = *p.City
	}
	if p.District != nil {
		input.District = *p.District
	}
	if p.Street != nil {
		input.Street = *p.Street
	}
	if p.Date != nil {
		input.Date = strings.TrimSpace(*p.Date)
	}
	if p.Time != nil {
		input.Time = strings.TrimSpace(*p.Time)
	}

	return input
}

type GetCategoriesInput struct {
	AccountBookID int64
	Type          string
}

type StatementFrequentCategoryItem struct {
	ID       int64                        `json:"id"`
	Name     string                       `json:"name"`
	IconPath string                       `json:"icon_path"`
	Parent   *StatementCategoryParentItem `json:"parent"`
}

type StatementCategoryParentItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementCategoryChildItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type StatementCategoryTreeItem struct {
	ID       int64                        `json:"id"`
	Name     string                       `json:"name"`
	IconPath string                       `json:"icon_path"`
	Childs   []StatementCategoryChildItem `json:"childs"`
}

type StatementCategoriesResult struct {
	Frequent   []StatementFrequentCategoryItem `json:"frequent"`
	Categories []StatementCategoryTreeItem     `json:"categories"`
}

func (s StatementService) GetCategories(ctx context.Context, input GetCategoriesInput) (StatementCategoriesResult, error) {
	filter := repository.CategoryListFilter{
		AccountBookID: input.AccountBookID,
		Type:          input.Type,
	}
	parents, err := s.categoryRepo.ListParents(ctx, filter)
	if err != nil {
		return StatementCategoriesResult{}, err
	}
	parentIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentIDs = append(parentIDs, p.ID)
	}

	children, err := s.categoryRepo.ListChildrenByParentIDs(ctx, filter, parentIDs)
	if err != nil {
		return StatementCategoriesResult{}, err
	}
	frequents, err := s.categoryRepo.ListFrequentChildren(ctx, filter, 10)
	if err != nil {
		return StatementCategoriesResult{}, err
	}

	childrenByParent := make(map[int64][]StatementCategoryChildItem, len(parents))
	for _, child := range children {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], StatementCategoryChildItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: s.rowMapper.BuildPublicURL(child.IconPath),
		})
	}

	categories := make([]StatementCategoryTreeItem, 0, len(parents))
	for _, p := range parents {
		childs := childrenByParent[p.ID]
		if childs == nil {
			childs = []StatementCategoryChildItem{}
		}
		categories = append(categories, StatementCategoryTreeItem{
			ID:       p.ID,
			Name:     p.Name,
			IconPath: s.rowMapper.BuildPublicURL(p.IconPath),
			Childs:   childs,
		})
	}

	frequentItems := make([]StatementFrequentCategoryItem, 0, len(frequents))
	for _, f := range frequents {
		var parent *StatementCategoryParentItem
		if f.HasParent {
			parent = &StatementCategoryParentItem{
				ID:   f.ParentID,
				Name: f.ParentName,
			}
		}
		frequentItems = append(frequentItems, StatementFrequentCategoryItem{
			ID:       f.ID,
			Name:     f.Name,
			IconPath: s.rowMapper.BuildPublicURL(f.IconPath),
			Parent:   parent,
		})
	}

	return StatementCategoriesResult{
		Frequent:   frequentItems,
		Categories: categories,
	}, nil
}

type StatementFrequentAssetItem struct {
	ID       int64                     `json:"id"`
	Name     string                    `json:"name"`
	IconPath string                    `json:"icon_path"`
	Parent   *StatementAssetParentItem `json:"parent"`
}

type StatementAssetParentItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementAssetChildItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type StatementAssetTreeItem struct {
	ID       int64                     `json:"id"`
	Name     string                    `json:"name"`
	IconPath string                    `json:"icon_path"`
	Childs   []StatementAssetChildItem `json:"childs"`
}

type StatementAssetsResult struct {
	Frequent   []StatementFrequentAssetItem `json:"frequent"`
	Categories []StatementAssetTreeItem     `json:"categories"`
}

type StatementDefaultCategoryAssetItem struct {
	CategoryName string `json:"category_name"`
	AssetName    string `json:"asset_name"`
	CategoryID   int64  `json:"category_id"`
	AssetID      int64  `json:"asset_id"`
}

func (s StatementService) GetAssets(ctx context.Context, input GetCategoriesInput) (StatementAssetsResult, error) {
	filter := repository.AssetFilter{
		AccountBookID: input.AccountBookID,
		Type:          input.Type,
	}

	var assetResult []StatementAssetTreeItem
	parents, err := s.assetRepo.ListParents(ctx, filter)
	if err != nil {
		return StatementAssetsResult{}, err
	}
	parentsIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentsIDs = append(parentsIDs, p.ID)
	}

	children, err := s.assetRepo.ListChildrenByParentIDs(ctx, filter, parentsIDs)
	if err != nil {
		return StatementAssetsResult{}, err
	}
	childrenByParent := make(map[int64][]StatementAssetChildItem, len(parents))
	for _, child := range children {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], StatementAssetChildItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: s.rowMapper.BuildPublicURL(child.IconPath),
		})
	}

	for _, p := range parents {
		childs := childrenByParent[p.ID]
		if childs == nil {
			childs = []StatementAssetChildItem{}
		}
		assetResult = append(assetResult, StatementAssetTreeItem{
			ID:       p.ID,
			Name:     p.Name,
			IconPath: s.rowMapper.BuildPublicURL(p.IconPath),
			Childs:   childs,
		})
	}

	frequentResult := make([]StatementFrequentAssetItem, 0)
	frequent, err := s.assetRepo.ListFrequentChildren(ctx, filter, 10)
	if err != nil {
		return StatementAssetsResult{}, err
	}
	for _, f := range frequent {
		var parent *StatementAssetParentItem
		if f.HasParent {
			parent = &StatementAssetParentItem{
				ID:   f.ParentID,
				Name: f.ParentName,
			}
		}
		frequentResult = append(frequentResult, StatementFrequentAssetItem{
			ID:       f.ID,
			Name:     f.Name,
			IconPath: s.rowMapper.BuildPublicURL(f.IconPath),
			Parent:   parent,
		})
	}
	return StatementAssetsResult{
		Frequent:   frequentResult,
		Categories: assetResult,
	}, nil
}

func (s StatementService) GetDefaultCategoryAsset(ctx context.Context, input GetCategoriesInput) (*StatementDefaultCategoryAssetItem, error) {
	record, err := s.queryRepo.GetLatestCategoryAssetByType(ctx, input.AccountBookID, input.Type)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return &StatementDefaultCategoryAssetItem{
		CategoryName: record.CategoryName,
		AssetName:    record.AssetName,
		CategoryID:   record.CategoryID,
		AssetID:      record.AssetID,
	}, nil
}

func (s StatementService) GetTargetObjects(ctx context.Context, input GetCategoriesInput) ([]string, error) {
	return s.queryRepo.ListDistinctTargetObjectsByType(ctx, input.AccountBookID, input.Type)
}

func (s StatementService) CategoriesGuess(ctx context.Context, input GetCategoriesInput) ([]StatementFrequentCategoryItem, error) {
	statementType := input.Type
	if statementType == "" {
		statementType = "expend"
	}

	filter := repository.CategoryGuessFilter{
		AccountBookID: input.AccountBookID,
		StatementType: statementType,
		Now:           time.Now(),
		Limit:         3,
	}
	rows, err := s.categoryRepo.ListGuessedFrequentByStatementType(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]StatementFrequentCategoryItem, 0, len(rows))
	for _, row := range rows {
		var parent *StatementCategoryParentItem
		if row.HasParent {
			parent = &StatementCategoryParentItem{
				ID:   row.ParentID,
				Name: row.ParentName,
			}
		}
		items = append(items, StatementFrequentCategoryItem{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: s.rowMapper.BuildPublicURL(row.IconPath),
			Parent:   parent,
		})
	}
	return items, nil
}

func (s StatementService) AssetsGuess(ctx context.Context, input GetCategoriesInput) ([]StatementFrequentAssetItem, error) {
	rows, err := s.assetRepo.ListGuessedFrequentByStatementTime(ctx, repository.AssetGuessFilter{
		AccountBookID: input.AccountBookID,
		Now:           time.Now(),
		Limit:         3,
	})
	if err != nil {
		return nil, err
	}

	items := make([]StatementFrequentAssetItem, 0, len(rows))
	for _, row := range rows {
		var parent *StatementAssetParentItem
		if row.HasParent {
			parent = &StatementAssetParentItem{
				ID:   row.ParentID,
				Name: row.ParentName,
			}
		}
		items = append(items, StatementFrequentAssetItem{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: s.rowMapper.BuildPublicURL(row.IconPath),
			Parent:   parent,
		})
	}
	return items, nil
}

func (s StatementService) GetStatementByID(ctx context.Context, statementID int64, accountBookID int64) (statementdto.DetailItem, error) {
	row, err := s.queryRepo.GetRowByIDWithRelations(ctx, statementID, accountBookID)
	if err != nil {
		return statementdto.DetailItem{}, err
	}

	return s.rowMapper.ToDetailItem(row), nil
}

func (s StatementService) GenerateShareKey(ctx context.Context, input StatementGenerateShareKeyInput) (string, error) {
	if input.AccountBookID <= 0 || input.UserID <= 0 {
		return "", ErrStatementInvalidInput
	}

	cacheKey := statementShareCacheKey(input.AccountBookID, input.UserID, input.StartDate, input.EndDate, input.CategoryIDs)
	if s.cache != nil {
		if cached, ok := s.cache.Get(cacheKey); ok && strings.TrimSpace(cached) != "" {
			return cached, nil
		}
	}

	payload := StatementShareTokenPayload{
		AccountBookID:      input.AccountBookID,
		UserID:             input.UserID,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		CategoryIDs:        input.CategoryIDs,
		ExceptStatementIDs: input.ExceptStatementIDs,
	}
	token, err := s.encryptSharePayload(payload)
	if err != nil {
		return "", err
	}

	if s.cache != nil {
		s.cache.Set(cacheKey, token, 365*24*time.Hour)
	}
	return token, nil
}

func (s StatementService) ListByToken(ctx context.Context, input StatementListByTokenInput) (StatementListByTokenResult, error) {
	token := strings.TrimSpace(input.Token)
	if token == "" {
		return StatementListByTokenResult{}, ErrStatementInvalidInput
	}

	payload, err := s.decryptSharePayload(token)
	if err != nil {
		return StatementListByTokenResult{}, ErrStatementDecodeFailed
	}

	cacheKey := statementShareCacheKey(payload.AccountBookID, payload.UserID, payload.StartDate, payload.EndDate, payload.CategoryIDs)
	if s.cache == nil {
		return StatementListByTokenResult{}, ErrStatementInvalidToken
	}
	if _, ok := s.cache.Get(cacheKey); !ok {
		return StatementListByTokenResult{}, ErrStatementInvalidToken
	}

	listInput := statementdto.ListInput{
		AccountBookID:     payload.AccountBookID,
		OrderBy:           input.OrderBy,
		Limit:             1000,
		Offset:            0,
		ParentCategoryIDs: parseCSVInt64OrEmpty(payload.CategoryIDs),
		ExceptIDs:         parseCSVInt64OrEmpty(payload.ExceptStatementIDs),
	}
	if t, ok := parseDateOnly(payload.StartDate); ok {
		listInput.StartDate = &t
	}
	if t, ok := parseDateOnly(payload.EndDate); ok {
		listInput.EndDate = &t
	}

	statements, err := s.GetStatements(ctx, listInput)
	if err != nil {
		return StatementListByTokenResult{}, err
	}

	user, err := s.userRepo.FindByID(ctx, payload.UserID)
	if err != nil {
		return StatementListByTokenResult{}, err
	}

	return StatementListByTokenResult{
		Data: statements,
		DateRange: StatementDateRangeItem{
			StartDate: payload.StartDate,
			EndDate:   payload.EndDate,
		},
		SharedUser: StatementSharedUserItem{
			Nickname:   user.Nickname,
			AvatarPath: s.rowMapper.BuildPublicURL(user.AvatarUrl),
		},
	}, nil
}

func (s StatementService) GetImages(ctx context.Context, accountBookID int64) (StatementImagesResult, error) {
	rows, err := s.queryRepo.ListAvatarRows(ctx, accountBookID)
	if err != nil {
		return StatementImagesResult{}, err
	}

	yearOrder := make([]int, 0)
	monthOrderByYear := make(map[int][]int)
	timeline := make(map[int]map[int][]StatementImageItem)
	seenImage := make(map[string]struct{})
	avatars := make([]string, 0)

	for _, row := range rows {
		year := row.Year
		month := row.Month
		path := s.rowMapper.BuildPublicURL(row.AvatarPath)
		imageItem := StatementImageItem{
			StatementID: row.StatementID,
			AvatarID:    row.AvatarID,
			Path:        path,
		}

		if _, ok := timeline[year]; !ok {
			timeline[year] = make(map[int][]StatementImageItem)
			yearOrder = append(yearOrder, year)
		}
		if _, ok := timeline[year][month]; !ok {
			timeline[year][month] = make([]StatementImageItem, 0)
			monthOrderByYear[year] = append(monthOrderByYear[year], month)
		}
		timeline[year][month] = append(timeline[year][month], imageItem)

		if _, ok := seenImage[path]; !ok && path != "" {
			seenImage[path] = struct{}{}
			avatars = append(avatars, path)
		}
	}

	resultTimeline := make([]StatementImageYearGroup, 0, len(yearOrder))
	for _, year := range yearOrder {
		months := monthOrderByYear[year]
		monthGroups := make([]StatementImageMonthGroup, 0, len(months))
		for _, month := range months {
			monthGroups = append(monthGroups, StatementImageMonthGroup{
				Month: month,
				Data:  dedupeImageItems(timeline[year][month]),
			})
		}
		resultTimeline = append(resultTimeline, StatementImageYearGroup{
			Year: year,
			Data: monthGroups,
		})
	}

	return StatementImagesResult{
		AvatarTimeline: resultTimeline,
		Avatars:        avatars,
	}, nil
}

func (s StatementService) RemoveAvatar(ctx context.Context, accountBookID int64, statementID int64, avatarID int64) error {
	if accountBookID <= 0 || statementID <= 0 || avatarID <= 0 {
		return ErrStatementInvalidInput
	}
	return s.statementRepo.DeleteAvatarByID(ctx, accountBookID, statementID, avatarID)
}

func (s StatementService) ExportCheck(_ context.Context, userID int64) (StatementExportCheckResult, error) {
	if userID <= 0 {
		return StatementExportCheckResult{}, ErrStatementInvalidInput
	}
	count, err := s.getTodayExportCount(userID)
	if err != nil {
		return StatementExportCheckResult{}, err
	}
	if count >= 5 {
		return StatementExportCheckResult{}, ErrStatementExportLimited
	}
	return StatementExportCheckResult{TodayCount: count}, nil
}

func (s StatementService) ExportRows(ctx context.Context, input StatementExportInput) (StatementExportResult, error) {
	if input.AccountBookID <= 0 || input.UserID <= 0 {
		return StatementExportResult{}, ErrStatementInvalidInput
	}
	count, err := s.getTodayExportCount(input.UserID)
	if err != nil {
		return StatementExportResult{}, err
	}
	if count >= 5 {
		return StatementExportResult{}, ErrStatementExportLimited
	}
	if err := s.setTodayExportCount(input.UserID, count+1); err != nil {
		return StatementExportResult{}, err
	}

	end := time.Now()
	start := end.AddDate(0, -1, 0)
	switch strings.TrimSpace(input.Range) {
	case "3months":
		start = end.AddDate(0, -3, 0)
	case "all":
		start = end.AddDate(-100, 0, 0)
	case "1month", "":
	default:
		start = end.AddDate(0, -1, 0)
	}

	rows, err := s.queryRepo.ListExportRows(ctx, repository.StatementExportFilter{
		AccountBookID: input.AccountBookID,
		StartDate:     start,
		EndDate:       end,
		Limit:         3000,
	})
	if err != nil {
		return StatementExportResult{}, err
	}

	items := make([]StatementExportRowItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, StatementExportRowItem{
			Category:       row.CategoryName,
			ParentCategory: row.ParentCategoryName,
			Type:           row.Type,
			TypeName:       statementTypeCNForExport(row.Type),
			Asset:          row.AssetName,
			Description:    row.Description,
			Amount:         row.Amount,
			CreatedAt:      row.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      row.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return StatementExportResult{Rows: items}, nil
}

func (s StatementService) ExportExcelFile(ctx context.Context, input StatementExportInput) ([]byte, error) {
	result, err := s.ExportRows(ctx, input)
	if err != nil {
		return nil, err
	}
	return buildStatementsExcel(result.Rows)
}

func statementShareCacheKey(accountBookID int64, userID int64, startDate string, endDate string, categoryIDs string) string {
	return fmt.Sprintf("share_key_%d_%d_%s_%s_%s", accountBookID, userID, strings.TrimSpace(startDate), strings.TrimSpace(endDate), strings.TrimSpace(categoryIDs))
}

func statementExportCacheKey(userID int64, now time.Time) string {
	return fmt.Sprintf("export_excel_limit_%d_%s", userID, now.Format("20060102"))
}

func (s StatementService) getTodayExportCount(userID int64) (int, error) {
	if s.cache == nil {
		return 0, nil
	}
	key := statementExportCacheKey(userID, time.Now())
	raw, ok := s.cache.Get(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return 0, nil
	}
	count, err := parseCounterInt(strings.TrimSpace(raw))
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func (s StatementService) setTodayExportCount(userID int64, count int) error {
	if s.cache == nil {
		return nil
	}
	if count < 0 {
		count = 0
	}
	key := statementExportCacheKey(userID, time.Now())
	ttl := 24 * time.Hour
	s.cache.Set(key, fmt.Sprintf("%d", count), ttl)
	return nil
}

func (s StatementService) encryptSharePayload(payload StatementShareTokenPayload) (string, error) {
	if strings.TrimSpace(s.tokenSecret) == "" {
		return "", ErrStatementInvalidInput
	}
	plain, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	key := sha256.Sum256([]byte(s.tokenSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	plain = pkcs7PadStatement(plain, aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plain))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plain)

	combined := append(iv, ciphertext...)
	first := base64.StdEncoding.EncodeToString(combined)
	second := base64.StdEncoding.EncodeToString([]byte(first))
	return second, nil
}

func (s StatementService) decryptSharePayload(token string) (StatementShareTokenPayload, error) {
	if strings.TrimSpace(s.tokenSecret) == "" {
		return StatementShareTokenPayload{}, ErrStatementDecodeFailed
	}
	firstBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return StatementShareTokenPayload{}, err
	}
	combined, err := base64.StdEncoding.DecodeString(string(firstBytes))
	if err != nil {
		return StatementShareTokenPayload{}, err
	}
	if len(combined) < aes.BlockSize || len(combined)%aes.BlockSize != 0 {
		return StatementShareTokenPayload{}, ErrStatementDecodeFailed
	}

	key := sha256.Sum256([]byte(s.tokenSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return StatementShareTokenPayload{}, err
	}
	iv := combined[:aes.BlockSize]
	ciphertext := combined[aes.BlockSize:]
	plain := make([]byte, len(ciphertext))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plain, ciphertext)

	plain, err = pkcs7UnpadStatement(plain, aes.BlockSize)
	if err != nil {
		return StatementShareTokenPayload{}, err
	}

	var payload StatementShareTokenPayload
	if err := json.Unmarshal(plain, &payload); err != nil {
		return StatementShareTokenPayload{}, err
	}
	if payload.AccountBookID <= 0 || payload.UserID <= 0 {
		return StatementShareTokenPayload{}, ErrStatementDecodeFailed
	}
	return payload, nil
}

func pkcs7PadStatement(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := make([]byte, padding)
	for i := 0; i < padding; i++ {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

func pkcs7UnpadStatement(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, ErrStatementDecodeFailed
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, ErrStatementDecodeFailed
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, ErrStatementDecodeFailed
		}
	}
	return data[:len(data)-padding], nil
}

func parseCSVInt64OrEmpty(v string) []int64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		num, err := parseInt64(part)
		if err != nil || num <= 0 {
			continue
		}
		out = append(out, num)
	}
	return out
}

func parseDateOnly(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", v, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func parseInt64(v string) (int64, error) {
	var n int64
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return 0, ErrStatementInvalidInput
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}

func parseCounterInt(v string) (int, error) {
	n64, err := parseInt64(v)
	if err != nil {
		return 0, err
	}
	return int(n64), nil
}

func dedupeImageItems(items []StatementImageItem) []StatementImageItem {
	seen := make(map[int64]struct{}, len(items))
	out := make([]StatementImageItem, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item.AvatarID]; ok {
			continue
		}
		seen[item.AvatarID] = struct{}{}
		out = append(out, item)
	}
	return out
}

func statementTypeCNForExport(statementType string) string {
	switch strings.TrimSpace(statementType) {
	case "income":
		return "收入"
	case "expend":
		return "支出"
	case "transfer":
		return "转账"
	case "repayment":
		return "还款"
	case "loan_in":
		return "借入"
	case "loan_out":
		return "借出"
	case "reimburse":
		return "报销"
	case "payment_proxy":
		return "代付"
	default:
		return strings.TrimSpace(statementType)
	}
}

func statementAvatarSecureURL(baseURL string, statementID int64, avatarID int64) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return fmt.Sprintf("/api/statements/%d/avatars/%d", statementID, avatarID)
	}
	return fmt.Sprintf("%s/api/statements/%d/avatars/%d", strings.TrimRight(baseURL, "/"), statementID, avatarID)
}

// Keep import alive for now to preserve compile context with domain in service package.
var _ = domain.User{}
