package models

import (
	"time"

	"github.com/google/uuid"
)

type MediaItem struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HouseholdID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"householdId"`
	Path             string     `gorm:"size:512;not null" json:"path"`
	Type             string     `gorm:"size:10;not null" json:"type"` // "photo" or "video"
	MimeType         string     `gorm:"size:100" json:"mimeType,omitempty"`
	SizeBytes        int64      `json:"sizeBytes,omitempty"`
	SHA256           string     `gorm:"size:64" json:"sha256,omitempty"`
	TakenAt          *time.Time `gorm:"index" json:"takenAt,omitempty"`
	Width            int        `json:"width,omitempty"`
	Height           int        `json:"height,omitempty"`
	DurationSec      int        `json:"durationSec,omitempty"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Preview, thumbnail, and original file paths
	PreviewPath      string `gorm:"size:512" json:"previewPath,omitempty"`
	ThumbnailPath    string `gorm:"size:512" json:"thumbnailPath,omitempty"`
	OriginalFilename string `gorm:"size:255" json:"originalFilename,omitempty"`

	// Camera metadata (from EXIF)
	CameraMake  string `gorm:"size:100" json:"cameraMake,omitempty"`
	CameraModel string `gorm:"size:100" json:"cameraModel,omitempty"`

	// GPS coordinates (from EXIF)
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	// Technical metadata (from EXIF)
	Orientation  int      `json:"orientation,omitempty"`
	ISO          int      `json:"iso,omitempty"`
	FNumber      *float64 `json:"fNumber,omitempty"`
	ExposureTime string   `gorm:"size:20" json:"exposureTime,omitempty"`
	FocalLength  *float64 `json:"focalLength,omitempty"`
}

// TableName specifies the table name for GORM
func (MediaItem) TableName() string {
	return "storage_items"
}

// MediaListResponse is the API response for listing media
type MediaListResponse struct {
	Items      []MediaItem `json:"items"`
	TotalCount int64       `json:"totalCount"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
}
