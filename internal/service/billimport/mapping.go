package billimport

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

// These are explicit equivalences, not fuzzy substrings. Breakfast and dinner
// remain distinct; a bank card's account suffix is never discarded.
var equivalences = [][]string{
	{"早餐", "早饭", "早点", "早餐费"},
	{"午餐", "午饭", "中餐", "午餐费"},
	{"晚餐", "晚饭", "晚餐费"},
	{"伙食餐饮", "餐饮美食", "餐饮", "饮食"},
	{"日常饮食", "日常餐饮", "日常用餐"},
	{"奶茶咖啡", "咖啡奶茶", "茶饮咖啡"},
	{"交通出行", "交通", "出行"},
	{"公共交通", "公交地铁", "公交", "地铁"},
	{"打车租车", "打车", "出租车", "网约车"},
	{"其他杂项", "其他", "其它"},
	{"其他支出", "其它支出"},
	{"其他收入", "其它收入"},
}

var unfinishedCardSuffix = regexp.MustCompile(`\([0-9]{4}$`)

func categoryName(name string) string {
	n := normalize(name)
	for _, group := range equivalences {
		for _, alias := range group {
			if n == normalize(alias) {
				return normalize(group[0])
			}
		}
	}
	return n
}

func assetName(name string) string {
	n := normalize(name)
	for _, group := range [][]string{{"支付宝", "支付宝余额", "余额"}, {"微信", "微信零钱", "零钱"}, {"储蓄账户", "储蓄", "银行卡账户"}, {"虚拟账户", "网络账户"}, {"负债账户", "负债"}} {
		for _, alias := range group {
			if n == normalize(alias) {
				return normalize(group[0])
			}
		}
	}
	for _, pair := range [][2]string{{"中国工商银行", "工商银行"}, {"工行", "工商银行"}, {"中国建设银行", "建设银行"}, {"建行", "建设银行"}, {"中国农业银行", "农业银行"}, {"农行", "农业银行"}, {"中国银行", "中行"}, {"招商银行", "招行"}} {
		n = strings.ReplaceAll(n, normalize(pair[0]), normalize(pair[1]))
	}
	if strings.Contains(n, "银行") || strings.Contains(n, "招行") || strings.Contains(n, "中行") {
		for _, label := range []string{"储蓄卡", "借记卡", "信用卡", "银行卡"} {
			n = strings.ReplaceAll(n, label, "")
		}
		n = strings.ReplaceAll(n, "尾号", "(")
		if unfinishedCardSuffix.MatchString(n) {
			n += ")"
		}
	}
	return n
}

func strictTarget(target Target) Target { target.strictParent = target.Parent != ""; return target }

func categoryTarget(e entry, mapping Mapping) Target {
	if target, ok := mapping.Categories[e.Kind+":"+e.Category]; ok {
		return strictTarget(target)
	}
	text := e.Product + " " + e.Counterparty
	for _, rule := range mapping.Rules {
		if rule.Type != e.Kind {
			continue
		}
		for _, keyword := range rule.Contains {
			if clean(keyword) != "" && strings.Contains(text, keyword) {
				return strictTarget(rule.Target)
			}
		}
	}
	if e.Kind == "income" {
		return Target{Name: "其他收入", Parent: "其他杂项"}
	}
	for _, rule := range []struct {
		words        []string
		name, parent string
	}{
		{[]string{"早餐", "早饭", "早点"}, "早餐", "伙食餐饮"},
		{[]string{"午餐", "午饭"}, "午餐", "伙食餐饮"},
		{[]string{"晚餐", "晚饭"}, "晚餐", "伙食餐饮"},
		{[]string{"奶茶", "咖啡"}, "奶茶咖啡", "伙食餐饮"},
		{[]string{"地铁", "公交"}, "公共交通", "交通出行"},
		{[]string{"出租车", "网约车", "打车"}, "打车租车", "交通出行"},
	} {
		if e.Source == "alipay" && categoryName(e.Category) != categoryName(rule.parent) {
			continue
		}
		for _, word := range rule.words {
			if strings.Contains(text, word) {
				return Target{Name: rule.name, Parent: rule.parent}
			}
		}
	}
	if categoryName(e.Category) == categoryName("伙食餐饮") {
		return Target{Name: "日常饮食", Parent: "伙食餐饮"}
	}
	if e.Source == "alipay" && e.Category != "" && e.Category != "其他" {
		return Target{Name: e.Category, Parent: "导入支出"}
	}
	return Target{Name: "其他支出", Parent: "其他杂项"}
}

func assetTarget(e entry, mapping Mapping) (Target, string) {
	if target, ok := mapping.Assets[e.Source+":"+e.Payment]; ok {
		return strictTarget(target), ""
	}
	name, parent, kind := e.Payment, "虚拟账户", "deposit"
	if name == "" || name == "-" || name == "/" || name == "--" {
		if e.Source == "wechat" {
			name = "微信未注明账户"
		} else {
			name = "支付宝未注明账户"
		}
	} else if strings.Contains(name, "信用卡") || strings.Contains(name, "花呗") || strings.Contains(name, "白条") {
		kind, parent = "debt", "负债账户"
	} else if strings.Contains(name, "银行") || strings.Contains(name, "储蓄卡") || strings.Contains(name, "借记卡") {
		parent = "储蓄账户"
	} else if name == "余额" && e.Source == "alipay" {
		name = "支付宝"
	} else if name == "零钱" && e.Source == "wechat" {
		name = "微信"
	}
	return Target{Name: name, Parent: parent}, kind
}

type resolver struct {
	catalog    repo.ImportCatalog
	user, book int64
	admin      bool
	next       int64
	categories []model.Category
	assets     []model.Asset
}

func newResolver(c repo.ImportCatalog, user, book int64) (*resolver, error) {
	r := &resolver{catalog: c, user: user, book: book, next: -1}
	allowed := c.Book.ID == book && c.Book.UserID == user
	r.admin = allowed
	for _, m := range c.Members {
		if m.AccountBookID == book && m.UserID == user {
			allowed = c.Book.ID == book
			r.admin = r.admin || m.Role == "owner" || m.Role == "admin"
		}
	}
	if !allowed {
		return nil, invalid("account book access denied")
	}
	return r, nil
}

func validTarget(t Target) bool {
	return t.ID >= 0 && (t.ID > 0 || (clean(t.Name) != "" && len([]rune(t.Name)) <= 255 && len([]rune(t.Parent)) <= 255))
}

func (r *resolver) category(target Target, kind string) (int64, error) {
	if !validTarget(target) {
		return 0, invalid("invalid category mapping")
	}
	if target.ID > 0 {
		for _, c := range r.catalog.Categories {
			if c.ID == target.ID && c.AccountBookID == r.book && c.Type == kind && c.SpecialType == "" && c.ParentID > 0 {
				return c.ID, nil
			}
		}
		return 0, invalid("category ID must be an ordinary leaf in this book with matching type")
	}
	// Exact/alias leaf matches take priority. A supplied parent disambiguates.
	var matches []model.Category
	var exact []model.Category
	for _, c := range r.catalog.Categories {
		if c.AccountBookID != r.book || c.Type != kind || c.SpecialType != "" || c.ParentID == 0 || categoryName(c.Name) != categoryName(target.Name) {
			continue
		}
		if target.strictParent && !r.categoryParentMatches(c.ParentID, target.Parent, kind) {
			continue
		}
		matches = append(matches, c)
		if normalize(c.Name) == normalize(target.Name) {
			exact = append(exact, c)
		}
	}
	if len(exact) > 0 {
		matches = exact
	}
	if len(matches) == 1 {
		return matches[0].ID, nil
	}
	if len(matches) > 1 {
		return 0, invalid("ambiguous category; supply an ID or parent")
	}
	// When the source category names an existing root, reuse that hierarchy.
	for _, c := range r.catalog.Categories {
		if c.AccountBookID == r.book && c.Type == kind && c.ParentID == 0 && c.SpecialType == "" && categoryName(c.Name) == categoryName(target.Name) && categoryName(target.Parent) != categoryName(c.Name) {
			return r.category(strictTarget(Target{Name: "其他", Parent: c.Name}), kind)
		}
	}
	parentName := target.Parent
	if parentName == "" {
		parentName = "导入支出"
		if kind == "income" {
			parentName = "导入收入"
		}
	}
	parent, err := r.categoryParent(parentName, kind)
	if err != nil {
		return 0, err
	}
	if !r.admin {
		return 0, invalid("creating categories requires account book owner/admin")
	}
	row := model.Category{ID: r.next, UserID: r.user, AccountBookID: r.book, Name: clean(target.Name), ParentID: parent, Type: kind}
	r.next--
	r.catalog.Categories = append(r.catalog.Categories, row)
	r.categories = append(r.categories, row)
	return row.ID, nil
}

func (r *resolver) categoryParentMatches(id int64, name, kind string) bool {
	for _, c := range r.catalog.Categories {
		if c.ID == id && c.Type == kind && categoryName(c.Name) == categoryName(name) {
			return true
		}
	}
	return false
}
func (r *resolver) categoryParent(name, kind string) (int64, error) {
	var ids []int64
	for _, c := range r.catalog.Categories {
		if c.AccountBookID == r.book && c.Type == kind && c.ParentID == 0 && c.SpecialType == "" && categoryName(c.Name) == categoryName(name) {
			ids = append(ids, c.ID)
		}
	}
	if len(ids) == 1 {
		return ids[0], nil
	}
	if len(ids) > 1 {
		return 0, invalid("ambiguous parent category")
	}
	if !r.admin {
		return 0, invalid("creating categories requires account book owner/admin")
	}
	row := model.Category{ID: r.next, UserID: r.user, AccountBookID: r.book, Name: clean(name), Type: kind}
	r.next--
	r.catalog.Categories = append(r.catalog.Categories, row)
	r.categories = append(r.categories, row)
	return row.ID, nil
}

func (r *resolver) asset(target Target, kind string) (int64, error) {
	if !validTarget(target) {
		return 0, invalid("invalid asset mapping")
	}
	if target.ID > 0 {
		for _, a := range r.catalog.Assets {
			if a.ID == target.ID && a.AccountBookID == r.book && a.ParentID > 0 {
				return a.ID, nil
			}
		}
		return 0, invalid("asset ID must be a leaf in this book")
	}
	matchKind := kind != ""
	if kind == "" {
		kind = "deposit"
		if strings.Contains(target.Name, "信用卡") || strings.Contains(target.Name, "花呗") || strings.Contains(target.Name, "白条") || assetName(target.Parent) == assetName("负债账户") {
			kind = "debt"
		}
	}
	var ids []int64
	var exact []int64
	for _, a := range r.catalog.Assets {
		if a.AccountBookID != r.book || a.ParentID == 0 || (matchKind && r.assetKind(a) != kind) || assetName(a.Name) != assetName(target.Name) {
			continue
		}
		if target.strictParent {
			matched := false
			for _, p := range r.catalog.Assets {
				if p.ID == a.ParentID && assetName(p.Name) == assetName(target.Parent) {
					matched = true
				}
			}
			if !matched {
				continue
			}
		}
		ids = append(ids, a.ID)
		if normalize(a.Name) == normalize(target.Name) {
			exact = append(exact, a.ID)
		}
	}
	if len(exact) > 0 {
		ids = exact
	}
	if len(ids) == 1 {
		return ids[0], nil
	}
	if len(ids) > 1 {
		return 0, invalid("ambiguous asset; supply an ID or parent")
	}
	if !r.admin {
		return 0, invalid("creating assets requires account book owner/admin")
	}
	parentName := target.Parent
	if parentName == "" {
		parentName = "导入资产"
	}
	var parents []int64
	for _, a := range r.catalog.Assets {
		if a.AccountBookID == r.book && a.ParentID == 0 && r.assetKind(a) == kind && assetName(a.Name) == assetName(parentName) {
			parents = append(parents, a.ID)
		}
	}
	if len(parents) > 1 {
		return 0, invalid("ambiguous parent asset")
	}
	var parent int64
	if len(parents) == 1 {
		parent = parents[0]
	} else {
		p := model.Asset{ID: r.next, CreatorID: r.user, AccountBookID: r.book, Name: parentName, Type: kind}
		r.next--
		r.catalog.Assets = append(r.catalog.Assets, p)
		r.assets = append(r.assets, p)
		parent = p.ID
	}
	a := model.Asset{ID: r.next, CreatorID: r.user, AccountBookID: r.book, Name: clean(target.Name), ParentID: parent, Type: kind}
	r.next--
	r.catalog.Assets = append(r.catalog.Assets, a)
	r.assets = append(r.assets, a)
	return a.ID, nil
}

func mappingError(err error) string {
	return strings.TrimPrefix(fmt.Sprint(err), ErrInvalid.Error()+": ")
}

func (r *resolver) assetKind(a model.Asset) string {
	if a.Type != "" {
		return a.Type
	}
	for _, parent := range r.catalog.Assets {
		if parent.ID == a.ParentID && parent.Type != "" {
			return parent.Type
		}
	}
	return "deposit"
}
