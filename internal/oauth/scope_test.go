package oauth_test

import (
	"testing"

	"github.com/timsgarage/tapu/internal/oauth"
)

func TestNormalizeScope(t *testing.T) {
	actual, ok := oauth.NormalizeScope("profile openid openid")
	if !ok {
		t.Fatal("expected scope to be accepted")
	}

	if actual != "openid profile" {
		t.Fatalf("unexpected normalized scope: %q", actual)
	}
}

func TestNormalizeScopeRequiresOpenID(t *testing.T) {
	if _, ok := oauth.NormalizeScope("profile"); ok {
		t.Fatal("expected scope without openid to be rejected")
	}
}

func TestNormalizeScopeRejectsUnknownScope(t *testing.T) {
	if _, ok := oauth.NormalizeScope("openid email"); ok {
		t.Fatal("expected unknown scope to be rejected")
	}
}
