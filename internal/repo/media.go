package repo

// URLBuilder formats externally usable media URLs without exposing configuration.
type URLBuilder interface{ BuildPublicURL(string) string }
type TokenCodec interface {
	Encrypt([]byte) (string, error)
	Decrypt(string) ([]byte, error)
}
