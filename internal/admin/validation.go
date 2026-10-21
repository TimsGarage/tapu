package admin

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var keyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,63}$`)

func validateServiceKey(value string) error {
	if !keyPattern.MatchString(value) {
		return errors.New(
			"service key must contain 2 to 64 lowercase letters, numbers, dots, underscores, or hyphens",
		)
	}

	return nil
}

func validateName(value string) error {
	length := len([]rune(strings.TrimSpace(value)))

	if length < 1 || length > 128 {
		return errors.New("name must contain between 1 and 128 characters")
	}

	return nil
}

func validateAudience(value string) error {
	value = strings.TrimSpace(value)

	if len(value) < 2 || len(value) > 255 {
		return errors.New("audience must contain between 2 and 255 characters")
	}

	if strings.ContainsAny(value, " \t\r\n ") {
		return errors.New("audience must not contain whitespace")
	}

	return nil
}

func validateClientType(value string) error {
	switch value {
	case ClientTypePublic, ClientTypeConfidential:
		return nil
	default:
		return errors.New("client type must be public or confidential")
	}
}

func validateRedirectURI(value string) error {
	if len(value) > 2048 {
		return errors.New("redirect URI is too long")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return errors.New("redirect URI is invalid")
	}

	if !parsed.IsAbs() || parsed.Host == "" {
		return errors.New("redirect URI must be absolute")
	}

	if parsed.User != nil || parsed.Fragment != "" {
		return errors.New(
			"redirect URI must not contain user information or a fragment",
		)
	}

	if parsed.Scheme == "https" {
		return nil
	}

	host := parsed.Hostname()

	if parsed.Scheme == "http" &&
		(host == "localhost" || net.ParseIP(host).IsLoopback()) {
		return nil
	}

	return errors.New(
		"redirect URI must use HTTPS, except for loopback development addresses",
	)
}
