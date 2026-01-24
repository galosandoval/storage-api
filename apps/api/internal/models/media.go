package models

import "time"

type MediaItem struct {
	ID          string     `json:"id"`
	HouseholdID string     `json:"householdId"`
	Path        string     `json:"path"`
	Type        string     `json:"type"` // "photo" or "video"
	MimeType    string     `json:"mimeType,omitempty"`
	SizeBytes   int64      `json:"sizeBytes,omitempty"`
	SHA256      string     `json:"sha256,omitempty"`
	TakenAt     *time.Time `json:"takenAt,omitempty"`
	Width       int        `json:"width,omitempty"`
	Height      int        `json:"height,omitempty"`
	DurationSec int        `json:"durationSec,omitempty"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`

	// Preview, thumbnail, and original file paths
	PreviewPath      string `json:"previewPath,omitempty"`      // Path to JPEG preview (for HEIC files)
	ThumbnailPath    string `json:"thumbnailPath,omitempty"`    // Path to thumbnail image
	OriginalFilename string `json:"originalFilename,omitempty"` // Original filename from upload

	// Camera metadata (from EXIF)
	CameraMake  string `json:"cameraMake,omitempty"`
	CameraModel string `json:"cameraModel,omitempty"`

	// GPS coordinates (from EXIF)
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	// Technical metadata (from EXIF)
	Orientation  int      `json:"orientation,omitempty"`  // EXIF orientation (1-8)
	ISO          int      `json:"iso,omitempty"`          // ISO sensitivity
	FNumber      *float64 `json:"fNumber,omitempty"`      // Aperture f-number
	ExposureTime string   `json:"exposureTime,omitempty"` // Shutter speed (e.g., "1/125")
	FocalLength  *float64 `json:"focalLength,omitempty"`  // Focal length in mm
}

type MediaUploadRequest struct {
	Type string `json:"type"` // "photo" or "video"
}

type MediaListResponse struct {
	Items      []MediaItem `json:"items"`
	TotalCount int         `json:"totalCount"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
}
