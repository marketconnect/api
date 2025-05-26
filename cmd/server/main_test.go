package main

import (
	"testing"

	apiv1 "api/gen/api/v1"
)

func TestProductServer_NewProductServer(t *testing.T) {
	server := NewProductServer("http://localhost:8000/test")
	if server == nil {
		t.Fatal("NewProductServer returned nil")
	}
	if server.pythonAPIURL != "http://localhost:8000/test" {
		t.Errorf("Expected pythonAPIURL to be 'http://localhost:8000/test', got '%s'", server.pythonAPIURL)
	}
}

func TestProductRequest_Fields(t *testing.T) {
	req := &apiv1.ProductRequest{
		ProductTitle:       "Test Title",
		ProductDescription: "Test Description",
	}

	if req.ProductTitle != "Test Title" {
		t.Errorf("Expected ProductTitle to be 'Test Title', got '%s'", req.ProductTitle)
	}
	if req.ProductDescription != "Test Description" {
		t.Errorf("Expected ProductDescription to be 'Test Description', got '%s'", req.ProductDescription)
	}
}

func TestProductResponse_Fields(t *testing.T) {
	resp := &apiv1.ProductResponse{
		Title:           "Response Title",
		Description:     "Response Description",
		GenerateContent: true,
		Ozon:            true,
		Wb:              false,
	}

	if resp.Title != "Response Title" {
		t.Errorf("Expected Title to be 'Response Title', got '%s'", resp.Title)
	}
	if resp.Description != "Response Description" {
		t.Errorf("Expected Description to be 'Response Description', got '%s'", resp.Description)
	}
	if !resp.GenerateContent {
		t.Error("Expected GenerateContent to be true")
	}
	if !resp.Ozon {
		t.Error("Expected Ozon to be true")
	}
	if resp.Wb {
		t.Error("Expected Wb to be false")
	}
}
