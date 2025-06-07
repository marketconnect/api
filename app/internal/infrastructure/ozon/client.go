package ozon

import (
	"api/app/domain/entities"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const ozonAPIHost = "https://api-seller.ozon.ru"

// Client manages communication with the Ozon Seller API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Ozon API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

// ImportProductsV3 creates or updates products on Ozon.
// Corresponds to POST /v3/product/import
func (c *Client) ImportProductsV3(ctx context.Context, clientID, apiKey string, request entities.OzonProductImportRequest) (*entities.OzonProductImportResponse, error) {
	if clientID == "" {
		return nil, fmt.Errorf("ozon Client-Id is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("ozon Api-Key is required")
	}

	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Ozon product import request: %w", err)
	}

	importURL := fmt.Sprintf("%s/v3/product/import", ozonAPIHost)
	log.Printf("Importing products to Ozon: %s, Payload: %s", importURL, string(payloadBytes))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, importURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Ozon product import request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Client-Id", clientID)
	httpReq.Header.Set("Api-Key", apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ozon product import API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Ozon product import response body: %w", err)
	}

	log.Printf("Ozon product import API response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ozon product import API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var ozonResp entities.OzonProductImportResponse
	if err := json.Unmarshal(respBody, &ozonResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Ozon product import response: %w. Body: %s", err, string(respBody))
	}

	return &ozonResp, nil
}
