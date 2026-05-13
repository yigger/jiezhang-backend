package urlbuilder

import (
	"net/url"
	"strings"
)

type PublicURLBuilder struct {
	baseURL string
}

func NewPublicURLBuilder(baseURL string) PublicURLBuilder {
	return PublicURLBuilder{baseURL: strings.TrimSpace(baseURL)}
}

func (b PublicURLBuilder) BuildPublicURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	if b.baseURL == "" {
		return raw
	}

	base, err := url.Parse(b.baseURL)
	if err != nil {
		return raw
	}
	path := raw
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawQuery = ""
	base.Fragment = ""
	return base.String()
}
