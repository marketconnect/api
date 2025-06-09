package file_storage

import (
	"api/app/domain/entities"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TemporaryFileStorage struct {
	uploadDir string        // Directory to store uploaded files
	baseURL   string        // Base URL to access files
	fileTTL   time.Duration // How long to keep files
}

func NewTemporaryFileStorage(uploadDir, baseURL string, fileTTL time.Duration) *TemporaryFileStorage {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Warning: Failed to create upload directory %s: %v", uploadDir, err)
	}

	tfs := &TemporaryFileStorage{
		uploadDir: uploadDir,
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		fileTTL:   fileTTL,
	}

	// Start background cleanup routine
	go tfs.startCleanupRoutine()

	return tfs
}

func (tfs *TemporaryFileStorage) UploadFiles(ctx context.Context, files []entities.FileUploadRequest) ([]entities.FileUploadResult, error) {
	results := make([]entities.FileUploadResult, len(files))

	for i, file := range files {
		// Generate unique filename
		uniqueFilename, err := tfs.generateUniqueFilename(file.Filename)
		if err != nil {
			results[i] = entities.FileUploadResult{
				Filename: file.Filename,
				Error:    fmt.Errorf("failed to generate unique filename: %w", err),
			}
			continue
		}

		// Full path for the file
		filePath := filepath.Join(tfs.uploadDir, uniqueFilename)

		// Write file to disk
		if err := os.WriteFile(filePath, file.Content, 0644); err != nil {
			results[i] = entities.FileUploadResult{
				Filename: file.Filename,
				Error:    fmt.Errorf("failed to write file to disk: %w", err),
			}
			continue
		}

		// Generate public URL
		publicURL := fmt.Sprintf("%s/%s", tfs.baseURL, uniqueFilename)

		results[i] = entities.FileUploadResult{
			URL:      publicURL,
			Filename: file.Filename,
			Error:    nil,
		}

		log.Printf("Successfully uploaded temporary file: %s -> %s (TTL: %v)", file.Filename, publicURL, tfs.fileTTL)
	}

	return results, nil
}

// generateUniqueFilename generates a unique filename using timestamp and random bytes
func (tfs *TemporaryFileStorage) generateUniqueFilename(originalFilename string) (string, error) {
	// Extract file extension
	ext := filepath.Ext(originalFilename)

	// Generate random bytes
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Create unique filename: timestamp_random_original
	timestamp := time.Now().Unix()
	randomStr := fmt.Sprintf("%x", randomBytes)

	// Clean original filename (remove extension and special chars)
	baseName := strings.TrimSuffix(originalFilename, ext)
	baseName = strings.ReplaceAll(baseName, " ", "_")
	baseName = strings.ReplaceAll(baseName, "/", "_")

	uniqueFilename := fmt.Sprintf("%d_%s_%s%s", timestamp, randomStr, baseName, ext)

	return uniqueFilename, nil
}

// startCleanupRoutine runs a background process to clean up old files
func (tfs *TemporaryFileStorage) startCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Minute) // Check every 30 minutes
	defer ticker.Stop()

	log.Printf("Started file cleanup routine: checking every 30 minutes, TTL: %v", tfs.fileTTL)

	for {
		select {
		case <-ticker.C:
			tfs.cleanupOldFiles()
		}
	}
}

// cleanupOldFiles removes files older than TTL
func (tfs *TemporaryFileStorage) cleanupOldFiles() {
	log.Printf("[CLEANUP] Starting cleanup of files older than %v", tfs.fileTTL)

	cutoffTime := time.Now().Add(-tfs.fileTTL)
	deletedCount := 0
	errorCount := 0

	err := filepath.Walk(tfs.uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is older than TTL
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(path); err != nil {
				log.Printf("[CLEANUP] Failed to delete old file %s: %v", path, err)
				errorCount++
			} else {
				log.Printf("[CLEANUP] Deleted old file: %s (age: %v)", path, time.Since(info.ModTime()))
				deletedCount++
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("[CLEANUP] Error during cleanup: %v", err)
	} else {
		log.Printf("[CLEANUP] Cleanup completed: deleted %d files, %d errors", deletedCount, errorCount)
	}
}
