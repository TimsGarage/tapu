package admin

import (
	"time"

	"github.com/google/uuid"
)

const (
	ClientTypePublic       = "public"
	ClientTypeConfidential = "confidential"
)

type Service struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Key       string    `gorm:"not null;uniqueIndex" json:"key"`
	Name      string    `gorm:"not null" json:"name"`
	Audience  string    `gorm:"not null;uniqueIndex" json:"audience"`
	Enabled   bool      `gorm:"not null" json:"enabled"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (Service) TableName() string {
	return "services"
}

type UserServiceAccess struct {
	UserID    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	ServiceID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"service_id"`
	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time  `gorm:"not null" json:"created_at"`
}

func (UserServiceAccess) TableName() string {
	return "user_service_access"
}

type OAuthClient struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ClientID   string    `gorm:"not null;uniqueIndex" json:"client_id"`
	Name       string    `gorm:"not null" json:"name"`
	ClientType string    `gorm:"not null" json:"client_type"`
	SecretHash []byte    `gorm:"type:bytea" json:"secret_hash"`
	Enabled    bool      `gorm:"not null" json:"enabled"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`
}

func (OAuthClient) TableName() string {
	return "oauth_clients"
}

type OAuthClientRedirectURI struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OAuthClientID uuid.UUID `gorm:"column:oauth_client_id;type:uuid;not null" json:"oauth_client_id"`
	RedirectURI   string    `gorm:"not null" json:"redirect_uri"`
}

func (OAuthClientRedirectURI) TableName() string {
	return "oauth_client_redirect_uris"
}

type OAuthClientService struct {
	OAuthClientID uuid.UUID `gorm:"column:oauth_client_id;type:uuid;primaryKey" json:"oauth_client_id"`
	ServiceID     uuid.UUID `gorm:"column:service_id;type:uuid;primaryKey" json:"service_id"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
}

func (OAuthClientService) TableName() string {
	return "oauth_client_services"
}
