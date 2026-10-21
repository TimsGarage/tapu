package identity

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username          string    `gorm:"not null;uniqueIndex"`
	PasswordHash      string    `gorm:"not null"`
	IsAdmin           bool      `gorm:"not null"`
	Status            string    `gorm:"not null"`
	PasswordChangedAt time.Time `gorm:"not null"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`
}

func (User) TableName() string {
	return "users"
}

func (u User) IsActive() bool {
	return u.Status == StatusActive
}
