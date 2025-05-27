package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	apiv1 "api/gen/api/v1"
	"api/gen/api/v1/apiv1connect"
)

// PythonAPIRequest represents the request structure for the Python API
type PythonAPIRequest struct {
	ProductTitle       string `json:"product_title"`
	ProductDescription string `json:"product_description"`
}

// PythonAPIResponse represents the response structure from the Python API
type PythonAPIResponse struct {
	Title       string            `json:"title"`
	Attributes  map[string]string `json:"attributes"`
	Description string            `json:"description"`
	ParentID    *int32            `json:"parent_id"`
	SubjectID   *int32            `json:"subject_id"`
	TypeID      *int32            `json:"type_id"`
	RootID      *int32            `json:"root_id"`
	SubID       *int32            `json:"sub_id"`
	Keywords    []string          `json:"keywords"`
}

// ProductServer implements the ProductService
type ProductServer struct {
	pythonAPIURL string
	httpClient   *http.Client
}

// NewProductServer creates a new ProductServer instance
func NewProductServer(pythonAPIURL string) *ProductServer {
	return &ProductServer{
		pythonAPIURL: pythonAPIURL,
		httpClient:   &http.Client{},
	}
}

// GetProductCard implements the ProductService.GetProductCard method
func (s *ProductServer) GetProductCard(
	ctx context.Context,
	req *connect.Request[apiv1.ProductRequest],
) (*connect.Response[apiv1.ProductResponse], error) {
	log.Printf("Received request: title=%s, description=%s, generate_content=%t, ozon=%t, wb=%t",
		req.Msg.Title, req.Msg.Description, req.Msg.GenerateContent, req.Msg.Ozon, req.Msg.Wb)

	// Prepare request for Python API using the title and description from input
	pythonReq := PythonAPIRequest{
		ProductTitle:       req.Msg.Title,
		ProductDescription: req.Msg.Description,
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(pythonReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to marshal request: %w", err))
	}

	// Make request to Python API
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.pythonAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create HTTP request: %w", err))
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to call Python API: %w", err))
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read response body: %w", err))
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Python API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	// Parse Python API response
	var pythonResp PythonAPIResponse
	if err := json.Unmarshal(respBody, &pythonResp); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to unmarshal Python API response: %w", err))
	}

	log.Printf("Python API response: title=%s, description=%s", pythonResp.Title, pythonResp.Description)

	// Create ConnectRPC response with the comprehensive data from Python API
	response := &apiv1.ProductResponse{
		Title:       pythonResp.Title,
		Attributes:  pythonResp.Attributes,
		Description: pythonResp.Description,
		Keywords:    pythonResp.Keywords,
	}

	// Handle nullable int32 fields
	if pythonResp.ParentID != nil {
		response.ParentId = *pythonResp.ParentID
	}
	if pythonResp.SubjectID != nil {
		response.SubjectId = *pythonResp.SubjectID
	}
	if pythonResp.TypeID != nil {
		response.TypeId = *pythonResp.TypeID
	}
	if pythonResp.RootID != nil {
		response.RootId = *pythonResp.RootID
	}
	if pythonResp.SubID != nil {
		response.SubId = *pythonResp.SubID
	}

	// Create and return the Connect response
	connectResp := connect.NewResponse(response)
	connectResp.Header().Set("Content-Type", "application/json")

	return connectResp, nil
}

func main() {
	// Get Python API URL from environment variable or use default
	pythonAPIURL := os.Getenv("PYTHON_API_URL")
	if pythonAPIURL == "" {
		pythonAPIURL = "http://localhost:8000/v1/product_card_comprehensive"
	}

	// Create the product server
	productServer := NewProductServer(pythonAPIURL)

	// Create HTTP mux and register the service
	mux := http.NewServeMux()
	path, handler := apiv1connect.NewProductServiceHandler(productServer)
	mux.Handle(path, handler)

	// Add a health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Starting ConnectRPC server on %s", addr)
	log.Printf("Python API URL: %s", pythonAPIURL)

	// Start the server with HTTP/2 support
	err := http.ListenAndServe(
		addr,
		// Use h2c so we can serve HTTP/2 without TLS
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
