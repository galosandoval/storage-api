package repository

import (
	"context"

	"storage-api/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	Create(ctx context.Context, item *models.MediaItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.MediaItem, error)
	GetByPath(ctx context.Context, householdID uuid.UUID, path string) (*models.MediaItem, error)
	List(ctx context.Context, householdID uuid.UUID, page, pageSize int) ([]models.MediaItem, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type mediaRepo struct {
	db *gorm.DB
}

// NewMediaRepository creates a new MediaRepository
func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepo{db: db}
}

func (r *mediaRepo) Create(ctx context.Context, item *models.MediaItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *mediaRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.MediaItem, error) {
	var item models.MediaItem
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *mediaRepo) GetByPath(ctx context.Context, householdID uuid.UUID, path string) (*models.MediaItem, error) {
	var item models.MediaItem
	err := r.db.WithContext(ctx).
		Where("household_id = ? AND path = ?", householdID, path).
		First(&item).Error
	if err == gorm.ErrRecordNotFound {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *mediaRepo) List(ctx context.Context, householdID uuid.UUID, page, pageSize int) ([]models.MediaItem, int64, error) {
	var items []models.MediaItem
	var total int64

	db := r.db.WithContext(ctx).Model(&models.MediaItem{}).Where("household_id = ?", householdID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("COALESCE(taken_at, created_at) DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *mediaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.MediaItem{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrNotFound
	}
	return nil
}
