package oauth

import (
	"time"

	"github.com/google/uuid"
)

type AuthorizationCode struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CodeHash      []byte     `gorm:"type:bytea;not null;uniqueIndex" json:"-"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	OAuthClientID uuid.UUID  `gorm:"column:oauth_client_id;type:uuid;not null" json:"oauth_client_id"`
	ServiceID     uuid.UUID  `gorm:"type:uuid;not null" json:"service_id"`
	RedirectURI   string     `gorm:"not null" json:"redirect_uri"`
	Scope         string     `gorm:"not null" json:"scope"`
	Nonce         *string    `gorm:"type:text" json:"nonce,omitempty"`
	PKCEChallenge string     `gorm:"column:pkce_challenge;not null" json:"pkce_challenge"`
	PKCEMethod    string     `gorm:"column:pkce_method;not null" json:"pkce_method"`
	CreatedAt     time.Time  `gorm:"not null" json:"created_at"`
	ExpiresAt     time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt        *time.Time `gorm:"type:timestamptz" json:"used_at,omitempty"`
}

func (AuthorizationCode) TableName() string {
	return "oauth_authorization_codes"
}
