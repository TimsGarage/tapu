package authsdk

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	PublicClient bool

	RedirectURL   string
	Audience      string
	AfterLoginURL string

	AllowedOrigins []string

	SessionCookieName     string
	TransactionCookieName string
	SecureCookies         bool

	SessionTTL     time.Duration
	TransactionTTL time.Duration

	HTTPClient *http.Client
	Store      SessionStore
}

func (c *Config) applyDefaults() {
	c.IssuerURL = strings.TrimRight(
		strings.TrimSpace(c.IssuerURL),
		"/",
	)
	c.ClientID = strings.TrimSpace(c.ClientID)
	c.RedirectURL = strings.TrimSpace(c.RedirectURL)
	c.Audience = strings.TrimSpace(c.Audience)

	for index, origin := range c.AllowedOrigins {
		c.AllowedOrigins[index] = strings.TrimSuffix(
			strings.TrimSpace(origin),
			"/",
		)
	}

	if c.AfterLoginURL == "" {
		c.AfterLoginURL = "/"
	}

	if c.SessionCookieName == "" {
		c.SessionCookieName = "app_session"
	}

	if c.TransactionCookieName == "" {
		c.TransactionCookieName = "app_auth_transaction"
	}

	if c.SessionTTL == 0 {
		c.SessionTTL = 8 * time.Hour
	}

	if c.TransactionTTL == 0 {
		c.TransactionTTL = 10 * time.Minute
	}

	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}

	if c.Store == nil {
		c.Store = NewMemorySessionStore()
	}
}

func (c Config) validate() error {
	var errs []error

	if err := validateAbsoluteHTTPURL(
		"IssuerURL",
		c.IssuerURL,
	); err != nil {
		errs = append(errs, err)
	}

	if err := validateAbsoluteHTTPURL(
		"RedirectURL",
		c.RedirectURL,
	); err != nil {
		errs = append(errs, err)
	}

	if c.ClientID == "" {
		errs = append(errs, errors.New("ClientID is required"))
	}

	if !c.PublicClient && c.ClientSecret == "" {
		errs = append(
			errs,
			errors.New(
				"ClientSecret is required for a confidential client",
			),
		)
	}

	if c.PublicClient && c.ClientSecret != "" {
		errs = append(
			errs,
			errors.New(
				"ClientSecret must be empty for a public client",
			),
		)
	}

	if c.Audience == "" {
		errs = append(errs, errors.New("Audience is required"))
	}

	if err := validateAfterLoginURL(c.AfterLoginURL); err != nil {
		errs = append(errs, err)
	}

	for _, origin := range c.AllowedOrigins {
		if err := validateOrigin(origin); err != nil {
			errs = append(errs, err)
		}
	}

	if c.SessionCookieName == c.TransactionCookieName {
		errs = append(
			errs,
			errors.New(
				"session and transaction cookie names must differ",
			),
		)
	}

	if c.SessionTTL <= 0 {
		errs = append(
			errs,
			errors.New("SessionTTL must be positive"),
		)
	}

	if c.TransactionTTL <= 0 ||
		c.TransactionTTL > 30*time.Minute {
		errs = append(
			errs,
			errors.New(
				"TransactionTTL must be positive and no longer than 30 minutes",
			),
		)
	}

	if c.HTTPClient == nil {
		errs = append(errs, errors.New("HTTPClient is required"))
	}

	if c.Store == nil {
		errs = append(errs, errors.New("Store is required"))
	}

	return errors.Join(errs...)
}

func validateAbsoluteHTTPURL(
	name string,
	value string,
) error {
	parsed, err := url.Parse(value)
	if err != nil ||
		!parsed.IsAbs() ||
		parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New(
			name + " must be an absolute HTTP or HTTPS URL",
		)
	}

	if parsed.User != nil || parsed.Fragment != "" {
		return errors.New(
			name + " must not contain user information or a fragment",
		)
	}

	return nil
}

func validLocalPath(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	return strings.HasPrefix(parsed.Path, "/") &&
		!strings.HasPrefix(parsed.Path, "//") &&
		!parsed.IsAbs() &&
		parsed.Host == "" &&
		parsed.Fragment == ""
}

func validateAfterLoginURL(value string) error {
	if validLocalPath(value) {
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil ||
		!parsed.IsAbs() ||
		parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New(
			"AfterLoginURL must be a local absolute path or an absolute HTTP or HTTPS URL",
		)
	}

	if parsed.User != nil || parsed.Fragment != "" {
		return errors.New(
			"AfterLoginURL must not contain user information or a fragment",
		)
	}

	return nil
}

func validateOrigin(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil ||
		!parsed.IsAbs() ||
		parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New(
			"AllowedOrigins entries must be absolute HTTP or HTTPS origins",
		)
	}

	if parsed.User != nil ||
		parsed.Path != "" ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return errors.New(
			"AllowedOrigins entries must contain only a scheme and host",
		)
	}

	if strings.Contains(parsed.Host, "*") {
		return errors.New(
			"AllowedOrigins entries must not contain wildcards",
		)
	}

	return nil
}
