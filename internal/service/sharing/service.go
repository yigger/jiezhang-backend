package sharing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	statement "github.com/yigger/jiezhang-backend/internal/service/statement"
	"github.com/yigger/jiezhang-backend/internal/types"
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

func (s Service) GenerateShareKey(ctx context.Context, input StatementGenerateShareKeyInput) (string, error) {
	if input.AccountBookID <= 0 || input.UserID <= 0 {
		return "", statement.ErrStatementInvalidInput
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

func (s Service) ListByToken(ctx context.Context, input StatementListByTokenInput) (types.StatementListByTokenResult, error) {
	token := strings.TrimSpace(input.Token)
	if token == "" {
		return types.StatementListByTokenResult{}, statement.ErrStatementInvalidInput
	}

	payload, err := s.decryptSharePayload(token)
	if err != nil {
		return types.StatementListByTokenResult{}, ErrStatementDecodeFailed
	}

	cacheKey := statementShareCacheKey(payload.AccountBookID, payload.UserID, payload.StartDate, payload.EndDate, payload.CategoryIDs)
	if s.cache == nil {
		return types.StatementListByTokenResult{}, ErrStatementInvalidToken
	}
	if _, ok := s.cache.Get(cacheKey); !ok {
		return types.StatementListByTokenResult{}, ErrStatementInvalidToken
	}

	listInput := statement.ListInput{
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

	statements, err := s.reader.GetStatements(ctx, listInput)
	if err != nil {
		return types.StatementListByTokenResult{}, err
	}

	userRow, err := s.userRepo.FindByID(ctx, payload.UserID)
	if err != nil {
		return types.StatementListByTokenResult{}, err
	}
	user := helperservice.ToUserContext(userRow)

	return types.StatementListByTokenResult{
		Data: statements,
		DateRange: types.StatementDateRangeItem{
			StartDate: payload.StartDate,
			EndDate:   payload.EndDate,
		},
		SharedUser: types.StatementSharedUserItem{
			Nickname:   user.Nickname,
			AvatarPath: s.rowMapper.BuildPublicURL(user.AvatarUrl),
		},
	}, nil
}

func statementShareCacheKey(accountBookID int64, userID int64, startDate string, endDate string, categoryIDs string) string {
	return fmt.Sprintf("share_key_%d_%d_%s_%s_%s", accountBookID, userID, strings.TrimSpace(startDate), strings.TrimSpace(endDate), strings.TrimSpace(categoryIDs))
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
			return 0, statement.ErrStatementInvalidInput
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}

func (s Service) encryptSharePayload(payload StatementShareTokenPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return s.codec.Encrypt(raw)
}
func (s Service) decryptSharePayload(token string) (StatementShareTokenPayload, error) {
	raw, err := s.codec.Decrypt(token)
	if err != nil {
		return StatementShareTokenPayload{}, ErrStatementDecodeFailed
	}
	var payload StatementShareTokenPayload
	if err = json.Unmarshal(raw, &payload); err != nil || payload.AccountBookID <= 0 || payload.UserID <= 0 {
		return StatementShareTokenPayload{}, ErrStatementDecodeFailed
	}
	return payload, nil
}
