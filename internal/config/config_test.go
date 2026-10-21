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

	if cfg.SigningPrivateKeyPath != "./secrets/signing-private.pem" {
		t.Fatalf(
			"unexpected signing key path: %q",
			cfg.SigningPrivateKeyPath,
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
