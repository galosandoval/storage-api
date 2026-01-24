package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

// ThumbnailSize is the maximum dimension for generated thumbnails
const ThumbnailSize = 300

// SavedFile contains metadata about a successfully saved file.
type SavedFile struct {
	RelativePath        string
	FullPath            string
	Size                int64
	SHA256              string
	PreviewRelativePath string // Path to JPEG preview (for HEIC files)
	PreviewFullPath     string // Full path to preview
	ThumbnailRelPath    string // Path to thumbnail (relative)
	ThumbnailFullPath   string // Full path to thumbnail

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

			// Generate thumbnail from the JPEG preview
			if thumbFull, thumbRel, err := GenerateImageThumbnail(previewPath, relativePath); err == nil {
				result.ThumbnailFullPath = thumbFull
				result.ThumbnailRelPath = thumbRel
			} else {
				fmt.Printf("Warning: failed to generate thumbnail for HEIC: %v\n", err)
			}
		}
	} else if strings.HasPrefix(mimeType, "image/") {
		// Extract metadata from JPEG/PNG directly
		if meta, err := ExtractImageMetadata(fullPath); err == nil {
			result.Metadata = meta
		}

		// Generate thumbnail for image
		if thumbFull, thumbRel, err := GenerateImageThumbnail(fullPath, relativePath); err == nil {
			result.ThumbnailFullPath = thumbFull
			result.ThumbnailRelPath = thumbRel
		} else {
			fmt.Printf("Warning: failed to generate thumbnail: %v\n", err)
		}
	} else if strings.HasPrefix(mimeType, "video/") {
		// Generate thumbnail from video frame
		if thumbFull, thumbRel, err := GenerateVideoThumbnail(fullPath, relativePath); err == nil {
			result.ThumbnailFullPath = thumbFull
			result.ThumbnailRelPath = thumbRel
		} else {
			fmt.Printf("Warning: failed to generate video thumbnail: %v\n", err)
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

// getThumbnailPath generates the thumbnail path for a given relative path.
// Thumbnails are stored in .thumbs/ directory mirroring the original structure.
func getThumbnailPath(relativePath string) (fullPath, relPath string) {
	// Always use .jpg extension for thumbnails
	ext := filepath.Ext(relativePath)
	basePath := strings.TrimSuffix(relativePath, ext) + ".jpg"
	relPath = filepath.Join(".thumbs", basePath)
	fullPath = filepath.Join(getMediaBasePath(), relPath)
	return fullPath, relPath
}

// GenerateImageThumbnail creates a 300px thumbnail for an image file.
// Returns the full path and relative path to the thumbnail.
func GenerateImageThumbnail(srcPath, originalRelPath string) (fullPath, relPath string, err error) {
	// Open source image
	src, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return "", "", fmt.Errorf("failed to open image: %w", err)
	}

	// Resize to fit in ThumbnailSize x ThumbnailSize, maintaining aspect ratio
	thumb := imaging.Fit(src, ThumbnailSize, ThumbnailSize, imaging.Lanczos)

	// Get thumbnail paths
	fullPath, relPath = getThumbnailPath(originalRelPath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// Save thumbnail as JPEG with quality 80
	if err := imaging.Save(thumb, fullPath, imaging.JPEGQuality(80)); err != nil {
		return "", "", fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return fullPath, relPath, nil
}

// GenerateVideoThumbnail extracts a frame from a video and creates a thumbnail.
// Uses ffmpeg to extract a frame at 1 second (or first frame if video is shorter).
func GenerateVideoThumbnail(srcPath, originalRelPath string) (fullPath, relPath string, err error) {
	// Get thumbnail paths
	fullPath, relPath = getThumbnailPath(originalRelPath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// Create a temporary file for the extracted frame
	tempFrame := fullPath + ".temp.jpg"
	defer os.Remove(tempFrame) // Clean up temp file

	// Use ffmpeg to extract a frame at 1 second
	// -ss 1: seek to 1 second
	// -vframes 1: extract 1 frame
	// -q:v 2: high quality JPEG
	cmd := exec.Command("ffmpeg",
		"-y",             // Overwrite output
		"-ss", "1",       // Seek to 1 second
		"-i", srcPath,    // Input file
		"-vframes", "1",  // Extract 1 frame
		"-q:v", "2",      // High quality
		tempFrame,        // Output to temp file
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try extracting first frame if seeking fails (for very short videos)
		cmd = exec.Command("ffmpeg",
			"-y",
			"-i", srcPath,
			"-vframes", "1",
			"-q:v", "2",
			tempFrame,
		)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", "", fmt.Errorf("ffmpeg failed: %v, output: %s", err, string(output))
		}
	}

	// Resize the extracted frame to thumbnail size
	img, err := imaging.Open(tempFrame)
	if err != nil {
		return "", "", fmt.Errorf("failed to open extracted frame: %w", err)
	}

	thumb := imaging.Fit(img, ThumbnailSize, ThumbnailSize, imaging.Lanczos)
	if err := imaging.Save(thumb, fullPath, imaging.JPEGQuality(80)); err != nil {
		return "", "", fmt.Errorf("failed to save video thumbnail: %w", err)
	}

	return fullPath, relPath, nil
}

// CleanupThumbnail removes a thumbnail file from disk.
func CleanupThumbnail(relativePath string) {
	if relativePath == "" {
		return
	}
	fullPath := filepath.Join(getMediaBasePath(), relativePath)
	os.Remove(fullPath)
}

// Ensure imaging library types are recognized
var _ image.Image
