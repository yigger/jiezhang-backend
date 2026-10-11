package statement

import (
	"time"
)

type ListInput struct {
	UserID            int64
	AccountBookID     int64
	StartDate         *time.Time
	EndDate           *time.Time
	ParentCategoryIDs []int64
	ExceptIDs         []int64
	OrderBy           string
	Limit             int
	Offset            int
}

type WriteInput struct {
	ProjectID     int64
	ConsumerID    int64
	StatementID   int64
	UserID        int64
	AccountBookID int64

	Type         string
	Amount       float64
	Description  string
	Mood         string
	CategoryID   int64
	AssetID      int64
	FromAssetID  int64
	ToAssetID    int64
	PayeeID      int64
	TargetObject string

	Location string
	Nation   string
	Province string
	City     string
	District string
	Street   string

	Date string
	Time string
}

type PatchInput struct {
	Type         *string
	Amount       *float64
	Description  *string
	Mood         *string
	CategoryID   *int64
	AssetID      *int64
	FromAssetID  *int64
	ToAssetID    *int64
	PayeeID      *int64
	TargetObject *string
	Location     *string
	Nation       *string
	Province     *string
	City         *string
	District     *string
	Street       *string
	Date         *string
	Time         *string
}

type UpdateInput struct {
	StatementID   int64
	UserID        int64
	AccountBookID int64
	Patch         PatchInput
}
