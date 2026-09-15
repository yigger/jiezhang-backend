package types

import (
	"time"
)

type HomeSettingsAccountBook struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type HomeSettingsUser struct {
	UID         int64                   `json:"uid"`
	Name        string                  `json:"name"`
	AvatarUrl   string                  `json:"avatar_url"`
	Themes      []Theme                 `json:"themes"`
	ThemeID     int64                   `json:"theme_id"`
	Theme       Theme                   `json:"theme"`
	Persist     int64                   `json:"persist"`
	ShowDiamond bool                    `json:"show_diamond"`
	Remind      bool                    `json:"remind"`
	CreatedAt   time.Time               `json:"created_at"`
	AccountBook HomeSettingsAccountBook `json:"account_book"`
}

type HomeSettingsResponse struct {
	User    HomeSettingsUser `json:"user"`
	Version string           `json:"version"`
}

type HomeHeaderMessage struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
}

type HomeHeaderTrendItem struct {
	Ratio  float64 `json:"ratio"`
	Trend  string  `json:"trend"`
	Amount string  `json:"amount"`
}

type HomeHeaderTrends struct {
	Day   HomeHeaderTrendItem `json:"day"`
	Week  HomeHeaderTrendItem `json:"week"`
	Month HomeHeaderTrendItem `json:"month"`
}

type HomeHeaderResponse struct {
	Trends        HomeHeaderTrends   `json:"trends"`
	MonthExpend   string             `json:"month_expend"`
	TodayExpend   string             `json:"today_expend"`
	MonthBudget   string             `json:"month_budget"`
	UsePencentage int                `json:"use_pencentage"`
	Message       *HomeHeaderMessage `json:"message"`
}
