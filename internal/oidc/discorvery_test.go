package oidc_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timsgarage/tapu/internal/oidc"
	"github.com/timsgarage/tapu/internal/token"
)

func TestDiscovery(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	handler := oidc.NewPublicHandler(
		"https://auth.example.com",
		token.NewJSONWebKeySet("test-key", publicKey),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/.well-known/openid-configuration",
		nil,
	)
	response := httptest.NewRecorder()

	handler.Discovery(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			response.Code,
		)
	}

	var document oidc.DiscoveryDocument

	if err := json.Unmarshal(
		response.Body.Bytes(),
		&document,
	); err != nil {
		t.Fatalf("decode discovery document: %v", err)
	}

	if document.Issuer != "https://auth.example.com" {
		t.Fatalf(
			"unexpected issuer: %q",
			document.Issuer,
		)
	}

	if document.AuthorizationEndpoint !=
		"https://auth.example.com/oauth/authorize" {
		t.Fatalf(
			"unexpected authorization endpoint: %q",
			document.AuthorizationEndpoint,
		)
	}
}

func TestJWKS(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	handler := oidc.NewPublicHandler(
		"https://auth.example.com",
		token.NewJSONWebKeySet("test-key", publicKey),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/.well-known/jwks.json",
		nil,
	)
	response := httptest.NewRecorder()

	handler.JWKS(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			response.Code,
		)
	}

	var keySet token.JSONWebKeySet

	if err := json.Unmarshal(
		response.Body.Bytes(),
		&keySet,
	); err != nil {
		t.Fatalf("decode JWKS: %v", err)
	}

	if len(keySet.Keys) != 1 {
		t.Fatalf(
			"expected one signing key, got %d",
			len(keySet.Keys),
		)
	}

	if keySet.Keys[0].KeyID != "test-key" {
		t.Fatalf(
			"unexpected key ID: %q",
			keySet.Keys[0].KeyID,
		)
	}
}
