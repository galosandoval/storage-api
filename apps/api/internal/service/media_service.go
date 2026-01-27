package service

import (
	"context"

	"storage-api/internal/models"
	"storage-api/internal/repository"

	"github.com/google/uuid"
)

// MediaService handles business logic for media operations
type MediaService struct {
	repo repository.MediaRepository
}

// NewMediaService creates a new MediaService
func NewMediaService(repo repository.MediaRepository) *MediaService {
	return &MediaService{repo: repo}
}

// Create creates a new media item
func (s *MediaService) Create(ctx context.Context, item *models.MediaItem) error {
	return s.repo.Create(ctx, item)
}

// GetByID retrieves a media item by ID
func (s *MediaService) GetByID(ctx context.Context, id uuid.UUID) (*models.MediaItem, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByPath retrieves a media item by household and path
func (s *MediaService) GetByPath(ctx context.Context, householdID uuid.UUID, path string) (*models.MediaItem, error) {
	return s.repo.GetByPath(ctx, householdID, path)
}

// List retrieves paginated media items for a household with visibility filtering
func (s *MediaService) List(ctx context.Context, filter repository.MediaListFilter) ([]models.MediaItem, int64, error) {
	// Apply defaults and limits
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	return s.repo.List(ctx, filter)
}

// Delete removes a media item by ID
func (s *MediaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
