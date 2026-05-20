package dto

import (
	"encoding/json"
	"reflect"
	"strconv"
)

// FlexibleAmount 支持数字或字符串的金额字段
type FlexibleAmount string

func (a *FlexibleAmount) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*a = FlexibleAmount(s)
		return nil
	}
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*a = FlexibleAmount(strconv.FormatFloat(n, 'f', -1, 64))
		return nil
	}
	return &json.UnmarshalTypeError{Value: string(data), Type: typeString}
}

var typeString = reflect.TypeOf("")

type AssetWriteRequest struct {
	Wallet struct {
		Name     string `json:"name"`
		Amount   string `json:"amount"`
		ParentID int64  `json:"parent_id"`
		IconPath string `json:"icon_path"`
		Remark   string `json:"remark"`
		Type     string `json:"type"`
	} `json:"wallet"`
}

type AssetSurplusRequest struct {
	AssetID int64          `json:"asset_id"`
	Amount  FlexibleAmount `json:"amount"`
}
