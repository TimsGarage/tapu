package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	ActorUser   = "user"
	ActorSystem = "system"
	ActorCLI    = "cli"
)

type Event struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey"`
	OccurredAt time.Time       `gorm:"not null"`
	ActorType  string          `gorm:"not null"`
	ActorID    *uuid.UUID      `gorm:"type:uuid"`
	Action     string          `gorm:"not null"`
	TargetType *string         `gorm:"type:text"`
	TargetID   *uuid.UUID      `gorm:"type:uuid"`
	RequestID  *string         `gorm:"type:text"`
	SourceIP   *string         `gorm:"type:text"`
	Metadata   json.RawMessage `gorm:"type:jsonb;not null"`
}

func (Event) TableName() string {
	return "audit_events"
}

func NewEvent(actorType string, action string) Event {
	return Event{
		ID:         uuid.New(),
		OccurredAt: time.Now().UTC(),
		ActorType:  actorType,
		Action:     action,
		Metadata:   json.RawMessage(`{}`),
	}
}
