package urlbuilder

import (
	"net/url"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/infrastructure/signedurl"
)

type PublicURLBuilder struct {
	baseURL string
	signer  *signedurl.Signer
}

func NewPublicURLBuilder(baseURL string) PublicURLBuilder {
	return PublicURLBuilder{baseURL: strings.TrimSpace(baseURL)}
}

func NewPublicURLBuilderWithSigner(baseURL string, signer *signedurl.Signer) PublicURLBuilder {
	return PublicURLBuilder{baseURL: strings.TrimSpace(baseURL), signer: signer}
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

	result := base.String()

	if b.signer != nil && signedurl.IsPrivatePath(path) {
		result = b.signer.SignIfPrivate(result)
	}

	return result
}
