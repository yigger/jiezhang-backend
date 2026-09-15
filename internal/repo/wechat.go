package repo

import "context"

type SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type Client interface {
	Code2Session(ctx context.Context, code string) (SessionResponse, error)
}
