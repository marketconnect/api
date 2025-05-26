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
	ParentID    *int              `json:"parent_id"`
	SubjectID   *int              `json:"subject_id"`
	TypeID      *int              `json:"type_id"`
	RootID      *int              `json:"root_id"`
	SubID       *int              `json:"sub_id"`
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
	log.Printf("Received request: title=%s, description=%s", req.Msg.ProductTitle, req.Msg.ProductDescription)

	// Prepare request for Python API
	pythonReq := PythonAPIRequest{
		ProductTitle:       req.Msg.ProductTitle,
		ProductDescription: req.Msg.ProductDescription,
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

	// Create ConnectRPC response with the required fields
	response := &apiv1.ProductResponse{
		Title:           pythonResp.Title,
		Description:     pythonResp.Description,
		GenerateContent: true, // Set based on your business logic
		Ozon:            pythonResp.TypeID != nil && pythonResp.RootID != nil && pythonResp.SubID != nil,
		Wb:              pythonResp.SubjectID != nil && pythonResp.ParentID != nil,
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
