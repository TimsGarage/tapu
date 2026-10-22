package authsdk

import "testing"

func TestConfigAcceptsCrossOriginFrontend(t *testing.T) {
	config := Config{
		IssuerURL:     "https://auth.example.com",
		ClientID:      "test-client",
		ClientSecret:  "test-secret",
		RedirectURL:   "https://api.example.com/auth/callback",
		Audience:      "project-a-api",
		AfterLoginURL: "https://app.example.com",
		AllowedOrigins: []string{
			"https://app.example.com",
		},
	}

	config.applyDefaults()

	if err := config.validate(); err != nil {
		t.Fatalf(
			"expected configuration to be valid: %v",
			err,
		)
	}
}

func TestConfigRejectsWildcardOrigin(t *testing.T) {
	config := Config{
		IssuerURL:     "https://auth.example.com",
		ClientID:      "test-client",
		ClientSecret:  "test-secret",
		RedirectURL:   "https://api.example.com/auth/callback",
		Audience:      "project-a-api",
		AfterLoginURL: "https://app.example.com",
		AllowedOrigins: []string{
			"https://*.example.com",
		},
	}

	config.applyDefaults()

	if err := config.validate(); err == nil {
		t.Fatal("expected wildcard origin to be rejected")
	}
}

func TestConfigRejectsOriginWithPath(t *testing.T) {
	config := Config{
		IssuerURL:     "https://auth.example.com",
		ClientID:      "test-client",
		ClientSecret:  "test-secret",
		RedirectURL:   "https://api.example.com/auth/callback",
		Audience:      "project-a-api",
		AfterLoginURL: "https://app.example.com",
		AllowedOrigins: []string{
			"https://app.example.com/admin",
		},
	}

	config.applyDefaults()

	if err := config.validate(); err == nil {
		t.Fatal("expected origin with path to be rejected")
	}
}
