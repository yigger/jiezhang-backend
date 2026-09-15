package sessiontoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Generator struct{ Secret string }

func (g Generator) Generate(userID int64, sessionKey string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	raw := fmt.Sprintf(
		"%d|%s|%d|%s|%s",
		userID,
		sessionKey,
		time.Now().Unix(),
		hex.EncodeToString(randomBytes),
		g.Secret,
	)

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]), nil
}
