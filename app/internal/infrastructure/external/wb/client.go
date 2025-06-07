package wb

import (
	"api/app/domain/entities"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

// wildberriesAPIHost is the base URL for Wildberries content API.
const wildberriesAPIHost = "https://content-api.wildberries.ru"

type WBClient struct {
	// apiKey string // API key is now passed directly to UploadWBCard
}

func NewWBClient() *WBClient {
	return &WBClient{
		// apiKey: apiKey, // No longer stored in client
	}
}

func (c *WBClient) UploadWBCard(ctx context.Context, wbPayload entities.WBCardUploadPayload, apiKey string) (*entities.WBCardUploadResponse, error) {
	payloadBytes, err := json.Marshal(wbPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Wildberries card payload: %w", err)
	}

	uploadURL := fmt.Sprintf("%s/content/v2/cards/upload", wildberriesAPIHost)
	log.Printf("Uploading card to Wildberries: %s, Payload: %s", uploadURL, string(payloadBytes))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Wildberries card upload request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", apiKey)
	} else {
		// This case should ideally be prevented by the use case layer.
		// If an empty key is passed, WB API will likely reject it.
		return nil, fmt.Errorf("wildberries API key is required for uploading card")
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Wildberries card upload API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Wildberries card upload response body: %w", err)
	}

	log.Printf("Wildberries card upload API response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK { // WB API returns 200 for successful async queuing
		return nil, fmt.Errorf("wildberries card upload API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var wbResp entities.WBCardUploadResponse
	if err := json.Unmarshal(respBody, &wbResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Wildberries card upload response: %w", err)
	}
	return &wbResp, nil
}

// UploadMediaFiles uploads multiple media files to Wildberries.
// Corresponds to POST /content/v3/media/file
func (c *WBClient) UploadMediaFiles(ctx context.Context, apiKey string, nmID string, files []entities.WBClientMediaFile) ([]entities.WBMediaUploadResult, error) {
	if apiKey == "" {
		// This error applies to the whole batch.
		return nil, fmt.Errorf("wildberries API key is required for uploading media files")
	}

	uploadURL := fmt.Sprintf("%s/content/v3/media/file", wildberriesAPIHost)
	var results []entities.WBMediaUploadResult

	for _, file := range files {
		select {
		case <-ctx.Done():
			return results, ctx.Err() // Context cancelled or timed out
		default:
		}

		log.Printf("Uploading media file to Wildberries: %s, nmID: %s, photoNumber: %d, fileName: %s", uploadURL, nmID, file.PhotoNumber, file.Filename)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		// "uploadfile" is the field name specified in the OpenAPI spec for /content/v3/media/file
		part, err := writer.CreateFormFile("uploadfile", file.Filename)
		if err != nil {
			results = append(results, entities.WBMediaUploadResult{PhotoNumber: file.PhotoNumber, Error: fmt.Errorf("failed to create form file: %w", err)})
			continue
		}
		_, err = io.Copy(part, bytes.NewReader(file.Content))
		if err != nil {
			results = append(results, entities.WBMediaUploadResult{PhotoNumber: file.PhotoNumber, Error: fmt.Errorf("failed to copy file data to form: %w", err)})
			continue
		}
		err = writer.Close()
		if err != nil {
			results = append(results, entities.WBMediaUploadResult{PhotoNumber: file.PhotoNumber, Error: fmt.Errorf("failed to close multipart writer: %w", err)})
			continue
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", uploadURL, body)
		if err != nil {
			results = append(results, entities.WBMediaUploadResult{PhotoNumber: file.PhotoNumber, Error: fmt.Errorf("failed to create Wildberries media file upload request: %w", err)})
			continue
		}

		httpReq.Header.Set("Authorization", apiKey)
		httpReq.Header.Set("X-Nm-Id", nmID)
		httpReq.Header.Set("X-Photo-Number", fmt.Sprintf("%d", file.PhotoNumber))
		httpReq.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			results = append(results, entities.WBMediaUploadResult{PhotoNumber: file.PhotoNumber, Error: fmt.Errorf("failed to call Wildberries media file upload API: %w", err)})
			continue
		}

		currentResult := processWBMediaResponse(resp, file.PhotoNumber)
		results = append(results, currentResult)
		resp.Body.Close() // Close body inside the loop
	}
	return results, nil
}

// processWBMediaResponse is a helper to read, log, and unmarshal WB media API responses.
func processWBMediaResponse(resp *http.Response, photoNumber int32) entities.WBMediaUploadResult {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return entities.WBMediaUploadResult{PhotoNumber: photoNumber, Error: fmt.Errorf("failed to read Wildberries media file upload response body: %w", err)}
	}

	log.Printf("Wildberries media file upload API response (photo %d) status: %d, body: %s", photoNumber, resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return entities.WBMediaUploadResult{PhotoNumber: photoNumber, Error: fmt.Errorf("wildberries media file upload API (photo %d) returned status %d: %s", photoNumber, resp.StatusCode, string(respBody))}
	}

	var wbResp entities.WBMediaGenericResponse
	if err := json.Unmarshal(respBody, &wbResp); err != nil {
		return entities.WBMediaUploadResult{PhotoNumber: photoNumber, Error: fmt.Errorf("failed to unmarshal Wildberries media file upload response (photo %d): %w. Body: %s", photoNumber, err, string(respBody))}
	}
	return entities.WBMediaUploadResult{PhotoNumber: photoNumber, Response: &wbResp, Error: nil}
}

// SaveMediaByLinks uploads media to Wildberries using provided URLs.
// Corresponds to POST /content/v3/media/save
func (c *WBClient) SaveMediaByLinks(ctx context.Context, apiKey string, payload entities.WBSaveMediaPayload) (*entities.WBMediaGenericResponse, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Wildberries save media payload: %w", err)
	}

	saveURL := fmt.Sprintf("%s/content/v3/media/save", wildberriesAPIHost)
	log.Printf("Saving media by links to Wildberries: %s, Payload: %s", saveURL, string(payloadBytes))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", saveURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Wildberries save media by links request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", apiKey)

	// Response handling is identical to UploadWBCard, so we can reuse that logic or a helper
	// For now, duplicating the common parts:
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Wildberries save media by links API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Wildberries save media by links response body: %w", err)
	}

	log.Printf("Wildberries save media by links API response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		// Note: OpenAPI spec shows 409 and 422 as possible non-200 success/partial success for this endpoint.
		// Handling them specifically might be needed based on business logic.
		// For now, treating non-200 as an error.
		return nil, fmt.Errorf("wildberries save media by links API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var wbResp entities.WBMediaGenericResponse
	if err := json.Unmarshal(respBody, &wbResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Wildberries save media by links response: %w", err)
	}
	return &wbResp, nil
}

// GetCardList retrieves a list of cards from Wildberries.
// Corresponds to POST /content/v2/get/cards/list
func (c *WBClient) GetCardList(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("wildberries API key is required for getting card list")
	}

	payloadBytes, err := json.Marshal(listReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Wildberries get card list payload: %w", err)
	}

	listURL := fmt.Sprintf("%s/content/v2/get/cards/list", wildberriesAPIHost)
	log.Printf("Getting card list from Wildberries: %s, Payload: %s", listURL, string(payloadBytes))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", listURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Wildberries get card list request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Wildberries get card list API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Wildberries get card list response body: %w", err)
	}

	log.Printf("Wildberries get card list API response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wildberries get card list API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var wbResp entities.WBGetCardListResponse
	if err := json.Unmarshal(respBody, &wbResp); err != nil {
		// Log the body for debugging if unmarshal fails
		log.Printf("Failed to unmarshal Wildberries get card list response. Body: %s", string(respBody))
		return nil, fmt.Errorf("failed to unmarshal Wildberries get card list response: %w", err)
	}
	return &wbResp, nil
}
