package dto

type FeedbackRequest struct {
	Content string `json:"content"`
	Type    int    `json:"type"`
}
