package wechat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const jsCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

type SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type Client interface {
	Code2Session(ctx context.Context, code string) (SessionResponse, error)
}

type HTTPClient struct {
	appID      string
	appSecret  string
	httpClient *http.Client
}

func NewHTTPClient(appID, appSecret string) *HTTPClient {
	return &HTTPClient{
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (c *HTTPClient) Code2Session(ctx context.Context, code string) (SessionResponse, error) {
	if strings.TrimSpace(code) == "" {
		return SessionResponse{}, errors.New("empty wx code")
	}

	params := url.Values{}
	params.Set("appid", c.appID)
	params.Set("secret", c.appSecret)
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jsCode2SessionURL+"?"+params.Encode(), nil)
	if err != nil {
		return SessionResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return SessionResponse{}, err
	}
	defer resp.Body.Close()

	var data SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return SessionResponse{}, err
	}

	if data.ErrCode != 0 {
		return SessionResponse{}, fmt.Errorf("wechat errcode=%d errmsg=%s", data.ErrCode, data.ErrMsg)
	}

	if strings.TrimSpace(data.OpenID) == "" || strings.TrimSpace(data.SessionKey) == "" {
		return SessionResponse{}, errors.New("invalid wechat session response")
	}

	return data, nil
}
