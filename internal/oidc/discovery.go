package oidc

import (
	"encoding/json"
	"net/http"

	"github.com/timsgarage/tapu/internal/token"
)

type DiscoveryDocument struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JWKSURI                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgorithmsSupported []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
}

type PublicHandler struct {
	discovery DiscoveryDocument
	keySet    token.JSONWebKeySet
}

func NewPublicHandler(
	issuerURL string,
	keySet token.JSONWebKeySet,
) *PublicHandler {
	return &PublicHandler{
		discovery: DiscoveryDocument{
			Issuer: issuerURL,
			AuthorizationEndpoint: issuerURL +
				"/oauth/authorize",
			TokenEndpoint: issuerURL + "/oauth/token",
			UserInfoEndpoint: issuerURL +
				"/oauth/userinfo",
			JWKSURI: issuerURL +
				"/.well-known/jwks.json",
			ResponseTypesSupported: []string{
				"code",
			},
			ResponseModesSupported: []string{
				"query",
			},
			GrantTypesSupported: []string{
				"authorization_code",
			},
			SubjectTypesSupported: []string{
				"public",
			},
			IDTokenSigningAlgorithmsSupported: []string{
				"EdDSA",
			},
			TokenEndpointAuthMethodsSupported: []string{
				"client_secret_basic",
				"none",
			},
			CodeChallengeMethodsSupported: []string{
				"S256",
			},
			ScopesSupported: []string{
				"openid",
				"profile",
			},
			ClaimsSupported: []string{
				"iss",
				"sub",
				"aud",
				"exp",
				"iat",
				"nonce",
				"preferred_username",
			},
		},
		keySet: keySet,
	}
}

func (h *PublicHandler) Discovery(
	writer http.ResponseWriter,
	_ *http.Request,
) {
	writePublicJSON(writer, h.discovery)
}

func (h *PublicHandler) JWKS(
	writer http.ResponseWriter,
	_ *http.Request,
) {
	writer.Header().Set(
		"Cache-Control",
		"public, max-age=300",
	)

	writePublicJSON(writer, h.keySet)
}

func writePublicJSON(
	writer http.ResponseWriter,
	value any,
) {
	writer.Header().Set(
		"Content-Type",
		"application/json",
	)
	writer.Header().Set(
		"X-Content-Type-Options",
		"nosniff",
	)

	_ = json.NewEncoder(writer).Encode(value)
}
