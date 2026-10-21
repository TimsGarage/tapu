package authsdk

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func randomValue(numberOfBytes int) (string, error) {
	value := make([]byte, numberOfBytes)

	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf(
			"generate cryptographic random value: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
