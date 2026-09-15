package sharetoken

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

// Codec preserves the existing double-base64 AES-CBC token format for compatibility.
type Codec struct{ Secret string }

var ErrInvalidToken = errors.New("invalid share token")

func (s Codec) Encrypt(plain []byte) (string, error) {
	if strings.TrimSpace(s.Secret) == "" {
		return "", ErrInvalidToken
	}

	key := sha256.Sum256([]byte(s.Secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	plain = pkcs7PadStatement(plain, aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plain))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plain)

	combined := append(iv, ciphertext...)
	first := base64.StdEncoding.EncodeToString(combined)
	second := base64.StdEncoding.EncodeToString([]byte(first))
	return second, nil
}
func (s Codec) Decrypt(token string) ([]byte, error) {
	if strings.TrimSpace(s.Secret) == "" {
		return nil, ErrInvalidToken
	}
	firstBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	combined, err := base64.StdEncoding.DecodeString(string(firstBytes))
	if err != nil {
		return nil, err
	}
	if len(combined) < aes.BlockSize || len(combined)%aes.BlockSize != 0 {
		return nil, ErrInvalidToken
	}

	key := sha256.Sum256([]byte(s.Secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	iv := combined[:aes.BlockSize]
	ciphertext := combined[aes.BlockSize:]
	plain := make([]byte, len(ciphertext))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plain, ciphertext)

	plain, err = pkcs7UnpadStatement(plain, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	return plain, nil
}
func pkcs7PadStatement(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := make([]byte, padding)
	for i := 0; i < padding; i++ {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}
func pkcs7UnpadStatement(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, ErrInvalidToken
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, ErrInvalidToken
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, ErrInvalidToken
		}
	}
	return data[:len(data)-padding], nil
}
