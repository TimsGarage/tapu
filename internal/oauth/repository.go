package oauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/timsgarage/tapu/internal/admin"
)

var (
	ErrClientNotFound      = errors.New("OAuth client not found")
	ErrServiceNotFound     = errors.New("service not found")
	ErrRedirectURINotFound = errors.New("redirect URI not found")
	ErrClientAccessDenied  = errors.New("client access denied")
	ErrUserAccessDenied    = errors.New("user access denied")
)

type Repository struct {
	database *gorm.DB
}

func NewRepository(database *gorm.DB) *Repository {
	return &Repository{
		database: database,
	}
}

func (r *Repository) FindEnabledClientByClientID(
	ctx context.Context,
	clientID string,
) (*admin.OAuthClient, error) {
	var client admin.OAuthClient

	err := r.database.WithContext(ctx).
		Where(
			"client_id = ? AND enabled = TRUE",
			clientID,
		).
		First(&client).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrClientNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find OAuth client: %w", err)
	}

	return &client, nil
}

func (r *Repository) RedirectURIAllowed(
	ctx context.Context,
	clientID uuid.UUID,
	redirectURI string,
) (bool, error) {
	var count int64

	err := r.database.WithContext(ctx).
		Model(&admin.OAuthClientRedirectURI{}).
		Where(
			"oauth_client_id = ? AND redirect_uri = ?",
			clientID,
			redirectURI,
		).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf(
			"check OAuth redirect URI: %w",
			err,
		)
	}

	return count == 1, nil
}

func (r *Repository) FindEnabledServiceByAudience(
	ctx context.Context,
	audience string,
) (*admin.Service, error) {
	var service admin.Service

	err := r.database.WithContext(ctx).
		Where(
			"audience = ? AND enabled = TRUE",
			audience,
		).
		First(&service).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrServiceNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find service: %w", err)
	}

	return &service, nil
}

func (r *Repository) ClientCanAccessService(
	ctx context.Context,
	clientID uuid.UUID,
	serviceID uuid.UUID,
) (bool, error) {
	var count int64

	err := r.database.WithContext(ctx).
		Model(&admin.OAuthClientService{}).
		Where(
			"oauth_client_id = ? AND service_id = ?",
			clientID,
			serviceID,
		).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf(
			"check client service access: %w",
			err,
		)
	}

	return count == 1, nil
}

func (r *Repository) UserCanAccessService(
	ctx context.Context,
	userID uuid.UUID,
	serviceID uuid.UUID,
) (bool, error) {
	var count int64

	err := r.database.WithContext(ctx).
		Model(&admin.UserServiceAccess{}).
		Where(
			"user_id = ? AND service_id = ?",
			userID,
			serviceID,
		).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf(
			"check user service access: %w",
			err,
		)
	}

	return count == 1, nil
}

func (r *Repository) CreateAuthorizationCode(
	ctx context.Context,
	code *AuthorizationCode,
) error {
	if err := r.database.WithContext(ctx).
		Create(code).
		Error; err != nil {
		return fmt.Errorf("create authorization code: %w", err)
	}

	return nil
}

func (r *Repository) DeleteExpiredAuthorizationCodes(
	ctx context.Context,
) error {
	now := time.Now().UTC()

	if err := r.database.WithContext(ctx).
		Delete(
			&AuthorizationCode{},
			"expires_at <= ? OR used_at IS NOT NULL",
			now,
		).
		Error; err != nil {
		return fmt.Errorf(
			"delete expired authorization codes: %w",
			err,
		)
	}

	return nil
}
