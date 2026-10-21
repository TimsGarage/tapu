package config_test

import (
	"testing"

	"github.com/timsgarage/tapu/internal/config"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_ADDRESS", "")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SIGNING_PRIVATE_KEY_PATH", "")
	t.Setenv("ISSUER_URL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned an error: %v", err)
	}

	if cfg.Environment != config.EnvironmentDevelopment {
		t.Fatalf(
			"expected development environment, got %q",
			cfg.Environment,
		)
	}

	if cfg.HTTPAddress != ":8080" {
		t.Fatalf("expected :8080, got %q", cfg.HTTPAddress)
	}

	if cfg.IssuerURL != "http://127.0.0.1:8080" {
		t.Fatalf(
			"unexpected issuer URL: %q",
			cfg.IssuerURL,
		)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected an error for a missing DATABASE_URL")
	}
}

func TestLoadRejectsUnknownEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "unknown")
	t.Setenv("DATABASE_URL", "postgres://example")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected an error for an unknown environment")
	}
}

func TestProductionRequiresHTTPSIssuer(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("ISSUER_URL", "http://auth.example.com")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected HTTP issuer to be rejected in production")
	}
}

func TestProductionAcceptsHTTPSIssuer(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("ISSUER_URL", "https://auth.example.com")

	_, err := config.Load()
	if err != nil {
		t.Fatalf("expected HTTPS issuer to be accepted: %v", err)
	}
}

func TestIssuerRejectsQuery(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv(
		"ISSUER_URL",
		"http://127.0.0.1:8080?unexpected=true",
	)

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected issuer query to be rejected")
	}
}
