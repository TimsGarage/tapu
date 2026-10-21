package oauth_test

import (
	"testing"

	"github.com/timsgarage/tapu/internal/oauth"
)

func TestPKCEKnownExample(t *testing.T) {
	// Example from RFC 7636, Appendix B.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	expected := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	actual := oauth.PKCEChallenge(verifier)

	if actual != expected {
		t.Fatalf(
			"expected challenge %q, got %q",
			expected,
			actual,
		)
	}

	if !oauth.VerifyPKCE(verifier, expected) {
		t.Fatal("expected verifier to match challenge")
	}
}

func TestPKCERejectsWrongVerifier(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := oauth.PKCEChallenge(verifier)

	wrongVerifier := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	if oauth.VerifyPKCE(wrongVerifier, challenge) {
		t.Fatal("expected wrong verifier to be rejected")
	}
}

func TestPKCERejectsShortVerifier(t *testing.T) {
	if oauth.ValidPKCEVerifier("too-short") {
		t.Fatal("expected short verifier to be rejected")
	}
}
