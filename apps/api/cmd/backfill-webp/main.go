package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"storage-api/internal/config"
	"storage-api/internal/db"
	"storage-api/internal/handlers"
	"storage-api/internal/models"
)

func main() {
	cfg := config.Load()

	gormDB, err := db.New(cfg.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(gormDB)

	ctx := context.Background()

	// Find all photos without web_path
	var photos []models.MediaItem
	result := gormDB.WithContext(ctx).
		Where("type = ?", "photo").
		Where("web_path IS NULL OR web_path = ''").
		Find(&photos)

	if result.Error != nil {
		log.Fatalf("Failed to query photos: %v", result.Error)
	}

	log.Printf("Found %d photos to backfill", len(photos))

	mediaPath := os.Getenv("MEDIA_PATH")
	if mediaPath == "" {
		mediaPath = "/mnt/storage/media"
	}

	successCount := 0
	errorCount := 0

	for i, photo := range photos {
		log.Printf("[%d/%d] Processing %s", i+1, len(photos), photo.Path)

		// Determine source path (use preview for HEIC, otherwise original)
		srcPath := photo.Path
		if photo.PreviewPath != "" {
			srcPath = photo.PreviewPath
		}
		fullSrcPath := fmt.Sprintf("%s/%s", mediaPath, srcPath)

		// Generate WebP
		webFull, webRel, err := handlers.GenerateWebOptimizedImage(fullSrcPath, photo.Path)
		if err != nil {
			log.Printf("  ERROR: %v", err)
			errorCount++
			continue
		}

		// Update database
		updateResult := gormDB.WithContext(ctx).
			Model(&models.MediaItem{}).
			Where("id = ?", photo.ID).
			Update("web_path", webRel)

		if updateResult.Error != nil {
			log.Printf("  ERROR updating database: %v", updateResult.Error)
			os.Remove(webFull) // Clean up generated file
			errorCount++
			continue
		}

		log.Printf("  OK: %s", webRel)
		successCount++
	}

	log.Printf("Backfill complete: %d success, %d errors", successCount, errorCount)
}
