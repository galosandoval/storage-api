package service

import (
	"context"

	"storage-api/internal/models"
	"storage-api/internal/repository"
)

// UserService handles business logic for user operations
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetByExternalSub retrieves a user by external sub (for auth)
func (s *UserService) GetByExternalSub(ctx context.Context, sub string) (*models.User, error) {
	return s.repo.GetByExternalSub(ctx, sub)
}
