package wb

import (
	"api/app/domain/entities"
	"context"
	"strings"
	"testing"
)

func TestNewWBClient(t *testing.T) {
	client := NewWBClient()
	if client == nil {
		t.Fatal("NewWBClient returned nil")
	}
}

func TestWBClient_UploadWBCard(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-api-key"
	payload := entities.WBCardUploadPayload{
		entities.WBCardRequestItem{SubjectID: 123, Variants: []entities.WBVariant{{VendorCode: "VC001"}}},
	}

	t.Run("API key missing", func(t *testing.T) {
		client := NewWBClient()
		_, err := client.UploadWBCard(ctx, payload, "") // Empty API key
		if err == nil {
			t.Fatal("Expected an error for missing API key, got nil")
		}
		if !strings.Contains(err.Error(), "API key is required") {
			t.Errorf("Expected error message about API key, got: %s", err.Error())
		}
	})

	t.Run("With valid API key", func(t *testing.T) {
		client := NewWBClient()
		// We expect an error because we're not actually hitting a real WB server
		_, err := client.UploadWBCard(ctx, payload, apiKey)
		if err == nil {
			t.Error("Expected an error when calling real API endpoint, got nil")
		}
	})
}

func TestWBClient_UploadMediaFiles(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-api-key"
	nmID := "12345"
	files := []entities.WBClientMediaFile{
		{Filename: "photo1.jpg", Content: []byte("jpeg_data_1"), PhotoNumber: 1},
		{Filename: "photo2.png", Content: []byte("png_data_2"), PhotoNumber: 2},
	}

	t.Run("API key missing", func(t *testing.T) {
		client := NewWBClient()
		_, err := client.UploadMediaFiles(ctx, "", nmID, files) // Empty API key
		if err == nil {
			t.Fatal("Expected an error for missing API key, got nil")
		}
		if !strings.Contains(err.Error(), "API key is required") {
			t.Errorf("Expected error message about API key, got: %s", err.Error())
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel context immediately

		client := NewWBClient()
		results, err := client.UploadMediaFiles(cancelledCtx, apiKey, nmID, files)
		if err == nil {
			t.Fatal("Expected context cancellation error, got nil")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
		// With cancelled context, results might be empty or partial
		if len(results) > len(files) {
			t.Errorf("Results length should not exceed files length")
		}
	})

	t.Run("With valid parameters", func(t *testing.T) {
		client := NewWBClient()
		// UploadMediaFiles processes each file individually and returns results with errors
		// rather than returning a Go error when HTTP calls fail
		results, err := client.UploadMediaFiles(ctx, apiKey, nmID, files)
		if err != nil {
			t.Fatalf("UploadMediaFiles returned unexpected error: %v", err)
		}
		// We expect results for each file, but they should contain individual errors
		// due to unauthorized API calls (since we're using test credentials)
		if len(results) != len(files) {
			t.Errorf("Expected %d results, got %d", len(files), len(results))
		}
		// Each result should have an error since we're not using valid credentials
		for i, result := range results {
			if result.Error == nil {
				t.Errorf("Expected error in result %d, got nil", i)
			}
		}
	})
}

func TestWBClient_SaveMediaByLinks(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-api-key"
	payload := entities.WBSaveMediaPayload{NmID: 123, Data: []string{"http://example.com/img.jpg"}}

	client := NewWBClient()
	// We expect an error because we're not actually hitting a real WB server
	_, err := client.SaveMediaByLinks(ctx, apiKey, payload)
	if err == nil {
		t.Error("Expected an error when calling real API endpoint, got nil")
	}
}

func TestWBClient_GetCardList(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-api-key"
	payload := entities.WBGetCardListRequest{Settings: entities.WBGetCardListRequestSettings{
		Cursor: entities.WBGetCardListRequestCursor{Limit: 10},
	}}

	client := NewWBClient()
	// We expect an error because we're not actually hitting a real WB server
	_, err := client.GetCardList(ctx, apiKey, payload)
	if err == nil {
		t.Error("Expected an error when calling real API endpoint, got nil")
	}
}
