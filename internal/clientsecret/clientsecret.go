package clientsecret

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const secretBytes = 32

func Generate() (string, []byte, error) {
	value := make([]byte, secretBytes)

	if _, err := rand.Read(value); err != nil {
		return "", nil, fmt.Errorf("generate client secret: %w", err)
	}

	secret := base64.RawURLEncoding.EncodeToString(value)
	hash := Hash(secret)

	return secret, hash, nil
}

func Hash(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}
