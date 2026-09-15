package types

// APIResponse describes the legacy response envelope. Some endpoints return
// resources directly. Authentication errors use HTTP 200 with status/msg.
type APIResponse struct {
	Message string      `json:"message,omitempty"`
	Status  int         `json:"status,omitempty"`
	Msg     string      `json:"msg,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Session string      `json:"session,omitempty"`
}
