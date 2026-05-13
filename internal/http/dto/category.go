package dto

type CategoryWriteRequest struct {
	Category struct {
		Name     string `json:"name"`
		ParentID int64  `json:"parent_id"`
		IconPath string `json:"icon_path"`
		Type     string `json:"type"`
	} `json:"category"`
}
