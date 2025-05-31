package card_craft_ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api/app/domain/entities"
)

func TestNewCardCraftAiClient(t *testing.T) {
	apiURL := "http://localhost:8000"
	client := NewCardCraftAiClient(apiURL)

	if client == nil {
		t.Fatal("NewCardCraftAiClient returned nil")
	}
	if client.cardCraftAiAPIURL != apiURL {
		t.Errorf("expected cardCraftAiAPIURL %s, got %s", apiURL, client.cardCraftAiAPIURL)
	}
	expectedSessionURL := fmt.Sprintf("%s/v1/sessions", apiURL)
	if client.getSessionURL != expectedSessionURL {
		t.Errorf("expected getSessionURL %s, got %s", expectedSessionURL, client.getSessionURL)
	}
}

func TestCardCraftAiClient_GetSessionID(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful session ID retrieval", func(t *testing.T) {
		expectedSessionID := "test-session-123"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if !strings.HasSuffix(r.URL.Path, "/v1/sessions") {
				t.Errorf("Expected URL path to end with /v1/sessions, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(entities.CardCraftAiSessionResponse{SessionID: expectedSessionID})
		}))
		defer server.Close()

		client := NewCardCraftAiClient(server.URL)
		sessionID, err := client.GetSessionID(ctx)

		if err != nil {
			t.Fatalf("GetSessionID failed: %v", err)
		}
		if sessionID != expectedSessionID {
			t.Errorf("Expected session ID %s, got %s", expectedSessionID, sessionID)
		}
	})

	t.Run("HTTP error from server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		}))
		defer server.Close()

		client := NewCardCraftAiClient(server.URL)
		_, err := client.GetSessionID(ctx)

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "status 500") {
			t.Errorf("Expected error message to contain 'status 500', got '%s'", err.Error())
		}
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("this is not json"))
		}))
		defer server.Close()

		client := NewCardCraftAiClient(server.URL)
		_, err := client.GetSessionID(ctx)

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal session response") {
			t.Errorf("Expected error message about unmarshalling, got '%s'", err.Error())
		}
	})
}

func TestCardCraftAiClient_GetCardContent(t *testing.T) {
	ctx := context.Background()
	sessionID := "test-session-id"
	apiRequest := entities.CardCraftAiAPIRequest{ProductTitle: "Test Product"}

	t.Run("Successful card content retrieval", func(t *testing.T) {
		expectedResponse := &entities.CardCraftAiAPIResponse{
			Title: "AI Generated Title", Description: "AI Generated Description",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			expectedPath := fmt.Sprintf("/v1/sessions/%s/product_card_comprehensive", sessionID)
			if r.URL.Path != expectedPath {
				t.Errorf("Expected URL path %s, got %s", expectedPath, r.URL.Path)
			}

			var receivedReq entities.CardCraftAiAPIRequest
			if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if receivedReq.ProductTitle != apiRequest.ProductTitle {
				t.Errorf("Received product title '%s', expected '%s'", receivedReq.ProductTitle, apiRequest.ProductTitle)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client := NewCardCraftAiClient(server.URL)
		response, err := client.GetCardContent(ctx, sessionID, apiRequest)

		if err != nil {
			t.Fatalf("GetCardContent failed: %v", err)
		}
		if response.Title != expectedResponse.Title {
			t.Errorf("Expected title '%s', got '%s'", expectedResponse.Title, response.Title)
		}
	})

	t.Run("HTTP error from server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("access denied"))
		}))
		defer server.Close()

		client := NewCardCraftAiClient(server.URL)
		_, err := client.GetCardContent(ctx, sessionID, apiRequest)

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "status 403") {
			t.Errorf("Expected error message to contain 'status 403', got '%s'", err.Error())
		}
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not valid json content"))
		}))
		defer server.Close()

		client := NewCardCraftAiClient(server.URL)
		_, err := client.GetCardContent(ctx, sessionID, apiRequest)

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal CardCraftAI API response") {
			t.Errorf("Expected error message about unmarshalling, got '%s'", err.Error())
		}
	})

	// Test for request marshalling error (harder to trigger without specific broken input type)
	// For example, if CardCraftAiAPIRequest contained a channel or function.
	// The current struct is simple and should always marshal.
	// To achieve coverage for `failed to marshal request`, one might pass a type that json.Marshal fails on.
	// This is less about the client's logic and more about `json.Marshal` behavior.
	// For now, this path is considered low-priority for direct testing unless the request struct becomes complex.
}
