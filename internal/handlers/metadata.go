package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// ImageMetadata contains extracted EXIF and image data
type ImageMetadata struct {
	Width        int
	Height       int
	TakenAt      *time.Time
	CameraMake   string
	CameraModel  string
	Latitude     *float64
	Longitude    *float64
	Orientation  int
	ISO          int
	FNumber      *float64
	ExposureTime string
	FocalLength  *float64
}

// ExtractImageMetadata extracts EXIF metadata from a JPEG or PNG file
func ExtractImageMetadata(filePath string) (*ImageMetadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		// No EXIF data is not an error, just return empty metadata
		return &ImageMetadata{}, nil
	}

	meta := &ImageMetadata{}

	// Extract dimensions
	if w, err := x.Get(exif.PixelXDimension); err == nil {
		if val, err := w.Int(0); err == nil {
			meta.Width = val
		}
	}
	if h, err := x.Get(exif.PixelYDimension); err == nil {
		if val, err := h.Int(0); err == nil {
			meta.Height = val
		}
	}

	// Extract date taken
	if dt, err := x.DateTime(); err == nil {
		meta.TakenAt = &dt
	}

	// Extract camera info
	if make, err := x.Get(exif.Make); err == nil {
		meta.CameraMake = strings.Trim(strings.TrimSpace(make.String()), "\"")
	}
	if model, err := x.Get(exif.Model); err == nil {
		meta.CameraModel = strings.Trim(strings.TrimSpace(model.String()), "\"")
	}

	// Extract GPS coordinates
	if lat, lon, err := x.LatLong(); err == nil {
		meta.Latitude = &lat
		meta.Longitude = &lon
	}

	// Extract orientation
	if orient, err := x.Get(exif.Orientation); err == nil {
		if val, err := orient.Int(0); err == nil {
			meta.Orientation = val
		}
	}

	// Extract ISO
	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if val, err := iso.Int(0); err == nil {
			meta.ISO = val
		}
	}

	// Extract aperture (FNumber)
	if fn, err := x.Get(exif.FNumber); err == nil {
		if num, denom, err := fn.Rat2(0); err == nil && denom != 0 {
			fval := float64(num) / float64(denom)
			meta.FNumber = &fval
		}
	}

	// Extract exposure time
	if et, err := x.Get(exif.ExposureTime); err == nil {
		if num, denom, err := et.Rat2(0); err == nil {
			if denom == 1 {
				meta.ExposureTime = fmt.Sprintf("%d", num)
			} else {
				meta.ExposureTime = fmt.Sprintf("%d/%d", num, denom)
			}
		}
	}

	// Extract focal length
	if fl, err := x.Get(exif.FocalLength); err == nil {
		if num, denom, err := fl.Rat2(0); err == nil && denom != 0 {
			fval := float64(num) / float64(denom)
			meta.FocalLength = &fval
		}
	}

	return meta, nil
}

// ConvertHEICtoJPEG converts a HEIC file to JPEG using heif-convert
// Returns the path to the created JPEG file
func ConvertHEICtoJPEG(heicPath string) (string, error) {
	// Generate output path: same directory, .jpg extension
	dir := filepath.Dir(heicPath)
	base := filepath.Base(heicPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	jpegPath := filepath.Join(dir, nameWithoutExt+".jpg")

	// Run heif-convert command
	cmd := exec.Command("heif-convert", "-q", "85", heicPath, jpegPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("heif-convert failed: %v, output: %s", err, string(output))
	}

	// Verify the file was created
	if _, err := os.Stat(jpegPath); os.IsNotExist(err) {
		return "", fmt.Errorf("heif-convert did not create output file")
	}

	return jpegPath, nil
}

// IsHEIC checks if the MIME type indicates a HEIC file
func IsHEIC(mimeType string) bool {
	mimeType = strings.ToLower(mimeType)
	return mimeType == "image/heic" || mimeType == "image/heif"
}

// NeedsConversion checks if a file type needs conversion for web display
func NeedsConversion(mimeType string) bool {
	return IsHEIC(mimeType)
}
