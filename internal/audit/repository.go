package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	database *gorm.DB
}

func NewRepository(database *gorm.DB) *Repository {
	return &Repository{
		database: database,
	}
}

func (r *Repository) Record(
	ctx context.Context,
	event Event,
	metadata map[string]any,
) error {
	if metadata != nil {
		encoded, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("encode audit metadata: %w", err)
		}

		event.Metadata = encoded
	}

	if len(event.Metadata) == 0 {
		event.Metadata = json.RawMessage(`{}`)
	}

	if err := r.database.WithContext(ctx).Create(&event).Error; err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}

	return nil
}
