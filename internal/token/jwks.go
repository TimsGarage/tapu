package token

import (
	"crypto/ed25519"
	"encoding/base64"
)

type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}

type JSONWebKey struct {
	KeyType   string `json:"kty"`
	Curve     string `json:"crv"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	X         string `json:"x"`
}

func NewJSONWebKeySet(
	keyID string,
	publicKey ed25519.PublicKey,
) JSONWebKeySet {
	return JSONWebKeySet{
		Keys: []JSONWebKey{
			{
				KeyType:   "OKP",
				Curve:     "Ed25519",
				Use:       "sig",
				Algorithm: "EdDSA",
				KeyID:     keyID,
				X: base64.RawURLEncoding.EncodeToString(
					publicKey,
				),
			},
		},
	}
}
