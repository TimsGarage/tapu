package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const authorizationCodeBytes = 32

func GenerateAuthorizationCode() (
	rawCode string,
	codeHash []byte,
	err error,
) {
	value := make([]byte, authorizationCodeBytes)

	if _, err := rand.Read(value); err != nil {
		return "", nil, fmt.Errorf(
			"generate authorization code: %w",
			err,
		)
	}

	rawCode = base64.RawURLEncoding.EncodeToString(value)
	codeHash = HashAuthorizationCode(rawCode)

	return rawCode, codeHash, nil
}

func HashAuthorizationCode(rawCode string) []byte {
	sum := sha256.Sum256([]byte(rawCode))
	return sum[:]
}

func ValidAuthorizationCode(rawCode string) bool {
	decoded, err := base64.RawURLEncoding.Strict().
		DecodeString(rawCode)

	return err == nil && len(decoded) == authorizationCodeBytes
}
