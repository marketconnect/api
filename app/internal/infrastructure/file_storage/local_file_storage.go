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

type LocalFileStorage struct {
	uploadDir string // Directory to store uploaded files
	baseURL   string // Base URL to access files (e.g., "http://example.com/uploads")
}

func NewLocalFileStorage(uploadDir, baseURL string) *LocalFileStorage {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Warning: Failed to create upload directory %s: %v", uploadDir, err)
	}

	return &LocalFileStorage{
		uploadDir: uploadDir,
		baseURL:   strings.TrimSuffix(baseURL, "/"), // Remove trailing slash
	}
}

func (lfs *LocalFileStorage) UploadFiles(ctx context.Context, files []entities.FileUploadRequest) ([]entities.FileUploadResult, error) {
	results := make([]entities.FileUploadResult, len(files))

	for i, file := range files {
		// Generate unique filename
		uniqueFilename, err := lfs.generateUniqueFilename(file.Filename)
		if err != nil {
			results[i] = entities.FileUploadResult{
				Filename: file.Filename,
				Error:    fmt.Errorf("failed to generate unique filename: %w", err),
			}
			continue
		}

		// Full path for the file
		filePath := filepath.Join(lfs.uploadDir, uniqueFilename)

		// Write file to disk
		if err := os.WriteFile(filePath, file.Content, 0644); err != nil {
			results[i] = entities.FileUploadResult{
				Filename: file.Filename,
				Error:    fmt.Errorf("failed to write file to disk: %w", err),
			}
			continue
		}

		// Generate public URL
		publicURL := fmt.Sprintf("%s/%s", lfs.baseURL, uniqueFilename)

		results[i] = entities.FileUploadResult{
			URL:      publicURL,
			Filename: file.Filename,
			Error:    nil,
		}

		log.Printf("Successfully uploaded file: %s -> %s", file.Filename, publicURL)
	}

	return results, nil
}

// generateUniqueFilename generates a unique filename using timestamp and random bytes
func (lfs *LocalFileStorage) generateUniqueFilename(originalFilename string) (string, error) {
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
