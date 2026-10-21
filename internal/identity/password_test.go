package identity_test

import (
	"testing"

	"github.com/timsgarage/tapu/internal/identity"
)

func TestPasswordHashAndCompare(t *testing.T) {
	hasher := identity.NewPasswordHasher()
	password := "a-long-test-password"

	encoded, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() returned an error: %v", err)
	}

	matches, err := hasher.Compare(encoded, password)
	if err != nil {
		t.Fatalf("Compare() returned an error: %v", err)
	}

	if !matches {
		t.Fatal("expected password to match")
	}

	matches, err = hasher.Compare(encoded, "wrong-password-value")
	if err != nil {
		t.Fatalf("Compare() returned an error: %v", err)
	}

	if matches {
		t.Fatal("expected incorrect password not to match")
	}
}

func TestPasswordHashUsesRandomSalt(t *testing.T) {
	hasher := identity.NewPasswordHasher()
	password := "a-long-test-password"

	first, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("first Hash() returned an error: %v", err)
	}

	second, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("second Hash() returned an error: %v", err)
	}

	if first == second {
		t.Fatal("expected different hashes for the same password")
	}
}

func TestPasswordValidation(t *testing.T) {
	hasher := identity.NewPasswordHasher()

	if err := hasher.Validate("short"); err == nil {
		t.Fatal("expected short password validation to fail")
	}

	if err := hasher.Validate("this-is-long-enough"); err != nil {
		t.Fatalf("expected password validation to pass: %v", err)
	}
}
