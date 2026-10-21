package oauth

import (
	"net/http"
	"net/url"
	"testing"
)

func TestAuthorizationSuccessRedirect(t *testing.T) {
	actual, err := authorizationSuccessRedirect(
		"https://app.example.com/callback?existing=value",
		"test-code",
		"test-state",
	)
	if err != nil {
		t.Fatalf(
			"authorizationSuccessRedirect() returned an error: %v",
			err,
		)
	}

	parsed, err := url.Parse(actual)
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}

	query := parsed.Query()

	if query.Get("code") != "test-code" {
		t.Fatalf(
			"unexpected code: %q",
			query.Get("code"),
		)
	}

	if query.Get("state") != "test-state" {
		t.Fatalf(
			"unexpected state: %q",
			query.Get("state"),
		)
	}

	if query.Get("existing") != "value" {
		t.Fatal("existing redirect URI query was not retained")
	}
}

func TestBearerToken(t *testing.T) {
	request := newRequestWithAuthorization(
		"Bearer test-token",
	)

	actual, ok := bearerToken(request)
	if !ok {
		t.Fatal("expected bearer token to be accepted")
	}

	if actual != "test-token" {
		t.Fatalf("unexpected token: %q", actual)
	}
}

func newRequestWithAuthorization(
	value string,
) *http.Request {
	request, _ := http.NewRequest(
		http.MethodGet,
		"https://auth.example.com/oauth/userinfo",
		nil,
	)

	request.Header.Set("Authorization", value)
	return request
}
