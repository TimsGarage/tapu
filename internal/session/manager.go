package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const tokenLength = 32

var ErrUnauthenticated = errors.New("unauthenticated")

type CreatedSession struct {
	SessionToken string
	CSRFToken    string
	Session      *Session
}

type Manager struct {
	database    *gorm.DB
	idleTTL     time.Duration
	absoluteTTL time.Duration
}

func NewManager(
	database *gorm.DB,
	idleTTL time.Duration,
	absoluteTTL time.Duration,
) *Manager {
	return &Manager{
		database:    database,
		idleTTL:     idleTTL,
		absoluteTTL: absoluteTTL,
	}
}

func (m *Manager) Create(
	ctx context.Context,
	userID uuid.UUID,
) (*CreatedSession, error) {
	sessionToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	csrfToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	absoluteExpiration := now.Add(m.absoluteTTL)
	idleExpiration := now.Add(m.idleTTL)

	if idleExpiration.After(absoluteExpiration) {
		idleExpiration = absoluteExpiration
	}

	model := &Session{
		ID:                uuid.New(),
		UserID:            userID,
		TokenHash:         hashToken(sessionToken),
		CSRFTokenHash:     hashToken(csrfToken),
		CreatedAt:         now,
		LastSeenAt:        now,
		ExpiresAt:         idleExpiration,
		AbsoluteExpiresAt: absoluteExpiration,
	}

	if err := m.database.WithContext(ctx).Create(model).Error; err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &CreatedSession{
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
		Session:      model,
	}, nil
}

func (m *Manager) Authenticate(
	ctx context.Context,
	rawToken string,
) (*Session, error) {
	if !validToken(rawToken) {
		return nil, ErrUnauthenticated
	}

	now := time.Now().UTC()
	tokenHash := hashToken(rawToken)

	var found Session

	err := m.database.WithContext(ctx).
		Preload("User").
		Where(
			"token_hash = ? AND expires_at > ? AND absolute_expires_at > ?",
			tokenHash,
			now,
			now,
		).
		First(&found).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUnauthenticated
	}

	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}

	if !found.User.IsActive() {
		_ = m.database.WithContext(ctx).
			Delete(&Session{}, "id = ?", found.ID).
			Error

		return nil, ErrUnauthenticated
	}

	nextExpiration := now.Add(m.idleTTL)
	if nextExpiration.After(found.AbsoluteExpiresAt) {
		nextExpiration = found.AbsoluteExpiresAt
	}

	updateErr := m.database.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", found.ID).
		Updates(map[string]any{
			"last_seen_at": now,
			"expires_at":   nextExpiration,
		}).
		Error

	if updateErr != nil {
		return nil, fmt.Errorf("update session activity: %w", updateErr)
	}

	found.LastSeenAt = now
	found.ExpiresAt = nextExpiration

	return &found, nil
}

func (m *Manager) ValidateCSRF(
	found *Session,
	rawCSRFToken string,
) bool {
	if found == nil || !validToken(rawCSRFToken) {
		return false
	}

	actualHash := hashToken(rawCSRFToken)

	return subtle.ConstantTimeCompare(
		actualHash,
		found.CSRFTokenHash,
	) == 1
}

func (m *Manager) Delete(
	ctx context.Context,
	rawToken string,
) error {
	if !validToken(rawToken) {
		return nil
	}

	result := m.database.WithContext(ctx).
		Delete(&Session{}, "token_hash = ?", hashToken(rawToken))

	if result.Error != nil {
		return fmt.Errorf("delete session: %w", result.Error)
	}

	return nil
}

func (m *Manager) DeleteAllForUser(
	ctx context.Context,
	userID uuid.UUID,
) error {
	result := m.database.WithContext(ctx).
		Delete(&Session{}, "user_id = ?", userID)

	if result.Error != nil {
		return fmt.Errorf("delete user sessions: %w", result.Error)
	}

	return nil
}

func (m *Manager) DeleteExpired(ctx context.Context) error {
	now := time.Now().UTC()

	result := m.database.WithContext(ctx).
		Delete(
			&Session{},
			"expires_at <= ? OR absolute_expires_at <= ?",
			now,
			now,
		)

	if result.Error != nil {
		return fmt.Errorf("delete expired sessions: %w", result.Error)
	}

	return nil
}

func generateToken() (string, error) {
	value := make([]byte, tokenLength)

	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashToken(rawToken string) []byte {
	sum := sha256.Sum256([]byte(rawToken))
	return sum[:]
}

func validToken(rawToken string) bool {
	decoded, err := base64.RawURLEncoding.Strict().
		DecodeString(rawToken)

	return err == nil && len(decoded) == tokenLength
}
