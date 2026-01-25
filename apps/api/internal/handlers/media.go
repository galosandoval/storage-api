package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"storage-api/internal/models"
	"storage-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	// MaxUploadSize is 100MB
	MaxUploadSize = 100 << 20
)

type MediaHandler struct {
	svc *service.MediaService
}

func NewMediaHandler(svc *service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

// Upload handles POST /v1/media/upload
// Expects multipart form with:
// - file: the media file
// - type: "photo" or "video" (optional, auto-detected from mime type)
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Get household from header (dev mode)
	householdIDStr := r.Header.Get("X-Household-ID")
	if householdIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing X-Household-ID header",
		})
		return
	}

	householdID, err := uuid.Parse(householdIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid X-Household-ID header",
		})
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": fmt.Sprintf("file too large or invalid form: %v", err),
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing 'file' in form data",
		})
		return
	}
	defer file.Close()

	// Determine media type
	mediaType := r.FormValue("type")
	mimeType := header.Header.Get("Content-Type")
	if mediaType == "" {
		mediaType = detectMediaType(mimeType)
	}
	if mediaType != "photo" && mediaType != "video" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "type must be 'photo' or 'video'",
		})
		return
	}

	// Save file to storage (creates directories, computes hash, extracts metadata, converts HEIC)
	saved, err := SaveUploadedFile(file, mediaType, header.Filename, mimeType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Check if file already exists (by path)
	existing, err := h.svc.GetByPath(r.Context(), householdID, saved.RelativePath)
	if err == nil {
		CleanupFiles(saved.FullPath, saved.PreviewFullPath) // Remove the just-saved files
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":    "file already exists",
			"existing": existing,
		})
		return
	}

	// Create media item
	item := &models.MediaItem{
		HouseholdID:      householdID,
		Path:             saved.RelativePath,
		Type:             mediaType,
		MimeType:         mimeType,
		SizeBytes:        saved.Size,
		SHA256:           saved.SHA256,
		PreviewPath:      saved.PreviewRelativePath,
		ThumbnailPath:    saved.ThumbnailRelPath,
		OriginalFilename: header.Filename,
	}

	// Populate metadata if extracted
	if saved.Metadata != nil {
		meta := saved.Metadata
		item.TakenAt = meta.TakenAt
		item.Width = meta.Width
		item.Height = meta.Height
		item.CameraMake = meta.CameraMake
		item.CameraModel = meta.CameraModel
		item.Latitude = meta.Latitude
		item.Longitude = meta.Longitude
		item.Orientation = meta.Orientation
		item.ISO = meta.ISO
		item.FNumber = meta.FNumber
		item.ExposureTime = meta.ExposureTime
		item.FocalLength = meta.FocalLength
	}

	if err := h.svc.Create(r.Context(), item); err != nil {
		CleanupFiles(saved.FullPath, saved.PreviewFullPath) // Clean up on failure
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to save to database: %v", err),
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "upload successful",
		"item":    item,
	})
}

// List handles GET /v1/media
// Query params: page (default 1), pageSize (default 20, max 100)
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	householdIDStr := r.Header.Get("X-Household-ID")
	if householdIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing X-Household-ID header",
		})
		return
	}

	householdID, err := uuid.Parse(householdIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid X-Household-ID header",
		})
		return
	}

	// Parse pagination params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, totalCount, err := h.svc.List(r.Context(), householdID, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to list media: %v", err),
		})
		return
	}

	if items == nil {
		items = []models.MediaItem{}
	}

	writeJSON(w, http.StatusOK, models.MediaListResponse{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

// Get handles GET /v1/media/{id}
// Returns the media item metadata
func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing id parameter",
		})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid id parameter",
		})
		return
	}

	item, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": "media item not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to get media: %v", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"item": item})
}

// Download handles GET /v1/media/{id}/download
// Serves the web-friendly version (JPEG preview for HEIC, original for others)
func (h *MediaHandler) Download(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing id parameter",
		})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid id parameter",
		})
		return
	}

	item, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": "media item not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to get media: %v", err),
		})
		return
	}

	// Use preview if available (for HEIC files), otherwise use original
	var fullPath string
	var contentType string
	if item.PreviewPath != "" {
		fullPath = filepath.Join(getMediaBasePath(), item.PreviewPath)
		contentType = "image/jpeg"
	} else {
		fullPath = filepath.Join(getMediaBasePath(), item.Path)
		contentType = item.MimeType
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "file not found on disk",
		})
		return
	}

	// Set content type
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// Original handles GET /v1/media/{id}/original
// Serves the original file (HEIC, etc.) for download/archival
func (h *MediaHandler) Original(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing id parameter",
		})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid id parameter",
		})
		return
	}

	item, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": "media item not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to get media: %v", err),
		})
		return
	}

	fullPath := filepath.Join(getMediaBasePath(), item.Path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "file not found on disk",
		})
		return
	}

	// Set content type
	if item.MimeType != "" {
		w.Header().Set("Content-Type", item.MimeType)
	}

	// Set content-disposition for download
	if item.OriginalFilename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", item.OriginalFilename))
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// Thumbnail handles GET /v1/media/{id}/thumbnail
// Serves the thumbnail image with aggressive caching headers
func (h *MediaHandler) Thumbnail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing id parameter",
		})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid id parameter",
		})
		return
	}

	item, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": "media item not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to get media: %v", err),
		})
		return
	}

	// Check if thumbnail exists
	if item.ThumbnailPath == "" {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "thumbnail not available",
		})
		return
	}

	fullPath := filepath.Join(getMediaBasePath(), item.ThumbnailPath)

	// Check if file exists on disk
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "thumbnail file not found on disk",
		})
		return
	}

	// Set aggressive caching headers (1 year, immutable)
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// Delete handles DELETE /v1/media/{id}
func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing id parameter",
		})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid id parameter",
		})
		return
	}

	// Get item first to know the file paths
	item, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": "media item not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to get media: %v", err),
		})
		return
	}

	// Delete from database
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to delete from database: %v", err),
		})
		return
	}

	// Delete files from disk (don't fail if files don't exist)
	fullPath := filepath.Join(getMediaBasePath(), item.Path)
	os.Remove(fullPath)

	// Also delete preview if exists
	if item.PreviewPath != "" {
		previewFullPath := filepath.Join(getMediaBasePath(), item.PreviewPath)
		os.Remove(previewFullPath)
	}

	// Also delete thumbnail if exists
	if item.ThumbnailPath != "" {
		thumbnailFullPath := filepath.Join(getMediaBasePath(), item.ThumbnailPath)
		os.Remove(thumbnailFullPath)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "deleted successfully",
	})
}
