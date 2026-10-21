package signing_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/timsgarage/tapu/internal/signing"
)

func TestGenerateWriteAndLoad(t *testing.T) {
	publicKey, privateKey, err := signing.Generate()
	if err != nil {
		t.Fatalf("Generate() returned an error: %v", err)
	}

	directory := t.TempDir()
	privatePath := filepath.Join(directory, "private.pem")
	publicPath := filepath.Join(directory, "public.pem")

	err = signing.WriteKeyPair(
		privatePath,
		publicPath,
		publicKey,
		privateKey,
	)
	if err != nil {
		t.Fatalf("WriteKeyPair() returned an error: %v", err)
	}

	loaded, err := signing.LoadPrivateKey(privatePath)
	if err != nil {
		t.Fatalf("LoadPrivateKey() returned an error: %v", err)
	}

	if !bytes.Equal(privateKey, loaded) {
		t.Fatal("loaded key does not match generated key")
	}

	if signing.KeyID(publicKey) == "" {
		t.Fatal("expected a non-empty key ID")
	}
}

func TestWriteKeyPairDoesNotOverwriteExistingPrivateKey(t *testing.T) {
	publicKey, privateKey, err := signing.Generate()
	if err != nil {
		t.Fatalf("Generate() returned an error: %v", err)
	}

	directory := t.TempDir()
	privatePath := filepath.Join(directory, "private.pem")
	publicPath := filepath.Join(directory, "public.pem")

	if err := signing.WriteKeyPair(
		privatePath,
		publicPath,
		publicKey,
		privateKey,
	); err != nil {
		t.Fatalf("first WriteKeyPair() returned an error: %v", err)
	}

	err = signing.WriteKeyPair(
		privatePath,
		publicPath,
		publicKey,
		privateKey,
	)
	if err == nil {
		t.Fatal("expected overwriting an existing key to fail")
	}
}
