package types

type UploadResult struct {
	Status     int    `json:"status"`
	AvatarPath string `json:"avatar_path,omitempty"`
}
