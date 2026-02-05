package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"storage-api/internal/models"
	"storage-api/internal/repository"
	"storage-api/internal/service"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	// MaxUploadSize is 100MB
	MaxUploadSize = 100 << 20
)

type MediaHandler struct {
	svc     *service.MediaService
	userSvc *service.UserService
}

func NewMediaHandler(svc *service.MediaService, userSvc *service.UserService) *MediaHandler {
	return &MediaHandler{svc: svc, userSvc: userSvc}
}

// parseMediaID extracts and validates a UUID from the URL path parameter.
// Returns the parsed UUID, or writes an error response and returns false.
func parseMediaID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing id parameter",
		})
		return uuid.UUID{}, false
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid id parameter",
		})
		return uuid.UUID{}, false
	}

	return id, true
}

// getMediaItem fetches a media item by ID, handling not-found and error cases.
// Returns the item, or writes an error response and returns nil.
func (h *MediaHandler) getMediaItem(w http.ResponseWriter, r *http.Request, id uuid.UUID) *models.MediaItem {
	item, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": "media item not found",
			})
			return nil
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to get media: %v", err),
		})
		return nil
	}
	return item
}

// parseHouseholdID extracts and validates the household ID from the request header.
// Returns the parsed UUID, or writes an error response and returns false.
func parseHouseholdID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	householdIDStr := r.Header.Get("X-Household-ID")
	if householdIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "missing X-Household-ID header",
		})
		return uuid.UUID{}, false
	}

	householdID, err := uuid.Parse(householdIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid X-Household-ID header",
		})
		return uuid.UUID{}, false
	}

	return householdID, true
}

// Upload handles POST /media/upload
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	householdID, ok := parseHouseholdID(w, r)
	if !ok {
		return
	}

	// Get current user for uploader tracking
	currentUser := h.getCurrentUser(r)

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

	// Parse is_private flag (defaults to false/public)
	isPrivate := r.FormValue("is_private") == "true"

	// Save file to storage
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
		CleanupFiles(saved.FullPath, saved.PreviewFullPath, saved.WebFullPath)
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":    "file already exists",
			"existing": existing,
		})
		return
	}

	// Create media item
	item := &models.MediaItem{
		HouseholdID:      householdID,
		IsPrivate:        isPrivate,
		Path:             saved.RelativePath,
		Type:             mediaType,
		MimeType:         mimeType,
		SizeBytes:        saved.Size,
		SHA256:           saved.SHA256,
		PreviewPath:      saved.PreviewRelativePath,
		ThumbnailPath:    saved.ThumbnailRelPath,
		WebPath:          saved.WebRelPath,
		OriginalFilename: header.Filename,
	}

	// Set uploader if we have a current user
	if currentUser != nil {
		item.UploaderID = &currentUser.ID
	}

	// Populate metadata if extracted
	if saved.Metadata != nil {
		populateItemMetadata(item, saved.Metadata)
	}

	if err := h.svc.Create(r.Context(), item); err != nil {
		CleanupFiles(saved.FullPath, saved.PreviewFullPath, saved.WebFullPath)
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

func populateItemMetadata(item *models.MediaItem, meta *ImageMetadata) {
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

// List handles GET /media
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	householdID, ok := parseHouseholdID(w, r)
	if !ok {
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

	// Parse visibility and type filters
	visibility := r.URL.Query().Get("visibility")
	if visibility == "" {
		visibility = "all"
	}
	mediaType := r.URL.Query().Get("type")

	// Get current user ID from database (for visibility filtering)
	var userID *uuid.UUID
	if u := h.getCurrentUser(r); u != nil {
		userID = &u.ID
	}

	filter := repository.MediaListFilter{
		HouseholdID: householdID,
		UserID:      userID,
		Visibility:  visibility,
		MediaType:   mediaType,
		Page:        page,
		PageSize:    pageSize,
	}

	items, totalCount, err := h.svc.List(r.Context(), filter)
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

// getCurrentUser looks up the current user from Clerk claims
func (h *MediaHandler) getCurrentUser(r *http.Request) *models.User {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		return nil
	}

	// Fetch user from Clerk to get email
	clerkUser, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		return nil
	}

	// Get primary email
	var email string
	for _, e := range clerkUser.EmailAddresses {
		if e.ID == *clerkUser.PrimaryEmailAddressID {
			email = e.EmailAddress
			break
		}
	}

	if email == "" {
		return nil
	}

	// Look up user in database
	u, err := h.userSvc.GetByEmail(r.Context(), email)
	if err != nil {
		return nil
	}

	return u
}

// Get handles GET /media/{id}
func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseMediaID(w, r)
	if !ok {
		return
	}

	item := h.getMediaItem(w, r, id)
	if item == nil {
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"item": item})
}

// Download handles GET /media/{id}/download
// Serves web-optimized WebP for photos (fast), falls back to preview/original.
func (h *MediaHandler) Download(w http.ResponseWriter, r *http.Request) {
	id, ok := parseMediaID(w, r)
	if !ok {
		return
	}

	item := h.getMediaItem(w, r, id)
	if item == nil {
		return
	}

	// Priority: WebP (fast) > Preview (HEIC converted) > Original
	fullPath, contentType := resolveDownloadPath(item)

	if !fileExists(fullPath) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "file not found on disk",
		})
		return
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeFile(w, r, fullPath)
}

func resolveDownloadPath(item *models.MediaItem) (fullPath, contentType string) {
	basePath := getMediaBasePath()

	// Prefer web-optimized WebP for photos
	if item.WebPath != "" {
		webPath := filepath.Join(basePath, item.WebPath)
		if fileExists(webPath) {
			return webPath, "image/webp"
		}
	}

	// Fall back to preview (for HEIC files)
	if item.PreviewPath != "" {
		return filepath.Join(basePath, item.PreviewPath), "image/jpeg"
	}

	// Fall back to original
	return filepath.Join(basePath, item.Path), item.MimeType
}

// Original handles GET /media/{id}/original
func (h *MediaHandler) Original(w http.ResponseWriter, r *http.Request) {
	id, ok := parseMediaID(w, r)
	if !ok {
		return
	}

	item := h.getMediaItem(w, r, id)
	if item == nil {
		return
	}

	fullPath := filepath.Join(getMediaBasePath(), item.Path)

	if !fileExists(fullPath) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "file not found on disk",
		})
		return
	}

	if item.MimeType != "" {
		w.Header().Set("Content-Type", item.MimeType)
	}
	if item.OriginalFilename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", item.OriginalFilename))
	}
	http.ServeFile(w, r, fullPath)
}

// Thumbnail handles GET /media/{id}/thumbnail
func (h *MediaHandler) Thumbnail(w http.ResponseWriter, r *http.Request) {
	id, ok := parseMediaID(w, r)
	if !ok {
		return
	}

	item := h.getMediaItem(w, r, id)
	if item == nil {
		return
	}

	if item.ThumbnailPath == "" {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "thumbnail not available",
		})
		return
	}

	fullPath := filepath.Join(getMediaBasePath(), item.ThumbnailPath)

	if !fileExists(fullPath) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "thumbnail file not found on disk",
		})
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, fullPath)
}

// Delete handles DELETE /media/{id}
func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseMediaID(w, r)
	if !ok {
		return
	}

	item := h.getMediaItem(w, r, id)
	if item == nil {
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to delete from database: %v", err),
		})
		return
	}

	// Delete files from disk (don't fail if files don't exist)
	deleteMediaFiles(item)

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "deleted successfully",
	})
}

func deleteMediaFiles(item *models.MediaItem) {
	basePath := getMediaBasePath()
	os.Remove(filepath.Join(basePath, item.Path))
	if item.PreviewPath != "" {
		os.Remove(filepath.Join(basePath, item.PreviewPath))
	}
	if item.ThumbnailPath != "" {
		os.Remove(filepath.Join(basePath, item.ThumbnailPath))
	}
	if item.WebPath != "" {
		os.Remove(filepath.Join(basePath, item.WebPath))
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
