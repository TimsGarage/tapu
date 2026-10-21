package identity_test

import (
	"testing"

	"github.com/timsgarage/tapu/internal/identity"
)

func TestNormalizeUsername(t *testing.T) {
	actual := identity.NormalizeUsername("  Alice.Example  ")
	expected := "alice.example"

	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		shouldFail bool
	}{
		{
			name:     "simple username",
			username: "alice",
		},
		{
			name:     "allowed punctuation",
			username: "alice_example-1.test",
		},
		{
			name:       "too short",
			username:   "ab",
			shouldFail: true,
		},
		{
			name:       "starts with punctuation",
			username:   "_alice",
			shouldFail: true,
		},
		{
			name:       "contains spaces",
			username:   "alice example",
			shouldFail: true,
		},
		{
			name:       "contains uppercase before normalization",
			username:   "Alice",
			shouldFail: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := identity.ValidateUsername(test.username)

			if test.shouldFail && err == nil {
				t.Fatal("expected validation to fail")
			}

			if !test.shouldFail && err != nil {
				t.Fatalf("expected validation to pass: %v", err)
			}
		})
	}
}
