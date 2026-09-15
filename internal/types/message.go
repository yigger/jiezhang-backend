package types

type MessageDetailItem struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Time        string `json:"time"`
	ContentType string `json:"content_type"`
	MsgType     string `json:"msg_type"`
}

type MessageListItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	TargetType  int    `json:"target_type"`
	ContentType string `json:"content_type"`
	AlreadyRead int    `json:"already_read"`
	PageURL     string `json:"page_url"`
	MsgType     string `json:"msg_type"`
	SubTitle    string `json:"sub_title"`
	Time        string `json:"time"`
	ImageURL    string `json:"image_url"`
}
