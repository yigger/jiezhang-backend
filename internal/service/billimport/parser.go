package billimport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode"
)

func invalid(message string) error { return fmt.Errorf("%w: %s", ErrInvalid, message) }

func (s *Service) parse(ctx context.Context, file File, index int) ([]entry, error) {
	if file.Reader == nil {
		return nil, invalid("missing file reader")
	}
	if file.Source != "" && file.Source != "alipay" && file.Source != "wechat" {
		return nil, invalid("source must be alipay or wechat")
	}
	tables, err := s.reader.Read(ctx, file.Reader, file.Format)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, invalid("cannot read bill file; expected unencrypted CSV/XLSX")
	}
	var selected *Table
	var headers map[string]int
	var start int
	var source string
	for i := range tables {
		for row, cells := range tables[i].Rows {
			if row >= 100 {
				break
			}
			h := make(map[string]int)
			duplicate := false
			for col, c := range cells {
				name := clean(c.Text)
				if name != "" {
					if _, exists := h[name]; exists {
						duplicate = true
					}
					h[name] = col
				}
			}
			candidate := ""
			if _, ok := h["交易订单号"]; ok {
				candidate = "alipay"
			}
			if _, ok := h["交易单号"]; ok {
				candidate = "wechat"
			}
			if _, ok := h["交易时间"]; !ok || candidate == "" {
				continue
			}
			if duplicate {
				return nil, invalid("duplicate header")
			}
			if selected != nil {
				return nil, invalid("multiple bill tables; supply one table per file")
			}
			selected, headers, start, source = &tables[i], h, row+1, candidate
		}
	}
	if selected == nil {
		return nil, invalid("unrecognized bill headers")
	}
	if file.Source != "" && file.Source != source {
		return nil, invalid("source does not match bill headers")
	}
	columns := []string{"交易时间", "交易对方", "收/支", "备注"}
	if source == "alipay" {
		columns = append(columns, "交易分类", "商品说明", "金额", "收/付款方式", "交易状态", "交易订单号")
	} else {
		columns = append(columns, "交易类型", "商品", "金额(元)", "支付方式", "当前状态", "交易单号")
	}
	for _, name := range columns {
		if _, ok := headers[name]; !ok {
			return nil, invalid("missing column: " + name)
		}
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, invalid("cannot load bill timezone")
	}
	entries := make([]entry, 0)
	for row := start; row < len(selected.Rows); row++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		cells := selected.Rows[row]
		get := func(name string) Cell {
			i, ok := headers[name]
			if !ok || i >= len(cells) {
				return Cell{}
			}
			return cells[i]
		}
		text := func(name string) string { return clean(get(name).Text) }
		blank := true
		for _, c := range cells {
			if clean(c.Text) != "" {
				blank = false
				break
			}
		}
		if blank {
			continue
		}
		date := text("交易时间")
		if strings.HasPrefix(date, "---") || strings.HasPrefix(date, "说明") || strings.HasPrefix(date, "温馨提示") || strings.HasPrefix(date, "共") {
			break
		}
		e := entry{File: index, Row: row + 1, Source: source, Counterparty: text("交易对方"), Note: text("备注")}
		amount, id := "金额", "交易订单号"
		if source == "alipay" {
			e.Category, e.Product, e.Payment, e.Status = text("交易分类"), text("商品说明"), text("收/付款方式"), text("交易状态")
		} else {
			amount, id = "金额(元)", "交易单号"
			e.Category, e.Product, e.Payment, e.Status = text("交易类型"), text("商品"), text("支付方式"), text("当前状态")
		}
		e.ID = text(id)
		switch text("收/支") {
		case "收入":
			e.Kind = "income"
		case "支出":
			e.Kind = "expend"
		default:
			e.Skip = "non_income_expense"
		}
		if special(e) {
			e.Skip = "special_transaction"
		}
		if e.Skip == "" && !successful(e) {
			e.Skip = "unsupported_or_incomplete_status"
		}
		if strings.ContainsAny(e.Payment, "&+") || strings.Contains(e.Payment, "组合") {
			e.Skip = "mixed_payment"
		}
		e.Occurred, err = parseDate(get("交易时间"), selected.Excel1904, location)
		if err != nil {
			e.Invalid = "invalid_datetime"
		}
		e.Cents, err = parseCents(text(amount))
		if err != nil {
			e.Invalid = "invalid_amount"
		} else if e.Cents == 0 && e.Skip == "" {
			e.Skip = "zero_amount"
		}
		if e.ID == "" || e.ID == "/" || e.ID == "-" || len([]rune(e.ID)) > 255 {
			e.Invalid = "invalid_order_id"
		}
		if get(id).Numeric || get(id).Formula {
			e.Invalid = "order_id_must_be_text"
		}
		for _, name := range columns {
			if get(name).Formula {
				e.Invalid = "formula_not_allowed"
			}
		}
		if len(description(e)) > 65535 {
			e.Invalid = "description_too_long"
		}
		// Fingerprint original semantics, independent of category/asset mapping and file layout.
		payload, _ := json.Marshal([]any{e.Source, e.ID, e.Kind, e.Category, e.Counterparty, e.Product, e.Payment, e.Status, e.Note, e.Occurred.Format(time.RFC3339), e.Cents})
		sum := sha256.Sum256(payload)
		e.Fingerprint = hex.EncodeToString(sum[:])
		entries = append(entries, e)
		if len(entries) > MaxRows {
			return nil, invalid("too many bill rows")
		}
	}
	return entries, nil
}

func clean(s string) string { return strings.TrimSpace(strings.TrimPrefix(s, "\ufeff")) }

// Source order identity stays stable when a user edits the imported statement.
func importKey(source, orderID string) string {
	sum := sha256.Sum256([]byte(source + ":" + orderID))
	return hex.EncodeToString(sum[:])
}

func special(e entry) bool {
	for _, token := range []string{"退款", "退还", "转账", "红包", "群收款", "AA收款", "提现", "零钱充值", "余额充值", "账户充值", "还款", "借", "贷款", "放款", "理财", "基金", "账户存取", "余额宝", "零钱通"} {
		if strings.Contains(e.Category, token) || strings.Contains(e.Status, token) {
			return true
		}
	}
	return false
}
func successful(e entry) bool {
	if e.Source == "alipay" {
		return e.Status == "交易成功" || e.Status == "支付成功"
	}
	return e.Status == "支付成功" || e.Status == "已到账" || e.Status == "已存入零钱"
}

func parseCents(value string) (int64, error) {
	v := strings.TrimSpace(strings.TrimLeft(value, "¥￥"))
	parts := strings.Split(v, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, ErrInvalid
	}
	for _, p := range parts {
		for _, c := range p {
			if c < '0' || c > '9' {
				return 0, ErrInvalid
			}
		}
	}
	if len(parts) == 2 && (len(parts[1]) == 0 || len(parts[1]) > 2) {
		return 0, ErrInvalid
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole > 9999999999 {
		return 0, ErrInvalid
	}
	var fraction int64
	if len(parts) == 2 {
		fraction, _ = strconv.ParseInt(parts[1]+strings.Repeat("0", 2-len(parts[1])), 10, 64)
	}
	cents := whole*100 + fraction
	return cents, nil
}

func parseDate(cell Cell, use1904 bool, location *time.Location) (time.Time, error) {
	v := clean(cell.Text)
	if cell.Numeric {
		serial, err := strconv.ParseFloat(v, 64)
		if err != nil || math.IsNaN(serial) || math.IsInf(serial, 0) || serial < 61 || serial > 2958465 {
			return time.Time{}, ErrInvalid
		}
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, location)
		if use1904 {
			base = time.Date(1904, 1, 1, 0, 0, 0, 0, location)
		}
		days := int(math.Floor(serial))
		seconds := int64(math.Round((serial - float64(days)) * 86400))
		return base.AddDate(0, 0, days).Add(time.Duration(seconds) * time.Second), nil
	}
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006/1/2 15:04:05", "2006/1/2 15:04"} {
		if t, err := time.ParseInLocation(layout, v, location); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrInvalid
}

func normalize(name string) string {
	return strings.Map(func(c rune) rune {
		if unicode.IsSpace(c) {
			return -1
		}
		switch c {
		case '（':
			return '('
		case '）':
			return ')'
		}
		return unicode.ToLower(c)
	}, clean(name))
}
