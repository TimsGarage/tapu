package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	minimumPasswordLength = 12
	maximumPasswordLength = 1024

	argonMemory      = 64 * 1024
	argonIterations  = 3
	argonParallelism = 2
	argonSaltLength  = 16
	argonKeyLength   = 32
)

var (
	ErrPasswordTooShort = fmt.Errorf(
		"password must contain at least %d characters",
		minimumPasswordLength,
	)
	ErrPasswordTooLong = fmt.Errorf(
		"password must not exceed %d bytes",
		maximumPasswordLength,
	)
	ErrInvalidPasswordHash = errors.New("invalid password hash")
)

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

func (h *PasswordHasher) Validate(password string) error {
	if len([]rune(password)) < minimumPasswordLength {
		return ErrPasswordTooShort
	}

	if len([]byte(password)) > maximumPasswordLength {
		return ErrPasswordTooLong
	}

	return nil
}

func (h *PasswordHasher) Hash(password string) (string, error) {
	if err := h.Validate(password); err != nil {
		return "", err
	}

	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	derivedKey := argon2.IDKey(
		[]byte(password),
		salt,
		argonIterations,
		argonMemory,
		argonParallelism,
		argonKeyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedKey := base64.RawStdEncoding.EncodeToString(derivedKey)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallelism,
		encodedSalt,
		encodedKey,
	)

	return encoded, nil
}

func (h *PasswordHasher) Compare(
	encodedHash string,
	password string,
) (bool, error) {
	parameters, salt, expectedKey, err := parsePasswordHash(encodedHash)
	if err != nil {
		return false, err
	}

	actualKey := argon2.IDKey(
		[]byte(password),
		salt,
		parameters.iterations,
		parameters.memory,
		parameters.parallelism,
		uint32(len(expectedKey)),
	)

	matches := subtle.ConstantTimeCompare(actualKey, expectedKey) == 1
	return matches, nil
}

type passwordParameters struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

func parsePasswordHash(
	encodedHash string,
) (passwordParameters, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return passwordParameters{}, nil, nil, ErrInvalidPasswordHash
	}

	if parts[1] != "argon2id" {
		return passwordParameters{}, nil, nil, ErrInvalidPasswordHash
	}

	versionParts := strings.SplitN(parts[2], "=", 2)
	if len(versionParts) != 2 || versionParts[0] != "v" {
		return passwordParameters{}, nil, nil, ErrInvalidPasswordHash
	}

	version, err := strconv.Atoi(versionParts[1])
	if err != nil || version != argon2.Version {
		return passwordParameters{}, nil, nil, ErrInvalidPasswordHash
	}

	parameters, err := parsePasswordParameters(parts[3])
	if err != nil {
		return passwordParameters{}, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) < 8 {
		return passwordParameters{}, nil, nil, ErrInvalidPasswordHash
	}

	key, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(key) < 16 {
		return passwordParameters{}, nil, nil, ErrInvalidPasswordHash
	}

	return parameters, salt, key, nil
}

func parsePasswordParameters(
	encoded string,
) (passwordParameters, error) {
	var memory uint32
	var iterations uint32
	var parallelism uint8

	count, err := fmt.Sscanf(
		encoded,
		"m=%d,t=%d,p=%d",
		&memory,
		&iterations,
		&parallelism,
	)
	if err != nil || count != 3 {
		return passwordParameters{}, ErrInvalidPasswordHash
	}

	if memory < 8*1024 || memory > 1024*1024 {
		return passwordParameters{}, ErrInvalidPasswordHash
	}

	if iterations < 1 || iterations > 10 {
		return passwordParameters{}, ErrInvalidPasswordHash
	}

	if parallelism < 1 || parallelism > 16 {
		return passwordParameters{}, ErrInvalidPasswordHash
	}

	return passwordParameters{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
	}, nil
}
