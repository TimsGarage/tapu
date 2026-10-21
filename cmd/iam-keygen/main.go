package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/timsgarage/tapu/internal/signing"
)

func main() {
	privatePath := flag.String(
		"private",
		"./secrets/signing-private.pem",
		"path for the private signing key",
	)
	publicPath := flag.String(
		"public",
		"./secrets/signing-public.pem",
		"path for the public signing key",
	)

	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	publicKey, privateKey, err := signing.Generate()
	if err != nil {
		logger.Error("generate signing key", "error", err)
		os.Exit(1)
	}

	if err := signing.WriteKeyPair(
		*privatePath,
		*publicPath,
		publicKey,
		privateKey,
	); err != nil {
		logger.Error("write signing key pair", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Signing key generated successfully.")
	fmt.Printf("Private key: %s", *privatePath)
	fmt.Printf("Public key:  %s", *publicPath)
	fmt.Printf("Key ID:      %s", signing.KeyID(publicKey))
}
