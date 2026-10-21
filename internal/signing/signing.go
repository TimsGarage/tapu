package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	privateKeyPEMType = "PRIVATE KEY"
	publicKeyPEMType  = "PUBLIC KEY"
)

func Generate() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate Ed25519 key: %w", err)
	}

	return publicKey, privateKey, nil
}

func WriteKeyPair(
	privatePath string,
	publicPath string,
	publicKey ed25519.PublicKey,
	privateKey ed25519.PrivateKey,
) error {
	if err := writePrivateKey(privatePath, privateKey); err != nil {
		return err
	}

	if err := writePublicKey(publicPath, publicKey); err != nil {
		return err
	}

	return nil
}

func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private signing key: %w", err)
	}

	block, rest := pem.Decode(data)
	if block == nil {
		return nil, errors.New("private signing key is not valid PEM")
	}

	if len(rest) != 0 {
		return nil, errors.New("private signing key contains unexpected trailing data")
	}

	if block.Type != privateKeyPEMType {
		return nil, fmt.Errorf(
			"unexpected private key PEM type %q",
			block.Type,
		)
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private signing key: %w", err)
	}

	privateKey, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("private signing key is not an Ed25519 key")
	}

	return privateKey, nil
}

func KeyID(publicKey ed25519.PublicKey) string {
	sum := sha256.Sum256(publicKey)

	// A 128-bit prefix is sufficient as a non-secret key identifier.
	return base64.RawURLEncoding.EncodeToString(sum[:16])
}

func writePrivateKey(path string, privateKey ed25519.PrivateKey) error {
	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("encode private key: %w", err)
	}

	return writePEMFile(path, privateKeyPEMType, encoded, 0o600)
}

func writePublicKey(path string, publicKey ed25519.PublicKey) error {
	encoded, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("encode public key: %w", err)
	}

	return writePEMFile(path, publicKeyPEMType, encoded, 0o644)
}

func writePEMFile(
	path string,
	blockType string,
	data []byte,
	permissions os.FileMode,
) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		permissions,
	)
	if err != nil {
		return fmt.Errorf("create key file %q: %w", path, err)
	}

	writeErr := pem.Encode(file, &pem.Block{
		Type:  blockType,
		Bytes: data,
	})
	closeErr := file.Close()

	if writeErr != nil {
		return fmt.Errorf("write key file %q: %w", path, writeErr)
	}

	if closeErr != nil {
		return fmt.Errorf("close key file %q: %w", path, closeErr)
	}

	return nil
}
