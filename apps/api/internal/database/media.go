package database

import (
	"context"
	"time"

	"storage-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateMediaItem inserts a new media item into the database
func CreateMediaItem(ctx context.Context, db *pgxpool.Pool, item *models.MediaItem) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var createdAt, updatedAt time.Time
	err := db.QueryRow(ctx, `
		INSERT INTO storage_items (
			household_id, path, type, mime_type, size_bytes, sha256,
			taken_at, width, height, duration_sec,
			preview_path, thumbnail_path, original_filename,
			camera_make, camera_model,
			latitude, longitude,
			orientation, iso, f_number, exposure_time, focal_length
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		RETURNING id::text, created_at, updated_at
	`,
		item.HouseholdID, item.Path, item.Type, item.MimeType, item.SizeBytes, item.SHA256,
		item.TakenAt, item.Width, item.Height, item.DurationSec,
		nullIfEmpty(item.PreviewPath), nullIfEmpty(item.ThumbnailPath), nullIfEmpty(item.OriginalFilename),
		nullIfEmpty(item.CameraMake), nullIfEmpty(item.CameraModel),
		item.Latitude, item.Longitude,
		nullIfZero(item.Orientation), nullIfZero(item.ISO), item.FNumber, nullIfEmpty(item.ExposureTime), item.FocalLength,
	).Scan(&item.ID, &createdAt, &updatedAt)

	if err != nil {
		return err
	}

	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	return nil
}

// GetMediaItemByID retrieves a media item by its ID
func GetMediaItemByID(ctx context.Context, db *pgxpool.Pool, id string) (models.MediaItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var item models.MediaItem
	var createdAt, updatedAt time.Time
	var takenAt *time.Time
	var previewPath, thumbnailPath, originalFilename, cameraMake, cameraModel, exposureTime *string
	var orientation, iso *int
	var latitude, longitude, fNumber, focalLength *float64

	err := db.QueryRow(ctx, `
		SELECT id::text, household_id::text, path, type,
		       COALESCE(mime_type, ''), COALESCE(size_bytes, 0), COALESCE(sha256, ''),
		       taken_at, COALESCE(width, 0), COALESCE(height, 0), COALESCE(duration_sec, 0),
		       created_at, updated_at,
		       preview_path, thumbnail_path, original_filename,
		       camera_make, camera_model,
		       latitude, longitude,
		       orientation, iso, f_number, exposure_time, focal_length
		FROM storage_items
		WHERE id = $1
	`, id).Scan(
		&item.ID, &item.HouseholdID, &item.Path, &item.Type,
		&item.MimeType, &item.SizeBytes, &item.SHA256,
		&takenAt, &item.Width, &item.Height, &item.DurationSec,
		&createdAt, &updatedAt,
		&previewPath, &thumbnailPath, &originalFilename,
		&cameraMake, &cameraModel,
		&latitude, &longitude,
		&orientation, &iso, &fNumber, &exposureTime, &focalLength,
	)

	if err != nil {
		return models.MediaItem{}, ErrNotFound
	}

	item.TakenAt = takenAt
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)

	// Set optional fields
	if previewPath != nil {
		item.PreviewPath = *previewPath
	}
	if thumbnailPath != nil {
		item.ThumbnailPath = *thumbnailPath
	}
	if originalFilename != nil {
		item.OriginalFilename = *originalFilename
	}
	if cameraMake != nil {
		item.CameraMake = *cameraMake
	}
	if cameraModel != nil {
		item.CameraModel = *cameraModel
	}
	item.Latitude = latitude
	item.Longitude = longitude
	if orientation != nil {
		item.Orientation = *orientation
	}
	if iso != nil {
		item.ISO = *iso
	}
	item.FNumber = fNumber
	if exposureTime != nil {
		item.ExposureTime = *exposureTime
	}
	item.FocalLength = focalLength

	return item, nil
}

// ListMediaItems retrieves paginated media items for a household
func ListMediaItems(ctx context.Context, db *pgxpool.Pool, householdID string, page, pageSize int) ([]models.MediaItem, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get total count
	var totalCount int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM storage_items WHERE household_id = $1
	`, householdID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated items
	offset := (page - 1) * pageSize
	rows, err := db.Query(ctx, `
		SELECT id::text, household_id::text, path, type,
		       COALESCE(mime_type, ''), COALESCE(size_bytes, 0), COALESCE(sha256, ''),
		       taken_at, COALESCE(width, 0), COALESCE(height, 0), COALESCE(duration_sec, 0),
		       created_at, updated_at,
		       preview_path, thumbnail_path, original_filename,
		       camera_make, camera_model,
		       latitude, longitude,
		       orientation, iso, f_number, exposure_time, focal_length
		FROM storage_items
		WHERE household_id = $1
		ORDER BY COALESCE(taken_at, created_at) DESC
		LIMIT $2 OFFSET $3
	`, householdID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.MediaItem
	for rows.Next() {
		var item models.MediaItem
		var createdAt, updatedAt time.Time
		var takenAt *time.Time
		var previewPath, thumbnailPath, originalFilename, cameraMake, cameraModel, exposureTime *string
		var orientation, iso *int
		var latitude, longitude, fNumber, focalLength *float64

		err := rows.Scan(
			&item.ID, &item.HouseholdID, &item.Path, &item.Type,
			&item.MimeType, &item.SizeBytes, &item.SHA256,
			&takenAt, &item.Width, &item.Height, &item.DurationSec,
			&createdAt, &updatedAt,
			&previewPath, &thumbnailPath, &originalFilename,
			&cameraMake, &cameraModel,
			&latitude, &longitude,
			&orientation, &iso, &fNumber, &exposureTime, &focalLength,
		)
		if err != nil {
			return nil, 0, err
		}

		item.TakenAt = takenAt
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)

		// Set optional fields
		if previewPath != nil {
			item.PreviewPath = *previewPath
		}
		if thumbnailPath != nil {
			item.ThumbnailPath = *thumbnailPath
		}
		if originalFilename != nil {
			item.OriginalFilename = *originalFilename
		}
		if cameraMake != nil {
			item.CameraMake = *cameraMake
		}
		if cameraModel != nil {
			item.CameraModel = *cameraModel
		}
		item.Latitude = latitude
		item.Longitude = longitude
		if orientation != nil {
			item.Orientation = *orientation
		}
		if iso != nil {
			item.ISO = *iso
		}
		item.FNumber = fNumber
		if exposureTime != nil {
			item.ExposureTime = *exposureTime
		}
		item.FocalLength = focalLength

		items = append(items, item)
	}

	return items, totalCount, nil
}

// DeleteMediaItem removes a media item by ID
func DeleteMediaItem(ctx context.Context, db *pgxpool.Pool, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	result, err := db.Exec(ctx, `DELETE FROM storage_items WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetMediaItemByPath retrieves a media item by household and path (for deduplication)
func GetMediaItemByPath(ctx context.Context, db *pgxpool.Pool, householdID, path string) (models.MediaItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var item models.MediaItem
	var createdAt, updatedAt time.Time
	var takenAt *time.Time
	var previewPath, thumbnailPath, originalFilename, cameraMake, cameraModel, exposureTime *string
	var orientation, iso *int
	var latitude, longitude, fNumber, focalLength *float64

	err := db.QueryRow(ctx, `
		SELECT id::text, household_id::text, path, type,
		       COALESCE(mime_type, ''), COALESCE(size_bytes, 0), COALESCE(sha256, ''),
		       taken_at, COALESCE(width, 0), COALESCE(height, 0), COALESCE(duration_sec, 0),
		       created_at, updated_at,
		       preview_path, thumbnail_path, original_filename,
		       camera_make, camera_model,
		       latitude, longitude,
		       orientation, iso, f_number, exposure_time, focal_length
		FROM storage_items
		WHERE household_id = $1 AND path = $2
	`, householdID, path).Scan(
		&item.ID, &item.HouseholdID, &item.Path, &item.Type,
		&item.MimeType, &item.SizeBytes, &item.SHA256,
		&takenAt, &item.Width, &item.Height, &item.DurationSec,
		&createdAt, &updatedAt,
		&previewPath, &thumbnailPath, &originalFilename,
		&cameraMake, &cameraModel,
		&latitude, &longitude,
		&orientation, &iso, &fNumber, &exposureTime, &focalLength,
	)

	if err != nil {
		return models.MediaItem{}, ErrNotFound
	}

	item.TakenAt = takenAt
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)

	// Set optional fields
	if previewPath != nil {
		item.PreviewPath = *previewPath
	}
	if thumbnailPath != nil {
		item.ThumbnailPath = *thumbnailPath
	}
	if originalFilename != nil {
		item.OriginalFilename = *originalFilename
	}
	if cameraMake != nil {
		item.CameraMake = *cameraMake
	}
	if cameraModel != nil {
		item.CameraModel = *cameraModel
	}
	item.Latitude = latitude
	item.Longitude = longitude
	if orientation != nil {
		item.Orientation = *orientation
	}
	if iso != nil {
		item.ISO = *iso
	}
	item.FNumber = fNumber
	if exposureTime != nil {
		item.ExposureTime = *exposureTime
	}
	item.FocalLength = focalLength

	return item, nil
}

// Helper functions to convert empty/zero values to nil for database
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullIfZero(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}
