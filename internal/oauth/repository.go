package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/timsgarage/tapu/internal/admin"
	"github.com/timsgarage/tapu/internal/identity"
)

type ConsumeAuthorizationCodeInput struct {
	RawCode       string
	OAuthClientID uuid.UUID
	RedirectURI   string
	PKCEVerifier  string
}

type ConsumedAuthorizationCode struct {
	Code    AuthorizationCode
	User    identity.User
	Client  admin.OAuthClient
	Service admin.Service
}

var (
	ErrClientNotFound      = errors.New("OAuth client not found")
	ErrServiceNotFound     = errors.New("service not found")
	ErrRedirectURINotFound = errors.New("redirect URI not found")
	ErrClientAccessDenied  = errors.New("client access denied")
	ErrUserAccessDenied    = errors.New("user access denied")
	ErrInvalidGrant        = errors.New("invalid authorization grant")
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

func (r *Repository) ConsumeAuthorizationCode(
	ctx context.Context,
	input ConsumeAuthorizationCodeInput,
) (*ConsumedAuthorizationCode, error) {
	if !ValidAuthorizationCode(input.RawCode) ||
		!ValidPKCEVerifier(input.PKCEVerifier) {
		return nil, ErrInvalidGrant
	}

	codeHash := HashAuthorizationCode(input.RawCode)
	now := time.Now().UTC()

	var consumed ConsumedAuthorizationCode

	err := r.database.WithContext(ctx).Transaction(
		func(transaction *gorm.DB) error {
			var code AuthorizationCode

			err := transaction.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where(
					"code_hash = ? AND used_at IS NULL AND expires_at > ?",
					codeHash,
					now,
				).
				First(&code).
				Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidGrant
			}

			if err != nil {
				return fmt.Errorf(
					"lock authorization code: %w",
					err,
				)
			}

			if code.OAuthClientID != input.OAuthClientID ||
				code.RedirectURI != input.RedirectURI ||
				!VerifyPKCE(
					input.PKCEVerifier,
					code.PKCEChallenge,
				) {
				return ErrInvalidGrant
			}

			var user identity.User
			if err := transaction.
				Where(
					"id = ? AND status = ?",
					code.UserID,
					identity.StatusActive,
				).
				First(&user).
				Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrInvalidGrant
				}

				return fmt.Errorf(
					"find authorization user: %w",
					err,
				)
			}

			var client admin.OAuthClient
			if err := transaction.
				Where(
					"id = ? AND enabled = TRUE",
					code.OAuthClientID,
				).
				First(&client).
				Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrInvalidGrant
				}

				return fmt.Errorf(
					"find authorization client: %w",
					err,
				)
			}

			var service admin.Service
			if err := transaction.
				Where(
					"id = ? AND enabled = TRUE",
					code.ServiceID,
				).
				First(&service).
				Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrInvalidGrant
				}

				return fmt.Errorf(
					"find authorization service: %w",
					err,
				)
			}

			var clientAccessCount int64
			if err := transaction.
				Model(&admin.OAuthClientService{}).
				Where(
					"oauth_client_id = ? AND service_id = ?",
					client.ID,
					service.ID,
				).
				Count(&clientAccessCount).
				Error; err != nil {
				return fmt.Errorf(
					"check client access: %w",
					err,
				)
			}

			if clientAccessCount != 1 {
				return ErrInvalidGrant
			}

			var userAccessCount int64
			if err := transaction.
				Model(&admin.UserServiceAccess{}).
				Where(
					"user_id = ? AND service_id = ?",
					user.ID,
					service.ID,
				).
				Count(&userAccessCount).
				Error; err != nil {
				return fmt.Errorf(
					"check user access: %w",
					err,
				)
			}

			if userAccessCount != 1 {
				return ErrInvalidGrant
			}

			result := transaction.
				Model(&AuthorizationCode{}).
				Where(
					"id = ? AND used_at IS NULL",
					code.ID,
				).
				Update("used_at", now)

			if result.Error != nil {
				return fmt.Errorf(
					"consume authorization code: %w",
					result.Error,
				)
			}

			if result.RowsAffected != 1 {
				return ErrInvalidGrant
			}

			code.UsedAt = &now

			consumed = ConsumedAuthorizationCode{
				Code:    code,
				User:    user,
				Client:  client,
				Service: service,
			}

			return nil
		},
	)

	if errors.Is(err, ErrInvalidGrant) {
		return nil, ErrInvalidGrant
	}

	if err != nil {
		return nil, err
	}

	return &consumed, nil
}

func (r *Repository) FindActiveUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (*identity.User, error) {
	var user identity.User

	err := r.database.WithContext(ctx).
		Where(
			"id = ? AND status = ?",
			userID,
			identity.StatusActive,
		).
		First(&user).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, identity.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find active user: %w", err)
	}

	return &user, nil
}

func (r *Repository) PublicClientOriginAllowed(
	ctx context.Context,
	origin string,
) (bool, error) {
	requestOrigin, err := parseOrigin(origin)
	if err != nil {
		return false, nil
	}

	var redirectURIs []string

	err = r.database.WithContext(ctx).
		Table("oauth_client_redirect_uris").
		Select("oauth_client_redirect_uris.redirect_uri").
		Joins(`
			JOIN oauth_clients
			  ON oauth_clients.id =
			     oauth_client_redirect_uris.oauth_client_id
		`).
		Where(
			"oauth_clients.client_type = ? AND oauth_clients.enabled = TRUE",
			admin.ClientTypePublic,
		).
		Pluck(
			"oauth_client_redirect_uris.redirect_uri",
			&redirectURIs,
		).
		Error

	if err != nil {
		return false, fmt.Errorf(
			"find public-client redirect URIs: %w",
			err,
		)
	}

	for _, redirectURI := range redirectURIs {
		parsed, err := url.Parse(redirectURI)
		if err != nil {
			continue
		}

		registeredOrigin := strings.ToLower(
			parsed.Scheme + "://" + parsed.Host,
		)

		if registeredOrigin == requestOrigin {
			return true, nil
		}
	}

	return false, nil
}

func parseOrigin(value string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", err
	}

	if !parsed.IsAbs() ||
		parsed.Host == "" ||
		parsed.User != nil ||
		parsed.Path != "" ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return "", errors.New("invalid origin")
	}

	if parsed.Scheme != "http" &&
		parsed.Scheme != "https" {
		return "", errors.New("unsupported origin scheme")
	}

	if strings.Contains(parsed.Host, "*") {
		return "", errors.New(
			"origin must not contain wildcards",
		)
	}

	return strings.ToLower(
		parsed.Scheme + "://" + parsed.Host,
	), nil
}
