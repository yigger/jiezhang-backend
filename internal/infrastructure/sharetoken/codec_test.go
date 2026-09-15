package sharetoken

import (
	"bytes"
	"testing"
)

func TestLegacyTokenCompatibility(t *testing.T) {
	// Fixed IV fixture produced independently with OpenSSL AES-256-CBC/PKCS7.
	codec := Codec{Secret: "fixture-secret"}
	plain := []byte(`{"user_id":7,"account_book_id":3}`)
	got, err := codec.Decrypt("QUFFQ0F3UUZCZ2NJQ1FvTERBME9EeWhPNDh3TjVUUnJ1MlRZTnViQnZtY01RaUZrVlVQZGpxb1k5SEZMNDFKUzNXWWtSVmRZcEZWTVdUekoxMThLMnc9PQ==")
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("%s %v", got, err)
	}
	encoded, err := codec.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err = codec.Decrypt(encoded)
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("%s %v", got, err)
	}
	for _, invalid := range []string{"", "!!!", "YQ==", "YVdRPQ=="} {
		if _, err := codec.Decrypt(invalid); err == nil {
			t.Errorf("accepted %q", invalid)
		}
	}
}
