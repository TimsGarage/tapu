package token_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/timsgarage/tapu/internal/token"
)

func TestNewJSONWebKeySet(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	keySet := token.NewJSONWebKeySet(
		"test-key",
		publicKey,
	)

	if len(keySet.Keys) != 1 {
		t.Fatalf(
			"expected one key, got %d",
			len(keySet.Keys),
		)
	}

	key := keySet.Keys[0]

	if key.KeyType != "OKP" {
		t.Fatalf("unexpected key type: %q", key.KeyType)
	}

	if key.Curve != "Ed25519" {
		t.Fatalf("unexpected curve: %q", key.Curve)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		t.Fatalf("decode public key: %v", err)
	}

	if len(decoded) != ed25519.PublicKeySize {
		t.Fatalf(
			"unexpected public-key size: %d",
			len(decoded),
		)
	}
}
