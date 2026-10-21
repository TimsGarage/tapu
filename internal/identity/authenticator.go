package identity

import (
	"context"
	"errors"
	"fmt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type Authenticator struct {
	users     *Repository
	passwords *PasswordHasher
	dummyHash string
}

func NewAuthenticator(
	users *Repository,
	passwords *PasswordHasher,
) (*Authenticator, error) {
	dummyHash, err := passwords.Hash(
		"dummy-password-that-is-never-used",
	)
	if err != nil {
		return nil, fmt.Errorf("create dummy password hash: %w", err)
	}

	return &Authenticator{
		users:     users,
		passwords: passwords,
		dummyHash: dummyHash,
	}, nil
}

func (a *Authenticator) Authenticate(
	ctx context.Context,
	username string,
	password string,
) (*User, error) {
	user, err := a.users.FindByUsername(ctx, username)

	if errors.Is(err, ErrUserNotFound) {
		_, compareErr := a.passwords.Compare(a.dummyHash, password)
		if compareErr != nil {
			return nil, compareErr
		}

		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	matches, err := a.passwords.Compare(user.PasswordHash, password)
	if err != nil {
		return nil, fmt.Errorf("compare password: %w", err)
	}

	if !matches || !user.IsActive() {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
