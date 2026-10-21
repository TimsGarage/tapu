package session

import (
	"time"

	"github.com/google/uuid"
	"github.com/timsgarage/tapu/internal/identity"
)

type Session struct {
	ID                uuid.UUID     `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID     `gorm:"type:uuid;not null;index"`
	User              identity.User `gorm:"foreignKey:UserID"`
	TokenHash         []byte        `gorm:"type:bytea;not null;uniqueIndex"`
	CSRFTokenHash     []byte        `gorm:"type:bytea;not null"`
	CreatedAt         time.Time     `gorm:"not null"`
	LastSeenAt        time.Time     `gorm:"not null"`
	ExpiresAt         time.Time     `gorm:"not null;index"`
	AbsoluteExpiresAt time.Time     `gorm:"not null;index"`
}

func (Session) TableName() string {
	return "iam_sessions"
}
