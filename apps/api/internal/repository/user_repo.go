package repository

import (
	"context"

	"storage-api/internal/models"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByExternalSub(ctx context.Context, sub string) (*models.User, error)
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetByExternalSub(ctx context.Context, sub string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("external_sub = ?", sub).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
