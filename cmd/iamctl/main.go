package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/timsgarage/tapu/internal/audit"
	"github.com/timsgarage/tapu/internal/config"
	"github.com/timsgarage/tapu/internal/database"
	"github.com/timsgarage/tapu/internal/identity"
	"golang.org/x/term"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger, os.Args[1:]); err != nil {
		logger.Error("iamctl failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, arguments []string) error {
	if len(arguments) < 2 ||
		arguments[0] != "admin" ||
		arguments[1] != "create" {
		return errors.New(
			"usage: iamctl admin create --username <username>",
		)
	}

	flags := flag.NewFlagSet("admin create", flag.ContinueOnError)
	username := flags.String(
		"username",
		"",
		"username for the administrator",
	)
	passwordStdin := flags.Bool(
		"password-stdin",
		false,
		"read the password from standard input",
	)

	if err := flags.Parse(arguments[2:]); err != nil {
		return err
	}

	normalizedUsername := identity.NormalizeUsername(*username)

	if err := identity.ValidateUsername(normalizedUsername); err != nil {
		return err
	}

	password, err := readPassword(*passwordStdin)
	if err != nil {
		return err
	}

	hasher := identity.NewPasswordHasher()

	passwordHash, err := hasher.Hash(password)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	users := identity.NewRepository(db.GORM)

	now := time.Now().UTC()
	user := &identity.User{
		ID:                uuid.New(),
		Username:          normalizedUsername,
		PasswordHash:      passwordHash,
		IsAdmin:           true,
		Status:            identity.StatusActive,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := users.Create(ctx, user); err != nil {
		return err
	}

	auditRepository := audit.NewRepository(db.GORM)
	event := audit.NewEvent(audit.ActorCLI, "admin.created")
	event.TargetID = &user.ID

	targetType := "user"
	event.TargetType = &targetType

	if err := auditRepository.Record(
		ctx,
		event,
		map[string]any{
			"username": user.Username,
		},
	); err != nil {
		logger.Error(
			"administrator created but audit event failed",
			"error",
			err,
		)
	}

	logger.Info(
		"administrator created",
		"user_id",
		user.ID.String(),
		"username",
		user.Username,
	)

	return nil
}

func readPassword(fromStdin bool) (string, error) {
	if fromStdin {
		reader := bufio.NewReaderSize(os.Stdin, 2048)

		password, err := reader.ReadString('\n')
		if err != nil && len(password) == 0 {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}

		password = strings.TrimRight(password, "\r\n")

		if len(password) > 1024 {
			return "", errors.New("password input is too large")
		}

		return password, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New(
			"standard input is not a terminal; use --password-stdin",
		)
	}

	fmt.Fprint(os.Stderr, "Password: ")
	first, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}

	fmt.Fprint(os.Stderr, "Confirm password: ")
	second, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("confirm password: %w", err)
	}

	if string(first) != string(second) {
		return "", errors.New("passwords do not match")
	}

	return string(first), nil
}
