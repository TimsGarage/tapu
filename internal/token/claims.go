package token

import (
	"github.com/golang-jwt/jwt/v5"
)

const (
	TypeAccessToken = "access_token"
	TypeIDToken     = "id_token"
)

type AccessTokenClaims struct {
	jwt.RegisteredClaims

	TokenType string `json:"token_type"`
	ClientID  string `json:"client_id"`
	Username  string `json:"username,omitempty"`
	Scope     string `json:"scope"`
}

type IDTokenClaims struct {
	jwt.RegisteredClaims

	Nonce             string `json:"nonce,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}
