package clientsecret_test

import (
	"bytes"
	"testing"

	"github.com/timsgarage/tapu/internal/clientsecret"
)

func TestGenerate(t *testing.T) {
	first, firstHash, err := clientsecret.Generate()
	if err != nil {
		t.Fatalf("Generate() returned an error: %v", err)
	}

	second, secondHash, err := clientsecret.Generate()
	if err != nil {
		t.Fatalf("Generate() returned an error: %v", err)
	}

	if first == second {
		t.Fatal("expected unique client secrets")
	}

	if bytes.Equal(firstHash, secondHash) {
		t.Fatal("expected unique client secret hashes")
	}

	if !bytes.Equal(firstHash, clientsecret.Hash(first)) {
		t.Fatal("stored hash does not match the generated secret")
	}
}
