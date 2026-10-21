package token_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/timsgarage/tapu/internal/token"
)

func TestSignAccessToken(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}

	signer := token.NewSigner(
		"https://auth.example.com",
		"test-key",
		privateKey,
		5*time.Minute,
		5*time.Minute,
	)

	userID := uuid.New()

	signed, _, err := signer.SignAccessToken(
		token.UserTokenInput{
			UserID:          userID,
			Username:        "alice",
			ClientID:        "client-123",
			ServiceAudience: "project-a-api",
			Scope:           "openid profile",
		},
	)
	if err != nil {
		t.Fatalf("SignAccessToken() returned an error: %v", err)
	}

	claims := &token.AccessTokenClaims{}

	parsed, err := jwt.ParseWithClaims(
		signed,
		claims,
		func(parsedToken *jwt.Token) (any, error) {
			return publicKey, nil
		},
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithIssuer("https://auth.example.com"),
		jwt.WithAudience("project-a-api"),
	)
	if err != nil {
		t.Fatalf("parse signed access token: %v", err)
	}

	if !parsed.Valid {
		t.Fatal("expected access token to be valid")
	}

	if claims.Subject != userID.String() {
		t.Fatalf(
			"unexpected subject: %q",
			claims.Subject,
		)
	}

	if claims.ClientID != "client-123" {
		t.Fatalf(
			"unexpected client ID: %q",
			claims.ClientID,
		)
	}

	if claims.TokenType != token.TypeAccessToken {
		t.Fatalf(
			"unexpected token type: %q",
			claims.TokenType,
		)
	}
}

func TestSignIDToken(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}

	signer := token.NewSigner(
		"https://auth.example.com",
		"test-key",
		privateKey,
		5*time.Minute,
		5*time.Minute,
	)

	userID := uuid.New()

	signed, _, err := signer.SignIDToken(
		token.UserTokenInput{
			UserID:   userID,
			Username: "alice",
			ClientID: "client-123",
			Nonce:    "test-nonce",
		},
	)
	if err != nil {
		t.Fatalf("SignIDToken() returned an error: %v", err)
	}

	claims := &token.IDTokenClaims{}

	parsed, err := jwt.ParseWithClaims(
		signed,
		claims,
		func(parsedToken *jwt.Token) (any, error) {
			return publicKey, nil
		},
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithIssuer("https://auth.example.com"),
		jwt.WithAudience("client-123"),
	)
	if err != nil {
		t.Fatalf("parse signed ID token: %v", err)
	}

	if !parsed.Valid {
		t.Fatal("expected ID token to be valid")
	}

	if claims.Nonce != "test-nonce" {
		t.Fatalf("unexpected nonce: %q", claims.Nonce)
	}

	if claims.PreferredUsername != "" {
		t.Fatalf(
			"unexpected preferred username: %q",
			claims.PreferredUsername,
		)
	}
}
