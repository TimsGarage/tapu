package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentTest        = "test"
	EnvironmentProduction  = "production"
)

type Config struct {
	Environment           string
	HTTPAddress           string
	DatabaseURL           string
	SigningPrivateKeyPath string
}

func Load() (Config, error) {
	cfg := Config{
		Environment: strings.ToLower(getEnv("APP_ENV", EnvironmentDevelopment)),
		HTTPAddress: getEnv("HTTP_ADDRESS", ":8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		SigningPrivateKeyPath: getEnv(
			"SIGNING_PRIVATE_KEY_PATH",
			"./secrets/signing-private.pem",
		),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error

	switch c.Environment {
	case EnvironmentDevelopment, EnvironmentTest, EnvironmentProduction:
	default:
		errs = append(
			errs,
			fmt.Errorf(
				"APP_ENV must be one of %q, %q, or %q",
				EnvironmentDevelopment,
				EnvironmentTest,
				EnvironmentProduction,
			),
		)
	}

	if strings.TrimSpace(c.HTTPAddress) == "" {
		errs = append(errs, errors.New("HTTP_ADDRESS must not be empty"))
	}

	if strings.TrimSpace(c.DatabaseURL) == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}

	if strings.TrimSpace(c.SigningPrivateKeyPath) == "" {
		errs = append(
			errs,
			errors.New("SIGNING_PRIVATE_KEY_PATH must not be empty"),
		)
	}

	return errors.Join(errs...)
}

func getEnv(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}
