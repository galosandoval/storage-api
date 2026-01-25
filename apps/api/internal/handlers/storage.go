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
	Metadata            *ImageMetadata
}

// SaveUploadedFile saves a file to the media storage.
// Creates directories, writes the file, computes hash, extracts metadata, and generates previews.
func SaveUploadedFile(src io.Reader, mediaType, filename, mimeType string) (*SavedFile, error) {
	relativePath, fullPath := generateStoragePath(mediaType, filename)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	size, hash, err := saveFileWithHash(src, fullPath)
	if err != nil {
		return nil, err
	}

	result := &SavedFile{
		RelativePath: relativePath,
		FullPath:     fullPath,
		Size:         size,
		SHA256:       hash,
	}

	processMediaFile(result, fullPath, relativePath, mimeType)

	return result, nil
}

func generateStoragePath(mediaType, filename string) (relativePath, fullPath string) {
	now := time.Now()
	subDir := fmt.Sprintf("%ss/%d/%02d", mediaType, now.Year(), now.Month())
	safeFilename := sanitizeFilename(filename)
	relativePath = filepath.Join(subDir, safeFilename)
	fullPath = filepath.Join(getMediaBasePath(), relativePath)
	return relativePath, fullPath
}

func saveFileWithHash(src io.Reader, fullPath string) (int64, string, error) {
	dst, err := os.Create(fullPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)
	size, err := io.Copy(dst, tee)
	if err != nil {
		os.Remove(fullPath)
		return 0, "", fmt.Errorf("failed to save file: %w", err)
	}

	return size, hex.EncodeToString(hasher.Sum(nil)), nil
}

func processMediaFile(result *SavedFile, fullPath, relativePath, mimeType string) {
	switch {
	case IsHEIC(mimeType):
		processHEICFile(result, fullPath, relativePath)
	case strings.HasPrefix(mimeType, "image/"):
		processImageFile(result, fullPath, relativePath)
	case strings.HasPrefix(mimeType, "video/"):
		processVideoFile(result, fullPath, relativePath)
	}
}

func processHEICFile(result *SavedFile, fullPath, relativePath string) {
	previewPath, err := ConvertHEICtoJPEG(fullPath)
	if err != nil {
		fmt.Printf("Warning: failed to convert HEIC to JPEG: %v\n", err)
		return
	}

	result.PreviewFullPath = previewPath
	result.PreviewRelativePath = strings.TrimPrefix(previewPath, getMediaBasePath()+"/")

	if meta, err := ExtractImageMetadata(previewPath); err == nil {
		result.Metadata = meta
	}

	if thumbFull, thumbRel, err := GenerateImageThumbnail(previewPath, relativePath); err == nil {
		result.ThumbnailFullPath = thumbFull
		result.ThumbnailRelPath = thumbRel
	} else {
		fmt.Printf("Warning: failed to generate thumbnail for HEIC: %v\n", err)
	}
}

func processImageFile(result *SavedFile, fullPath, relativePath string) {
	if meta, err := ExtractImageMetadata(fullPath); err == nil {
		result.Metadata = meta
	}

	if thumbFull, thumbRel, err := GenerateImageThumbnail(fullPath, relativePath); err == nil {
		result.ThumbnailFullPath = thumbFull
		result.ThumbnailRelPath = thumbRel
	} else {
		fmt.Printf("Warning: failed to generate thumbnail: %v\n", err)
	}
}

func processVideoFile(result *SavedFile, fullPath, relativePath string) {
	if thumbFull, thumbRel, err := GenerateVideoThumbnail(fullPath, relativePath); err == nil {
		result.ThumbnailFullPath = thumbFull
		result.ThumbnailRelPath = thumbRel
	} else {
		fmt.Printf("Warning: failed to generate video thumbnail: %v\n", err)
	}
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

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "\x00", "")

	re := regexp.MustCompile(`[<>:"|?*]`)
	name = re.ReplaceAllString(name, "_")
	name = strings.Trim(name, " .")

	if name == "" {
		name = "unnamed"
	}
	return name
}

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

func getMediaBasePath() string {
	if path := os.Getenv("MEDIA_PATH"); path != "" {
		return path
	}
	return "/mnt/storage/media"
}

func getThumbnailPath(relativePath string) (fullPath, relPath string) {
	ext := filepath.Ext(relativePath)
	basePath := strings.TrimSuffix(relativePath, ext) + ".jpg"
	relPath = filepath.Join(".thumbs", basePath)
	fullPath = filepath.Join(getMediaBasePath(), relPath)
	return fullPath, relPath
}

// GenerateImageThumbnail creates a 300px thumbnail for an image file.
func GenerateImageThumbnail(srcPath, originalRelPath string) (fullPath, relPath string, err error) {
	src, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return "", "", fmt.Errorf("failed to open image: %w", err)
	}

	thumb := imaging.Fit(src, ThumbnailSize, ThumbnailSize, imaging.Lanczos)
	fullPath, relPath = getThumbnailPath(originalRelPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	if err := imaging.Save(thumb, fullPath, imaging.JPEGQuality(80)); err != nil {
		return "", "", fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return fullPath, relPath, nil
}

// GenerateVideoThumbnail extracts a frame from a video and creates a thumbnail.
func GenerateVideoThumbnail(srcPath, originalRelPath string) (fullPath, relPath string, err error) {
	fullPath, relPath = getThumbnailPath(originalRelPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	tempFrame := fullPath + ".temp.jpg"
	defer os.Remove(tempFrame)

	if err := extractVideoFrame(srcPath, tempFrame); err != nil {
		return "", "", err
	}

	if err := resizeFrameToThumbnail(tempFrame, fullPath); err != nil {
		return "", "", err
	}

	return fullPath, relPath, nil
}

func extractVideoFrame(srcPath, tempFrame string) error {
	// Try extracting frame at 1 second first
	cmd := exec.Command("ffmpeg",
		"-y", "-ss", "1", "-i", srcPath,
		"-vframes", "1", "-q:v", "2", tempFrame,
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// Fall back to first frame for short videos
	cmd = exec.Command("ffmpeg",
		"-y", "-i", srcPath,
		"-vframes", "1", "-q:v", "2", tempFrame,
	)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %v, output: %s", err, string(output))
	}
	return nil
}

func resizeFrameToThumbnail(tempFrame, destPath string) error {
	img, err := imaging.Open(tempFrame)
	if err != nil {
		return fmt.Errorf("failed to open extracted frame: %w", err)
	}

	thumb := imaging.Fit(img, ThumbnailSize, ThumbnailSize, imaging.Lanczos)
	if err := imaging.Save(thumb, destPath, imaging.JPEGQuality(80)); err != nil {
		return fmt.Errorf("failed to save video thumbnail: %w", err)
	}
	return nil
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
