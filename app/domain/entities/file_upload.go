package entities

// FileUploadRequest represents a request to upload a file
type FileUploadRequest struct {
	Content     []byte
	Filename    string
	ContentType string
}

// FileUploadResult represents the result of a file upload operation
type FileUploadResult struct {
	URL      string // Public URL of the uploaded file
	Filename string // Original filename
	Error    error  // Error if upload failed
}
