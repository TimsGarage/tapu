package oauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"regexp"
)

const PKCEMethodS256 = "S256"

var pkceValuePattern = regexp.MustCompile(
	`^[A-Za-z0-9._~-]{43,128}$`,
)

func ValidPKCEVerifier(verifier string) bool {
	return pkceValuePattern.MatchString(verifier)
}

func ValidPKCEChallenge(challenge string) bool {
	if len(challenge) != 43 {
		return false
	}

	decoded, err := base64.RawURLEncoding.Strict().
		DecodeString(challenge)

	return err == nil && len(decoded) == sha256.Size
}

func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func VerifyPKCE(
	verifier string,
	expectedChallenge string,
) bool {
	if !ValidPKCEVerifier(verifier) ||
		!ValidPKCEChallenge(expectedChallenge) {
		return false
	}

	actualChallenge := PKCEChallenge(verifier)

	return subtle.ConstantTimeCompare(
		[]byte(actualChallenge),
		[]byte(expectedChallenge),
	) == 1
}
