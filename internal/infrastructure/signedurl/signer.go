package signedurl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultTTL = 1 * time.Hour

type Signer struct {
	secret []byte
}

func NewSigner(secret string) *Signer {
	return &Signer{secret: []byte(secret)}
}

// Sign appends ?exp=<timestamp>&sign=<hmac> to path.
func (s *Signer) Sign(path string, ttl time.Duration) string {
	if path == "" {
		return ""
	}
	expiresAt := time.Now().Add(ttl).Unix()
	exp := strconv.FormatInt(expiresAt, 10)
	sign := s.compute(path, exp)
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + "exp=" + exp + "&sign=" + sign
}

// Verify checks the exp and sign query parameters in rawURL.
func (s *Signer) Verify(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	q := u.Query()
	exp := q.Get("exp")
	sign := q.Get("sign")
	if exp == "" || sign == "" {
		return false
	}
	expInt, err := strconv.ParseInt(exp, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix() > expInt {
		return false
	}
	pathWithoutQuery := (&url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		RawPath:  u.RawPath,
		Path:     u.Path,
		RawQuery: "",
	}).String()
	if strings.HasPrefix(rawURL, "/") {
		pathWithoutQuery = u.Path
	}
	expected := s.compute(pathWithoutQuery, exp)
	return hmac.Equal([]byte(sign), []byte(expected))
}

func (s *Signer) compute(path string, exp string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(path + "|" + exp))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignIfPrivate signs path only if it is a protected statement avatar path.
// The signature is computed on the bare path (without scheme/host) so it
// matches what the middleware sees in c.Request.URL.
func (s *Signer) SignIfPrivate(fullURL string) string {
	if !isProtectedPath(fullURL) {
		return fullURL
	}
	signPath := fullURL
	if u, err := url.Parse(fullURL); err == nil && u.Path != "" {
		signPath = u.Path
	}
	expiresAt := time.Now().Add(defaultTTL).Unix()
	exp := strconv.FormatInt(expiresAt, 10)
	sign := s.compute(signPath, exp)
	sep := "?"
	if strings.Contains(fullURL, "?") {
		sep = "&"
	}
	return fullURL + sep + "exp=" + exp + "&sign=" + sign
}

// IsPrivatePath checks whether a path should require a valid signature.
// Only /private/{id}/statements/ paths are protected.
func IsPrivatePath(p string) bool {
	return isProtectedPath(p)
}

func isProtectedPath(p string) bool {
	return strings.Contains(p, "/private/") && strings.Contains(p, "/statements/")
}

// StripSignature removes exp/sign query params, returning the raw path.
func StripSignature(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Del("exp")
	q.Del("sign")
	u.RawQuery = q.Encode()
	return u.String()
}

// FriendlyError message.
func FriendlyError() string {
	return fmt.Sprintf("access denied or link expired")
}
