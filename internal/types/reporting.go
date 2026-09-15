package types

type SuperWeekData struct {
	Weeks []string  `json:"weeks"`
	Data  []float64 `json:"data"`
}

type SuperLineChartData struct {
	Months  []int     `json:"months"`
	Expends []float64 `json:"expends"`
	Incomes []float64 `json:"incomes"`
	Surplus []float64 `json:"surplus"`
}

type SuperCategoryTopItem struct {
	Name         string  `json:"name"`
	Data         float64 `json:"data"`
	FormatAmount string  `json:"format_amount"`
	Percent      string  `json:"percent"`
	CategoryID   int64   `json:"category_id"`
}

type SuperPieItem struct {
	Name string `json:"name"`
	Data int64  `json:"data"`
}

type SuperTableSummaryItem struct {
	Date         string `json:"date"`
	Expend       string `json:"expend"`
	Income       string `json:"income"`
	TotalIncome  string `json:"total_income"`
	TotalExpend  string `json:"total_expend"`
	TotalSurplus string `json:"total_surplus"`
}

type SuperChartHeaderData struct {
	ExpendCount    float64 `json:"expend_count"`
	IncomeCount    float64 `json:"income_count"`
	Surplus        float64 `json:"surplus"`
	ExpendPercent  float64 `json:"expend_percent"`
	ExpendRise     string  `json:"expend_rise"`
	IncomePercent  float64 `json:"income_percent"`
	IncomeRise     string  `json:"income_rise"`
	SurplusPercent float64 `json:"surplus_percent"`
	SurplusRise    string  `json:"surplus_rise"`
}

type SuperTimeResponse struct {
	Statements []SuperMonthItem `json:"statements"`
	Header     SuperHeader      `json:"header"`
}

type SuperHeader struct {
	Expend string `json:"expend"`
	Income string `json:"income"`
	Left   string `json:"left"`
}

type SuperMonthItem struct {
	ExpendAmount float64 `json:"expend_amount"`
	IncomeAmount float64 `json:"income_amount"`
	Surplus      float64 `json:"surplus"`
	Year         int     `json:"year"`
	Month        int     `json:"month"`
	Hidden       int     `json:"hidden"`
}
