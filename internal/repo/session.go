package repo

import "time"

type Cache interface {
	Get(string) (string, bool)
	Set(string, string, time.Duration)
}

// TokenGenerator creates opaque login credentials.
type TokenGenerator interface {
	Generate(userID int64, sessionKey string) (string, error)
}
