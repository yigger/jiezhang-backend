package helper

import "time"

type StatementData struct {
	CategoryParentID   int64
	AssetParentID      int64
	ID                 int64
	UserID             int64
	Type               string
	Amount             float64
	Description        string
	CategoryID         int64
	AssetID            int64
	Remark             string
	Mood               string
	IconPath           string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CategoryName       string
	AssetName          string
	Location           string
	Nation             string
	Province           string
	City               string
	District           string
	Street             string
	HasPic             bool
	Residue            float64
	PayeeID            int64
	PayeeName          string
	TargetAssetID      int64
	TargetAssetName    string
	TargetObject       string
	CategoryParentName string
	AssetParentName    string
}
