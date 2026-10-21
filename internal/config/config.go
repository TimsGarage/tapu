package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
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
	IssuerURL             string
	SecureCookies         bool
	SessionIdleTTL        time.Duration
	SessionAbsoluteTTL    time.Duration
	AuthorizationCodeTTL  time.Duration
	AccessTokenTTL        time.Duration
	IDTokenTTL            time.Duration
}

func Load() (Config, error) {
	environment := strings.ToLower(
		getEnv("APP_ENV", EnvironmentDevelopment),
	)

	cfg := Config{
		Environment: environment,
		HTTPAddress: getEnv("HTTP_ADDRESS", ":8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		SigningPrivateKeyPath: getEnv(
			"SIGNING_PRIVATE_KEY_PATH",
			"./secrets/signing-private.pem",
		),
		IssuerURL: getEnv(
			"ISSUER_URL",
			"http://127.0.0.1:8080",
		),
		SecureCookies:        environment == EnvironmentProduction,
		SessionIdleTTL:       12 * time.Hour,
		SessionAbsoluteTTL:   7 * 24 * time.Hour,
		AuthorizationCodeTTL: time.Minute,
		AccessTokenTTL:       5 * time.Minute,
		IDTokenTTL:           5 * time.Minute,
	}

	cfg.IssuerURL = strings.TrimRight(cfg.IssuerURL, "/")

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
		errs = append(
			errs,
			errors.New("HTTP_ADDRESS must not be empty"),
		)
	}

	if strings.TrimSpace(c.DatabaseURL) == "" {
		errs = append(
			errs,
			errors.New("DATABASE_URL is required"),
		)
	}

	if strings.TrimSpace(c.SigningPrivateKeyPath) == "" {
		errs = append(
			errs,
			errors.New(
				"SIGNING_PRIVATE_KEY_PATH must not be empty",
			),
		)
	}

	if err := validateIssuerURL(
		c.IssuerURL,
		c.Environment,
	); err != nil {
		errs = append(errs, err)
	}

	if c.SessionIdleTTL <= 0 {
		errs = append(
			errs,
			errors.New("session idle lifetime must be positive"),
		)
	}

	if c.SessionAbsoluteTTL < c.SessionIdleTTL {
		errs = append(
			errs,
			errors.New(
				"session absolute lifetime must not be shorter than idle lifetime",
			),
		)
	}

	if c.AuthorizationCodeTTL <= 0 ||
		c.AuthorizationCodeTTL > 10*time.Minute {
		errs = append(
			errs,
			errors.New(
				"authorization code lifetime must be between zero and ten minutes",
			),
		)
	}

	if c.AccessTokenTTL <= 0 ||
		c.AccessTokenTTL > time.Hour {
		errs = append(
			errs,
			errors.New(
				"access token lifetime must be between zero and one hour",
			),
		)
	}

	if c.IDTokenTTL <= 0 ||
		c.IDTokenTTL > time.Hour {
		errs = append(
			errs,
			errors.New(
				"ID token lifetime must be between zero and one hour",
			),
		)
	}

	return errors.Join(errs...)
}

func validateIssuerURL(
	value string,
	environment string,
) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return errors.New("ISSUER_URL is invalid")
	}

	if !parsed.IsAbs() ||
		parsed.Host == "" ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" ||
		parsed.User != nil {
		return errors.New(
			"ISSUER_URL must be an absolute URL without query, fragment, or user information",
		)
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return errors.New(
			"ISSUER_URL must not contain a path",
		)
	}

	if environment == EnvironmentProduction &&
		parsed.Scheme != "https" {
		return errors.New(
			"ISSUER_URL must use HTTPS in production",
		)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New(
			"ISSUER_URL must use HTTP or HTTPS",
		)
	}

	return nil
}

func getEnv(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}
