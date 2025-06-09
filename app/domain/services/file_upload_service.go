package services

import (
	"api/app/domain/entities"
	"context"
	"fmt"
	"log"
	"mime"
	"path/filepath"
)

type fileStorageClient interface {
	UploadFiles(ctx context.Context, files []entities.FileUploadRequest) ([]entities.FileUploadResult, error)
}

type FileUploadService struct {
	fileStorageClient fileStorageClient
}

func NewFileUploadService(fileStorageClient fileStorageClient) *FileUploadService {
	return &FileUploadService{
		fileStorageClient: fileStorageClient,
	}
}

// UploadWBMediaFiles converts WB media files to upload requests and uploads them
func (fus *FileUploadService) UploadWBMediaFiles(ctx context.Context, wbFiles []*entities.WBClientMediaFile) ([]string, error) {
	if len(wbFiles) == 0 {
		return []string{}, nil
	}

	log.Printf("[FILE UPLOAD] Starting upload of %d files", len(wbFiles))

	// Convert WB files to upload requests
	uploadRequests := make([]entities.FileUploadRequest, len(wbFiles))
	for i, wbFile := range wbFiles {
		// Detect content type from filename
		contentType := mime.TypeByExtension(filepath.Ext(wbFile.Filename))
		if contentType == "" {
			contentType = "application/octet-stream" // Default fallback
		}

		uploadRequests[i] = entities.FileUploadRequest{
			Content:     wbFile.Content,
			Filename:    wbFile.Filename,
			ContentType: contentType,
		}

		log.Printf("[FILE UPLOAD] Prepared file %d: %s (content-type: %s, size: %d bytes)",
			i+1, wbFile.Filename, contentType, len(wbFile.Content))
	}

	// Upload files
	results, err := fus.fileStorageClient.UploadFiles(ctx, uploadRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to upload files: %w", err)
	}

	// Extract URLs and handle errors
	var urls []string
	var uploadErrors []string

	for i, result := range results {
		if result.Error != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("File %d (%s): %v", i+1, result.Filename, result.Error))
			log.Printf("[FILE UPLOAD] Error uploading file %s: %v", result.Filename, result.Error)
		} else {
			urls = append(urls, result.URL)
			log.Printf("[FILE UPLOAD] Successfully uploaded file %s -> %s", result.Filename, result.URL)
		}
	}

	if len(uploadErrors) > 0 {
		log.Printf("[FILE UPLOAD] Upload completed with %d errors out of %d files", len(uploadErrors), len(wbFiles))
		// Return partial success - URLs that worked
		if len(urls) > 0 {
			log.Printf("[FILE UPLOAD] Returning %d successful URLs despite errors", len(urls))
		}
	} else {
		log.Printf("[FILE UPLOAD] All %d files uploaded successfully", len(wbFiles))
	}

	return urls, nil
}
