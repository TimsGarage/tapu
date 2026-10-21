package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	database *gorm.DB
}

func NewRepository(database *gorm.DB) *Repository {
	return &Repository{
		database: database,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	user *User,
) error {
	if err := r.database.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *Repository) FindByUsername(
	ctx context.Context,
	username string,
) (*User, error) {
	var user User

	err := r.database.WithContext(ctx).
		Where("username = ?", NormalizeUsername(username)).
		First(&user).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}

	return &user, nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*User, error) {
	var user User

	err := r.database.WithContext(ctx).
		Where("id = ?", id).
		First(&user).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find user by ID: %w", err)
	}

	return &user, nil
}
