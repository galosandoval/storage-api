package repository

import (
	"context"

	"storage-api/internal/models"

	"gorm.io/gorm"
)

// HouseholdRepository defines the interface for household data access
type HouseholdRepository interface {
	List(ctx context.Context) ([]models.Household, error)
}

type householdRepo struct {
	db *gorm.DB
}

// NewHouseholdRepository creates a new HouseholdRepository
func NewHouseholdRepository(db *gorm.DB) HouseholdRepository {
	return &householdRepo{db: db}
}

func (r *householdRepo) List(ctx context.Context) ([]models.Household, error) {
	var households []models.Household
	err := r.db.WithContext(ctx).Order("name").Find(&households).Error
	if err != nil {
		return nil, err
	}
	return households, nil
}
