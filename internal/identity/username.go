package identity

import (
	"errors"
	"regexp"
	"strings"
)

var usernamePattern = regexp.MustCompile(
	`^[a-z0-9][a-z0-9._-]{2,63}$`,
)

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func ValidateUsername(username string) error {
	normalized := NormalizeUsername(username)

	if !usernamePattern.MatchString(normalized) {
		return errors.New(
			"username must be between 3 and 64 characters and contain only lowercase letters, numbers, dots, underscores, or hyphens",
		)
	}

	return nil
}
