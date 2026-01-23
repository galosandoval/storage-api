package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SavedFile contains metadata about a successfully saved file.
type SavedFile struct {
	RelativePath        string
	FullPath            string
	Size                int64
	SHA256              string
	PreviewRelativePath string // Path to JPEG preview (for HEIC files)
	PreviewFullPath     string // Full path to preview

	// Extracted metadata
	Metadata *ImageMetadata
}

// SaveUploadedFile saves a file to the media storage.
// It creates necessary directories, writes the file, computes its hash,
// extracts metadata, and converts HEIC to JPEG preview.
// Returns metadata about the saved file or an error.
func SaveUploadedFile(src io.Reader, mediaType, filename, mimeType string) (*SavedFile, error) {
	// Generate storage path: photos/2026/01/filename.jpg
	now := time.Now()
	subDir := fmt.Sprintf("%ss/%d/%02d", mediaType, now.Year(), now.Month())
	safeFilename := sanitizeFilename(filename)
	relativePath := filepath.Join(subDir, safeFilename)
	fullPath := filepath.Join(getMediaBasePath(), relativePath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file and calculate hash simultaneously
	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)
	size, err := io.Copy(dst, tee)
	if err != nil {
		os.Remove(fullPath) // Clean up on failure
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	result := &SavedFile{
		RelativePath: relativePath,
		FullPath:     fullPath,
		Size:         size,
		SHA256:       hex.EncodeToString(hasher.Sum(nil)),
	}

	// Handle HEIC files - convert to JPEG preview
	if IsHEIC(mimeType) {
		previewPath, err := ConvertHEICtoJPEG(fullPath)
		if err != nil {
			// Log but don't fail - original file is still saved
			fmt.Printf("Warning: failed to convert HEIC to JPEG: %v\n", err)
		} else {
			// Calculate relative path for preview
			result.PreviewFullPath = previewPath
			result.PreviewRelativePath = strings.TrimPrefix(previewPath, getMediaBasePath()+"/")

			// Extract metadata from the JPEG preview (easier than HEIC)
			if meta, err := ExtractImageMetadata(previewPath); err == nil {
				result.Metadata = meta
			}
		}
	} else if strings.HasPrefix(mimeType, "image/") {
		// Extract metadata from JPEG/PNG directly
		if meta, err := ExtractImageMetadata(fullPath); err == nil {
			result.Metadata = meta
		}
	}

	return result, nil
}

// CleanupFile removes a file from disk (used for rollback on DB errors).
func CleanupFile(fullPath string) {
	os.Remove(fullPath)
}

// CleanupFiles removes multiple files from disk.
func CleanupFiles(paths ...string) {
	for _, path := range paths {
		if path != "" {
			os.Remove(path)
		}
	}
}

// sanitizeFilename removes dangerous characters from filenames.
func sanitizeFilename(name string) string {
	// Remove path separators and null bytes
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "\x00", "")

	// Remove or replace other problematic characters
	re := regexp.MustCompile(`[<>:"|?*]`)
	name = re.ReplaceAllString(name, "_")

	// Trim spaces and dots from edges
	name = strings.Trim(name, " .")

	if name == "" {
		name = "unnamed"
	}

	return name
}

// detectMediaType determines if a file is a photo or video based on MIME type.
func detectMediaType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "photo"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	default:
		return ""
	}
}

// getMediaBasePath returns the base path for media storage.
// Uses MEDIA_PATH env var or defaults to /mnt/storage/media (for Pi).
func getMediaBasePath() string {
	if path := os.Getenv("MEDIA_PATH"); path != "" {
		return path
	}
	return "/mnt/storage/media"
}
