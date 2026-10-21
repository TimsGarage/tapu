package oauth_test

import (
	"bytes"
	"testing"

	"github.com/timsgarage/tapu/internal/oauth"
)

func TestGenerateAuthorizationCode(t *testing.T) {
	first, firstHash, err := oauth.GenerateAuthorizationCode()
	if err != nil {
		t.Fatalf(
			"GenerateAuthorizationCode() returned an error: %v",
			err,
		)
	}

	second, secondHash, err := oauth.GenerateAuthorizationCode()
	if err != nil {
		t.Fatalf(
			"GenerateAuthorizationCode() returned an error: %v",
			err,
		)
	}

	if first == second {
		t.Fatal("expected unique authorization codes")
	}

	if bytes.Equal(firstHash, secondHash) {
		t.Fatal("expected unique authorization-code hashes")
	}

	if !oauth.ValidAuthorizationCode(first) {
		t.Fatal("expected generated authorization code to be valid")
	}

	if !bytes.Equal(
		firstHash,
		oauth.HashAuthorizationCode(first),
	) {
		t.Fatal("authorization-code hash does not match")
	}
}

func TestValidAuthorizationCodeRejectsInvalidValue(t *testing.T) {
	if oauth.ValidAuthorizationCode("not-a-valid-code") {
		t.Fatal("expected invalid authorization code to be rejected")
	}
}
