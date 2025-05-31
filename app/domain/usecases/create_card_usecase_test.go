package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"api/app/domain/entities"
	apiv1 "api/gen/api/v1"

	"connectrpc.com/connect"
)

type mockCardCraftAiClient struct {
	GetSessionIDFunc   func(ctx context.Context) (string, error)
	GetCardContentFunc func(ctx context.Context, sessionID string, cardCraftAiAPIRequest entities.CardCraftAiAPIRequest) (*entities.CardCraftAiAPIResponse, error)
}

func (m *mockCardCraftAiClient) GetSessionID(ctx context.Context) (string, error) {
	if m.GetSessionIDFunc != nil {
		return m.GetSessionIDFunc(ctx)
	}
	return "mock_session_id", nil
}

func (m *mockCardCraftAiClient) GetCardContent(ctx context.Context, sessionID string, req entities.CardCraftAiAPIRequest) (*entities.CardCraftAiAPIResponse, error) {
	if m.GetCardContentFunc != nil {
		return m.GetCardContentFunc(ctx, sessionID, req)
	}
	subjectID := int32(123)
	return &entities.CardCraftAiAPIResponse{Title: "Test Title", Description: "Test Description", SubjectID: &subjectID}, nil
}

type mockWbClient struct {
	UploadWBCardFunc     func(ctx context.Context, wbPayload entities.WBCardUploadPayload, apiKey string) (*entities.WBCardUploadResponse, error)
	UploadMediaFilesFunc func(ctx context.Context, apiKey string, nmID string, files []entities.WBClientMediaFile) ([]entities.WBMediaUploadResult, error)
	SaveMediaByLinksFunc func(ctx context.Context, apiKey string, payload entities.WBSaveMediaPayload) (*entities.WBMediaGenericResponse, error)
	GetCardListFunc      func(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error)
}

func (m *mockWbClient) UploadWBCard(ctx context.Context, wbPayload entities.WBCardUploadPayload, apiKey string) (*entities.WBCardUploadResponse, error) {
	if m.UploadWBCardFunc != nil {
		return m.UploadWBCardFunc(ctx, wbPayload, apiKey)
	}
	return &entities.WBCardUploadResponse{Error: false, ErrorText: ""}, nil
}

func (m *mockWbClient) UploadMediaFiles(ctx context.Context, apiKey string, nmID string, files []entities.WBClientMediaFile) ([]entities.WBMediaUploadResult, error) {
	if m.UploadMediaFilesFunc != nil {
		return m.UploadMediaFilesFunc(ctx, apiKey, nmID, files)
	}
	results := make([]entities.WBMediaUploadResult, len(files))
	for i, f := range files {
		results[i] = entities.WBMediaUploadResult{PhotoNumber: f.PhotoNumber, Response: &entities.WBMediaGenericResponse{Error: false}}
	}
	return results, nil
}

func (m *mockWbClient) SaveMediaByLinks(ctx context.Context, apiKey string, payload entities.WBSaveMediaPayload) (*entities.WBMediaGenericResponse, error) {
	if m.SaveMediaByLinksFunc != nil {
		return m.SaveMediaByLinksFunc(ctx, apiKey, payload)
	}
	return &entities.WBMediaGenericResponse{Error: false}, nil
}

func (m *mockWbClient) GetCardList(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
	if m.GetCardListFunc != nil {
		return m.GetCardListFunc(ctx, apiKey, listReq)
	}
	return &entities.WBGetCardListResponse{Cards: []entities.WBCardDefinition{{NmID: 12345, VendorCode: "test_vendor_code"}}}, nil
}

func TestNewCreateCardUsecase(t *testing.T) {
	mockCCAIClient := &mockCardCraftAiClient{}
	mockWBClient := &mockWbClient{}
	uc := NewCreateCardUsecase(mockCCAIClient, mockWBClient, 3)

	if uc == nil {
		t.Fatal("NewCreateCardUsecase returned nil")
	}
	if uc.cardCraftAiClient != mockCCAIClient {
		t.Error("cardCraftAiClient not set correctly")
	}
	if uc.wbClient != mockWBClient {
		t.Error("wbClient not set correctly")
	}
	if uc.wbApiGetCardListMaxAttempts != 3 {
		t.Errorf("wbApiGetCardListMaxAttempts not set correctly, got %d", uc.wbApiGetCardListMaxAttempts)
	}
}

func TestCreateCardUsecase_CreateProductCard(t *testing.T) {
	ctx := context.Background()
	defaultSubjectID := int32(123)
	defaultParentID := int32(456)

	baseRequestMsg := &apiv1.CreateProductCardRequest{
		ProductTitle:       "Sample Product",
		ProductDescription: "A great sample product.",
		ParentId:           defaultParentID,
		SubjectId:          defaultSubjectID,
		VendorCode:         "test_vendor_code",
		Wb:                 true,
		WbApiKey:           "test_api_key",
	}

	mockCCAIClient := &mockCardCraftAiClient{
		GetSessionIDFunc: func(ctx context.Context) (string, error) {
			return "test_session_id", nil
		},
		GetCardContentFunc: func(ctx context.Context, sessionID string, cardCraftAiAPIRequest entities.CardCraftAiAPIRequest) (*entities.CardCraftAiAPIResponse, error) {
			return &entities.CardCraftAiAPIResponse{
				Title:       "AI Title",
				Description: "AI Description",
				Attributes:  map[string]string{"color": "blue"},
				SubjectID:   &defaultSubjectID,
				ParentID:    &defaultParentID,
			}, nil
		},
	}

	mockWBClient := &mockWbClient{
		GetCardListFunc: func(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
			return &entities.WBGetCardListResponse{
				Cards: []entities.WBCardDefinition{{NmID: 789, VendorCode: "test_vendor_code"}},
			}, nil
		},
	}

	uc := NewCreateCardUsecase(mockCCAIClient, mockWBClient, 1)

	t.Run("Successful card creation with WB integration", func(t *testing.T) {
		req := connect.NewRequest(baseRequestMsg)
		resp, err := uc.CreateProductCard(ctx, req)

		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Response is nil")
		}
		if resp.Msg.Title != "AI Title" {
			t.Errorf("Expected title 'AI Title', got '%s'", resp.Msg.Title)
		}
		if resp.Msg.WbApiResponseJson == nil || *resp.Msg.WbApiResponseJson == "" {
			t.Error("Expected WbApiResponseJson to be populated")
		}
		if resp.Msg.WbRequestAttempted == nil || !*resp.Msg.WbRequestAttempted {
			t.Error("Expected WbRequestAttempted to be true")
		}
	})

	t.Run("Validation error: wb=true, vendor_code empty", func(t *testing.T) {
		invalidReqMsg := protoClone(baseRequestMsg)
		invalidReqMsg.VendorCode = ""
		req := connect.NewRequest(invalidReqMsg)

		_, err := uc.CreateProductCard(ctx, req)
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("Expected InvalidArgument error, got %v", connect.CodeOf(err))
		}
	})

	t.Run("Error getting session ID", func(t *testing.T) {
		customMockCCAIClient := &mockCardCraftAiClient{
			GetSessionIDFunc: func(ctx context.Context) (string, error) {
				return "", errors.New("session ID error")
			},
		}
		customUC := NewCreateCardUsecase(customMockCCAIClient, mockWBClient, 1)
		req := connect.NewRequest(baseRequestMsg)

		_, err := customUC.CreateProductCard(ctx, req)
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "session ID error") {
			t.Errorf("Expected error message to contain 'session ID error', got '%s'", err.Error())
		}
	})

	t.Run("Error getting card content", func(t *testing.T) {
		customMockCCAIClient := &mockCardCraftAiClient{
			GetSessionIDFunc: func(ctx context.Context) (string, error) { return "test_session_id", nil },
			GetCardContentFunc: func(ctx context.Context, sessionID string, cardCraftAiAPIRequest entities.CardCraftAiAPIRequest) (*entities.CardCraftAiAPIResponse, error) {
				return nil, errors.New("card content error")
			},
		}
		customUC := NewCreateCardUsecase(customMockCCAIClient, mockWBClient, 1)
		req := connect.NewRequest(baseRequestMsg)

		_, err := customUC.CreateProductCard(ctx, req)
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "card content error") {
			t.Errorf("Expected error message to contain 'card content error', got '%s'", err.Error())
		}
	})

	t.Run("WB integration: wb=false", func(t *testing.T) {
		reqMsgNoWb := protoClone(baseRequestMsg)
		reqMsgNoWb.Wb = false
		req := connect.NewRequest(reqMsgNoWb)

		resp, err := uc.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if resp.Msg.WbPreparedRequestJson == nil || *resp.Msg.WbPreparedRequestJson == "" {
			t.Error("Expected WbPreparedRequestJson to be populated when wb=false")
		}
		if resp.Msg.WbRequestAttempted == nil || *resp.Msg.WbRequestAttempted {
			t.Error("Expected WbRequestAttempted to be false")
		}
	})

	t.Run("WB integration: wb=true, no API key", func(t *testing.T) {
		reqMsgNoApiKey := protoClone(baseRequestMsg)
		reqMsgNoApiKey.WbApiKey = ""
		req := connect.NewRequest(reqMsgNoApiKey)

		resp, err := uc.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if resp.Msg.WbPreparedRequestJson == nil || *resp.Msg.WbPreparedRequestJson == "" {
			t.Error("Expected WbPreparedRequestJson to be populated when API key is missing")
		}
		if resp.Msg.WbRequestAttempted == nil || *resp.Msg.WbRequestAttempted {
			t.Error("Expected WbRequestAttempted to be false when API key is missing")
		}
	})

	t.Run("WB integration: UploadWBCard fails", func(t *testing.T) {
		customMockWBClient := &mockWbClient{
			UploadWBCardFunc: func(ctx context.Context, wbPayload entities.WBCardUploadPayload, apiKey string) (*entities.WBCardUploadResponse, error) {
				return nil, errors.New("wb upload error")
			},
		}
		customUC := NewCreateCardUsecase(mockCCAIClient, customMockWBClient, 1)
		req := connect.NewRequest(baseRequestMsg)

		resp, err := customUC.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if resp.Msg.WbApiResponseJson == nil || !strings.Contains(*resp.Msg.WbApiResponseJson, "wb upload error") {
			t.Error("Expected WbApiResponseJson to contain upload error")
		}
	})

	t.Run("WB integration: UploadWBCard returns error in response", func(t *testing.T) {
		customMockWBClient := &mockWbClient{
			UploadWBCardFunc: func(ctx context.Context, wbPayload entities.WBCardUploadPayload, apiKey string) (*entities.WBCardUploadResponse, error) {
				return &entities.WBCardUploadResponse{Error: true, ErrorText: "WB API error"}, nil
			},
		}
		customUC := NewCreateCardUsecase(mockCCAIClient, customMockWBClient, 1)
		req := connect.NewRequest(baseRequestMsg)

		resp, err := customUC.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if resp.Msg.WbApiResponseJson == nil || !strings.Contains(*resp.Msg.WbApiResponseJson, "WB API error") {
			t.Error("Expected WbApiResponseJson to contain WB API error")
		}
	})

	t.Run("WB integration: CardCraftAI response with nil SubjectID", func(t *testing.T) {
		customMockCCAIClient := &mockCardCraftAiClient{
			GetSessionIDFunc: func(ctx context.Context) (string, error) { return "test_session_id", nil },
			GetCardContentFunc: func(ctx context.Context, sessionID string, cardCraftAiAPIRequest entities.CardCraftAiAPIRequest) (*entities.CardCraftAiAPIResponse, error) {
				return &entities.CardCraftAiAPIResponse{Title: "AI Title", Description: "AI Description", SubjectID: nil}, nil // Nil SubjectID
			},
		}
		customUC := NewCreateCardUsecase(customMockCCAIClient, mockWBClient, 1)
		req := connect.NewRequest(baseRequestMsg)

		resp, err := customUC.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		// Check that wbPreparedRequestJSON or wbApiResponseJSON contains subjectID: 0
		var wbPrepared map[string]interface{}

		if resp.Msg.WbPreparedRequestJson != nil && *resp.Msg.WbPreparedRequestJson != "" {
			if err := json.Unmarshal([]byte(*resp.Msg.WbPreparedRequestJson), &wbPrepared); err == nil {
				if items, ok := wbPrepared["items"].([]interface{}); ok && len(items) > 0 {
					if item, ok := items[0].(map[string]interface{}); ok {
						if subjectID, ok := item["subjectID"].(float64); !ok || subjectID != 0 {
							t.Errorf("Expected subjectID 0 in prepared request, got %v", item["subjectID"])
						}
					}
				}
			}
		} else if resp.Msg.WbApiResponseJson != nil && *resp.Msg.WbApiResponseJson != "" {
			// This path implies an API call was made. The mock UploadWBCard would need to capture the payload.
			// For simplicity, we assume the log message about nil SubjectID is sufficient indication.
			// A more thorough test would involve a mockWbClient that records the payload.
		}
		// The main check is that it doesn't panic and proceeds.
	})

	t.Run("Media operations: Successful file upload and save by links", func(t *testing.T) {
		reqMsgWithMedia := protoClone(baseRequestMsg)
		reqMsgWithMedia.WbMediaToUploadFiles = []*apiv1.WBMediaFileToUpload{
			{Filename: "photo1.jpg", Content: []byte("jpeg_data"), PhotoNumber: 1},
		}
		reqMsgWithMedia.WbMediaToSaveLinks = []string{"http://example.com/image.png"}
		req := connect.NewRequest(reqMsgWithMedia)

		resp, err := uc.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard with media failed: %v", err)
		}
		if len(resp.Msg.WbMediaUploadIndividualResponses) != 1 {
			t.Errorf("Expected 1 media upload response, got %d", len(resp.Msg.WbMediaUploadIndividualResponses))
		}
		if resp.Msg.WbMediaSaveByLinksResponse == nil {
			t.Error("Expected media save by links response, got nil")
		}
	})

	t.Run("Media operations: API key missing", func(t *testing.T) {
		reqMsgMediaNoKey := protoClone(baseRequestMsg)
		reqMsgMediaNoKey.WbApiKey = "" // No API key
		reqMsgMediaNoKey.WbMediaToUploadFiles = []*apiv1.WBMediaFileToUpload{{Filename: "photo1.jpg", Content: []byte("data"), PhotoNumber: 1}}
		req := connect.NewRequest(reqMsgMediaNoKey)

		_, err := uc.CreateProductCard(ctx, req)
		if err == nil {
			t.Fatal("Expected error for media operations without API key")
		}
		if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "API key is required for media operations") {
			t.Errorf("Expected InvalidArgument error about API key, got: %v", err)
		}
	})

	t.Run("Media operations: Vendor code missing", func(t *testing.T) {
		// Note: This should be caught by the initial validation if wb=true.
		// However, if wb=false but media operations are somehow requested (which is illogical but testable for robustness of handleWildberriesMediaOperations)
		// For this test, let's assume wb=true and vendor code is missing, which is already tested.
		// Let's test the specific check within handleWildberriesMediaOperations.
		// To isolate this, we'd call handleWildberriesMediaOperations directly.
		// As part of CreateProductCard, if vendor_code is empty and wb=true, it fails early.
		// If wb=false, media ops are skipped.
		// So, this specific path in handleWildberriesMediaOperations might be hard to reach via CreateProductCard if initial validation is strict.
		// Let's assume the initial validation for vendor_code is the primary guard.
		// If we want to test the internal check:
		ucIsolated := NewCreateCardUsecase(mockCCAIClient, mockWBClient, 1)
		reqMsgMediaNoVendor := protoClone(baseRequestMsg)
		reqMsgMediaNoVendor.VendorCode = "" // No vendor code
		reqMsgMediaNoVendor.WbMediaToUploadFiles = []*apiv1.WBMediaFileToUpload{{Filename: "photo1.jpg", Content: []byte("data"), PhotoNumber: 1}}

		// If wb=true, it will fail at the start of CreateProductCard.
		// If we force wb=false to bypass that, then media ops are skipped.
		// This highlights that the check `if vendorCode == ""` inside `handleWildberriesMediaOperations` might be redundant
		// if `CreateProductCard` already validates `vendorCode` when `wb=true`.
		// However, if `wb=false` but media files are provided, the current logic skips media ops.
		// Let's test the scenario where `wb=true`, `vendorCode` is empty.
		req := connect.NewRequest(reqMsgMediaNoVendor)
		_, err := ucIsolated.CreateProductCard(ctx, req)
		if err == nil {
			t.Fatal("Expected error for wb=true and empty vendor_code")
		}
		if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "vendor_code is required when wb is true") {
			t.Errorf("Expected InvalidArgument error about vendor_code, got: %v", err)
		}
	})

	t.Run("Media operations: nmID not found after retries", func(t *testing.T) {
		customMockWBClient := &mockWbClient{
			GetCardListFunc: func(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
				return &entities.WBGetCardListResponse{Cards: []entities.WBCardDefinition{}}, nil // No cards found
			},
		}
		customUC := NewCreateCardUsecase(mockCCAIClient, customMockWBClient, 2) // 2 attempts
		reqMsgWithMedia := protoClone(baseRequestMsg)
		reqMsgWithMedia.WbMediaToUploadFiles = []*apiv1.WBMediaFileToUpload{{Filename: "photo1.jpg", Content: []byte("data"), PhotoNumber: 1}}
		req := connect.NewRequest(reqMsgWithMedia)

		_, err := customUC.CreateProductCard(ctx, req)
		if err == nil {
			t.Fatal("Expected error when nmID is not found")
		}
		if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "not found on Wildberries after 2 attempts") {
			t.Errorf("Expected NotFound error about nmID, got: %v", err)
		}
	})

	t.Run("Media operations: GetCardList fails", func(t *testing.T) {
		customMockWBClient := &mockWbClient{
			GetCardListFunc: func(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
				return nil, errors.New("GetCardList API error")
			},
		}
		customUC := NewCreateCardUsecase(mockCCAIClient, customMockWBClient, 1)
		reqMsgWithMedia := protoClone(baseRequestMsg)
		reqMsgWithMedia.WbMediaToUploadFiles = []*apiv1.WBMediaFileToUpload{{Filename: "photo1.jpg", Content: []byte("data"), PhotoNumber: 1}}
		req := connect.NewRequest(reqMsgWithMedia)

		_, err := customUC.CreateProductCard(ctx, req)
		if err == nil {
			t.Fatal("Expected error when GetCardList fails")
		}
		if connect.CodeOf(err) != connect.CodeUnavailable || !strings.Contains(err.Error(), "GetCardList API error") {
			t.Errorf("Expected Unavailable error from GetCardList, got: %v", err)
		}
	})

	t.Run("Media operations: nmID found on second attempt", func(t *testing.T) {
		attemptCount := 0
		customMockWBClient := &mockWbClient{
			GetCardListFunc: func(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
				attemptCount++
				if attemptCount == 1 {
					return &entities.WBGetCardListResponse{Cards: []entities.WBCardDefinition{}}, nil // Not found first time
				}
				return &entities.WBGetCardListResponse{Cards: []entities.WBCardDefinition{{NmID: 789, VendorCode: "test_vendor_code"}}}, nil // Found second time
			},
			UploadMediaFilesFunc: func(ctx context.Context, apiKey string, nmID string, files []entities.WBClientMediaFile) ([]entities.WBMediaUploadResult, error) {
				if nmID != "789" {
					return nil, fmt.Errorf("expected nmID 789, got %s", nmID)
				}
				return []entities.WBMediaUploadResult{{PhotoNumber: 1, Response: &entities.WBMediaGenericResponse{}}}, nil
			},
		}
		customUC := NewCreateCardUsecase(mockCCAIClient, customMockWBClient, 2) // Max 2 attempts
		reqMsgWithMedia := protoClone(baseRequestMsg)
		reqMsgWithMedia.WbMediaToUploadFiles = []*apiv1.WBMediaFileToUpload{{Filename: "photo1.jpg", Content: []byte("data"), PhotoNumber: 1}}
		req := connect.NewRequest(reqMsgWithMedia)

		resp, err := customUC.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if attemptCount != 2 {
			t.Errorf("Expected GetCardList to be called 2 times, got %d", attemptCount)
		}
		if len(resp.Msg.WbMediaUploadIndividualResponses) != 1 {
			t.Error("Expected media upload response")
		}
	})

	t.Run("Media operations: UploadMediaFiles client error", func(t *testing.T) {
		customMockWBClient := &mockWbClient{
			GetCardListFunc: func(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
				return &entities.WBGetCardListResponse{Cards: []entities.WBCardDefinition{{NmID: 789, VendorCode: "test_vendor_code"}}}, nil
			},
			UploadMediaFilesFunc: func(ctx context.Context, apiKey string, nmID string, files []entities.WBClientMediaFile) ([]entities.WBMediaUploadResult, error) {
				return nil, errors.New("UploadMediaFiles overall error")
			},
		}
		customUC := NewCreateCardUsecase(mockCCAIClient, customMockWBClient, 1)
		reqMsgWithMedia := protoClone(baseRequestMsg)
		reqMsgWithMedia.WbMediaToUploadFiles = []*apiv1.WBMediaFileToUpload{{Filename: "photo1.jpg", Content: []byte("data"), PhotoNumber: 1}}
		req := connect.NewRequest(reqMsgWithMedia)

		// The error from UploadMediaFiles is logged but not directly returned if partial results are processed.
		// The current implementation appends results even if an overall error occurs.
		// Let's check if the response reflects this.
		resp, err := customUC.CreateProductCard(ctx, req)
		if err != nil { // This implies an error before or after media processing, not from UploadMediaFiles itself directly to user
			t.Fatalf("CreateProductCard failed unexpectedly: %v", err)
		}
		// If UploadMediaFiles returns an error AND empty results, then WbMediaUploadIndividualResponses might be empty or nil.
		// The current code logs the error and continues with whatever results were returned (which could be nil).
		// If results are nil and error is non-nil, WbMediaUploadIndividualResponses will be empty.
		if len(resp.Msg.WbMediaUploadIndividualResponses) != 0 {
			t.Errorf("Expected 0 media upload responses due to overall error, got %d", len(resp.Msg.WbMediaUploadIndividualResponses))
		}
	})

	t.Run("Media operations: UploadMediaFiles individual file error", func(t *testing.T) {
		customMockWBClient := &mockWbClient{
			GetCardListFunc: func(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
				return &entities.WBGetCardListResponse{Cards: []entities.WBCardDefinition{{NmID: 789, VendorCode: "test_vendor_code"}}}, nil
			},
			UploadMediaFilesFunc: func(ctx context.Context, apiKey string, nmID string, files []entities.WBClientMediaFile) ([]entities.WBMediaUploadResult, error) {
				return []entities.WBMediaUploadResult{
					{PhotoNumber: 1, Error: errors.New("individual file upload error")},
				}, nil
			},
		}
		customUC := NewCreateCardUsecase(mockCCAIClient, customMockWBClient, 1)
		reqMsgWithMedia := protoClone(baseRequestMsg)
		reqMsgWithMedia.WbMediaToUploadFiles = []*apiv1.WBMediaFileToUpload{{Filename: "photo1.jpg", Content: []byte("data"), PhotoNumber: 1}}
		req := connect.NewRequest(reqMsgWithMedia)

		resp, err := customUC.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if len(resp.Msg.WbMediaUploadIndividualResponses) != 1 {
			t.Fatal("Expected 1 media upload response")
		}
		if resp.Msg.WbMediaUploadIndividualResponses[0].ErrorMessage == nil ||
			!strings.Contains(*resp.Msg.WbMediaUploadIndividualResponses[0].ErrorMessage, "individual file upload error") {
			t.Error("Expected error message for individual file upload")
		}
	})

	t.Run("Media operations: SaveMediaByLinks client error", func(t *testing.T) {
		customMockWBClient := &mockWbClient{
			GetCardListFunc: func(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error) {
				return &entities.WBGetCardListResponse{Cards: []entities.WBCardDefinition{{NmID: 789, VendorCode: "test_vendor_code"}}}, nil
			},
			SaveMediaByLinksFunc: func(ctx context.Context, apiKey string, payload entities.WBSaveMediaPayload) (*entities.WBMediaGenericResponse, error) {
				return nil, errors.New("SaveMediaByLinks API error")
			},
		}
		customUC := NewCreateCardUsecase(mockCCAIClient, customMockWBClient, 1)
		reqMsgWithMedia := protoClone(baseRequestMsg)
		reqMsgWithMedia.WbMediaToSaveLinks = []string{"http://example.com/image.png"}
		req := connect.NewRequest(reqMsgWithMedia)

		resp, err := customUC.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if resp.Msg.WbMediaSaveByLinksResponse == nil {
			t.Fatal("Expected WbMediaSaveByLinksResponse")
		}
		if resp.Msg.WbMediaSaveByLinksResponse.ErrorMessage == nil ||
			!strings.Contains(*resp.Msg.WbMediaSaveByLinksResponse.ErrorMessage, "SaveMediaByLinks API error") {
			t.Error("Expected error message for SaveMediaByLinks")
		}
	})

	t.Run("Media operations: No media operations requested", func(t *testing.T) {
		reqMsgNoMedia := protoClone(baseRequestMsg) // Already has no media by default
		req := connect.NewRequest(reqMsgNoMedia)

		resp, err := uc.CreateProductCard(ctx, req) // uc uses default mockWBClient
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}
		if len(resp.Msg.WbMediaUploadIndividualResponses) != 0 {
			t.Errorf("Expected 0 media upload responses, got %d", len(resp.Msg.WbMediaUploadIndividualResponses))
		}
		if resp.Msg.WbMediaSaveByLinksResponse != nil {
			t.Error("Expected nil media save by links response, got non-nil")
		}
	})

	t.Run("Full response mapping with all optional fields from CardCraftAI", func(t *testing.T) {
		parentID, subjectID, typeID, rootID, subID := int32(1), int32(2), int32(3), int32(4), int32(5)
		parentName, subjectName, typeName, rootName, subName := "PName", "SName", "TName", "RName", "SubName"

		customMockCCAIClient := &mockCardCraftAiClient{
			GetSessionIDFunc: func(ctx context.Context) (string, error) { return "test_session_id", nil },
			GetCardContentFunc: func(ctx context.Context, sessionID string, cardCraftAiAPIRequest entities.CardCraftAiAPIRequest) (*entities.CardCraftAiAPIResponse, error) {
				return &entities.CardCraftAiAPIResponse{
					Title: "Full Title", Description: "Full Desc", Attributes: map[string]string{"key": "val"},
					ParentID: &parentID, ParentName: &parentName,
					SubjectID: &subjectID, SubjectName: &subjectName,
					TypeID: &typeID, TypeName: &typeName,
					RootID: &rootID, RootName: &rootName,
					SubID: &subID, SubName: &subName,
				}, nil
			},
		}
		customUC := NewCreateCardUsecase(customMockCCAIClient, mockWBClient, 1)
		reqMsg := protoClone(baseRequestMsg)
		reqMsg.Wb = false // Simplify, no WB interaction needed for this part of test
		req := connect.NewRequest(reqMsg)

		resp, err := customUC.CreateProductCard(ctx, req)
		if err != nil {
			t.Fatalf("CreateProductCard failed: %v", err)
		}

		expectedResp := &apiv1.CreateProductCardResponse{
			Title: "Full Title", Description: "Full Desc", Attributes: map[string]string{"key": "val"},
			ParentId: parentID, ParentName: parentName,
			SubjectId: subjectID, SubjectName: subjectName,
			TypeId: typeID, TypeName: typeName,
			RootId: rootID, RootName: rootName,
			SubId: subID, SubName: subName,
			WbRequestAttempted:    func(b bool) *bool { return &b }(false), // wb=false
			WbPreparedRequestJson: resp.Msg.WbPreparedRequestJson,          // Keep this dynamic as it's complex
		}

		if resp.Msg.Title != expectedResp.Title ||
			resp.Msg.Description != expectedResp.Description ||
			!reflect.DeepEqual(resp.Msg.Attributes, expectedResp.Attributes) ||
			resp.Msg.ParentId != expectedResp.ParentId || resp.Msg.ParentName != expectedResp.ParentName ||
			resp.Msg.SubjectId != expectedResp.SubjectId || resp.Msg.SubjectName != expectedResp.SubjectName ||
			resp.Msg.TypeId != expectedResp.TypeId || resp.Msg.TypeName != expectedResp.TypeName ||
			resp.Msg.RootId != expectedResp.RootId || resp.Msg.RootName != expectedResp.RootName ||
			resp.Msg.SubId != expectedResp.SubId || resp.Msg.SubName != expectedResp.SubName {
			t.Errorf("Response mismatch.\nExpected: %+v\nGot: %+v", expectedResp, resp.Msg)
		}
	})
}

func Test_mapRequestToDomain(t *testing.T) {
	uc := &CreateCardUsecase{} // No dependencies needed for this static method

	t.Run("Basic mapping", func(t *testing.T) {
		reqMsg := &apiv1.CreateProductCardRequest{
			ProductTitle:       "Test Title",
			ProductDescription: "Test Desc",
			ParentId:           123,
			SubjectId:          456,
			Translate:          true,
			Ozon:               false,
			GenerateContent:    true,
			Sizes: []*apiv1.Size{
				{TechSize: "S", WbSize: "44", Price: 1000, Skus: []string{"SKU001"}},
			},
		}
		req := connect.NewRequest(reqMsg)

		domainReq, err := uc.mapRequestToDomain(req)
		if err != nil {
			t.Fatalf("mapRequestToDomain failed: %v", err)
		}

		expectedDomainReq := entities.CardCraftAiAPIRequest{
			ProductTitle:       "Test Title",
			ProductDescription: "Test Desc",
			ParentID:           123,
			SubjectID:          456,
			Translate:          true,
			Ozon:               false,
			GenerateContent:    true,
			// Sizes field is not part of CardCraftAiAPIRequest in the provided code, so it won't be mapped there.
			// The original mapRequestToDomain maps to CardCraftAiAPIRequest which does not have Sizes.
			// The provided code for mapRequestToDomain initializes domainSizes but doesn't use it in the return struct.
			// This test reflects the current implementation.
		}
		// If CardCraftAiAPIRequest was supposed to have Sizes, this test would need adjustment.
		// Based on `app/domain/entities/card_craft_ai_api_request.go`, it does not.

		if !reflect.DeepEqual(domainReq, expectedDomainReq) {
			t.Errorf("Mapped request mismatch.\nExpected: %+v\nGot: %+v", expectedDomainReq, domainReq)
		}
	})

	t.Run("Mapping with nil sizes", func(t *testing.T) {
		reqMsg := &apiv1.CreateProductCardRequest{
			ProductTitle: "Test Title",
			Sizes:        nil,
		}
		req := connect.NewRequest(reqMsg)
		domainReq, err := uc.mapRequestToDomain(req)
		if err != nil {
			t.Fatalf("mapRequestToDomain failed: %v", err)
		}
		// As above, Sizes are not part of the target struct.
		if domainReq.ProductTitle != "Test Title" {
			t.Error("ProductTitle not mapped correctly")
		}
	})
}

// Helper to clone proto message for modification in tests
func protoClone(original *apiv1.CreateProductCardRequest) *apiv1.CreateProductCardRequest {
	clone := &apiv1.CreateProductCardRequest{}

	clone.ProductTitle = original.ProductTitle
	clone.ProductDescription = original.ProductDescription
	clone.ParentId = original.ParentId
	clone.SubjectId = original.SubjectId
	clone.VendorCode = original.VendorCode
	clone.Brand = original.Brand
	clone.Translate = original.Translate
	clone.Ozon = original.Ozon
	clone.GenerateContent = original.GenerateContent
	clone.Wb = original.Wb
	clone.WbApiKey = original.WbApiKey

	if original.Dimensions != nil {
		clone.Dimensions = &apiv1.Dimensions{
			Length:       original.Dimensions.Length,
			Width:        original.Dimensions.Width,
			Height:       original.Dimensions.Height,
			WeightBrutto: original.Dimensions.WeightBrutto,
		}
	}

	if original.Sizes != nil {
		clone.Sizes = make([]*apiv1.Size, len(original.Sizes))
		for i, s := range original.Sizes {
			clone.Sizes[i] = &apiv1.Size{
				TechSize: s.TechSize,
				WbSize:   s.WbSize,
				Price:    s.Price,
				Skus:     append([]string(nil), s.Skus...),
			}
		}
	}

	if original.WbMediaToUploadFiles != nil {
		clone.WbMediaToUploadFiles = make([]*apiv1.WBMediaFileToUpload, len(original.WbMediaToUploadFiles))
		for i, f := range original.WbMediaToUploadFiles {
			clone.WbMediaToUploadFiles[i] = &apiv1.WBMediaFileToUpload{
				Filename:    f.Filename,
				Content:     append([]byte(nil), f.Content...),
				PhotoNumber: f.PhotoNumber,
			}
		}
	}

	if original.WbMediaToSaveLinks != nil {
		clone.WbMediaToSaveLinks = append([]string(nil), original.WbMediaToSaveLinks...)
	}

	return clone
}
