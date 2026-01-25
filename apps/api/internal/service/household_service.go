package service

import (
	"context"

	"storage-api/internal/models"
	"storage-api/internal/repository"
)

// HouseholdService handles business logic for household operations
type HouseholdService struct {
	repo repository.HouseholdRepository
}

// NewHouseholdService creates a new HouseholdService
func NewHouseholdService(repo repository.HouseholdRepository) *HouseholdService {
	return &HouseholdService{repo: repo}
}

// List returns all households
func (s *HouseholdService) List(ctx context.Context) ([]models.Household, error) {
	return s.repo.List(ctx)
}
