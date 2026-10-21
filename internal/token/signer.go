package token

import (
	"crypto/ed25519"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Signer struct {
	issuerURL      string
	keyID          string
	privateKey     ed25519.PrivateKey
	accessTokenTTL time.Duration
	idTokenTTL     time.Duration
}

type UserTokenInput struct {
	UserID          uuid.UUID
	Username        string
	ClientID        string
	ServiceAudience string
	Scope           string
	Nonce           string
}

func NewSigner(
	issuerURL string,
	keyID string,
	privateKey ed25519.PrivateKey,
	accessTokenTTL time.Duration,
	idTokenTTL time.Duration,
) *Signer {
	return &Signer{
		issuerURL:      issuerURL,
		keyID:          keyID,
		privateKey:     privateKey,
		accessTokenTTL: accessTokenTTL,
		idTokenTTL:     idTokenTTL,
	}
}

func (s *Signer) SignAccessToken(
	input UserTokenInput,
) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenTTL)

	username := ""
	if scopeContains(input.Scope, "profile") {
		username = input.Username
	}

	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuerURL,
			Subject:   input.UserID.String(),
			Audience:  jwt.ClaimStrings{input.ServiceAudience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
		TokenType: TypeAccessToken,
		ClientID:  input.ClientID,
		Username:  username,
		Scope:     input.Scope,
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodEdDSA,
		claims,
	)
	token.Header["kid"] = s.keyID
	token.Header["typ"] = "at+jwt"

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf(
			"sign access token: %w",
			err,
		)
	}

	return signed, expiresAt, nil
}

func (s *Signer) SignIDToken(
	input UserTokenInput,
) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.idTokenTTL)

	preferredUsername := ""
	if scopeContains(input.Scope, "profile") {
		preferredUsername = input.Username
	}

	claims := IDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuerURL,
			Subject:   input.UserID.String(),
			Audience:  jwt.ClaimStrings{input.ClientID},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
		Nonce:             input.Nonce,
		PreferredUsername: preferredUsername,
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodEdDSA,
		claims,
	)
	token.Header["kid"] = s.keyID
	token.Header["typ"] = "JWT"

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf(
			"sign ID token: %w",
			err,
		)
	}

	return signed, expiresAt, nil
}

func scopeContains(scope string, expected string) bool {
	for _, value := range strings.Fields(scope) {
		if value == expected {
			return true
		}
	}

	return false
}
