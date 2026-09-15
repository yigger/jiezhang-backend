package types

type CalendarDataItem struct {
	Date   int     `json:"date"`
	Income float64 `json:"income"`
	Expend float64 `json:"expend"`
}

type OverviewHeaderData struct {
	TotalBalance float64 `json:"total"`
	Repay        float64 `json:"repay"`
	Transfer     float64 `json:"transfer"`
	Income       float64 `json:"income"`
	Expend       float64 `json:"expend"`
}
