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
		Title:           "Test Title",
		Description:     "Test Description",
		GenerateContent: true,
		Ozon:            true,
		Wb:              false,
	}

	if req.Title != "Test Title" {
		t.Errorf("Expected Title to be 'Test Title', got '%s'", req.Title)
	}
	if req.Description != "Test Description" {
		t.Errorf("Expected Description to be 'Test Description', got '%s'", req.Description)
	}
	if !req.GenerateContent {
		t.Error("Expected GenerateContent to be true")
	}
	if !req.Ozon {
		t.Error("Expected Ozon to be true")
	}
	if req.Wb {
		t.Error("Expected Wb to be false")
	}
}

func TestProductResponse_Fields(t *testing.T) {
	resp := &apiv1.ProductResponse{
		Title:       "Response Title",
		Description: "Response Description",
		Attributes:  map[string]string{"key": "value"},
		ParentId:    123,
		SubjectId:   456,
		TypeId:      789,
		RootId:      101,
		SubId:       112,
	}

	if resp.Title != "Response Title" {
		t.Errorf("Expected Title to be 'Response Title', got '%s'", resp.Title)
	}
	if resp.Description != "Response Description" {
		t.Errorf("Expected Description to be 'Response Description', got '%s'", resp.Description)
	}
	if resp.Attributes["key"] != "value" {
		t.Errorf("Expected Attributes[key] to be 'value', got '%s'", resp.Attributes["key"])
	}
	if resp.ParentId != 123 {
		t.Errorf("Expected ParentId to be 123, got %d", resp.ParentId)
	}
	if resp.SubjectId != 456 {
		t.Errorf("Expected SubjectId to be 456, got %d", resp.SubjectId)
	}
}
