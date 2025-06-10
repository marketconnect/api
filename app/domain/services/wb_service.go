package services

import (
	"api/app/domain/entities"
	"api/metrics"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"connectrpc.com/connect"
)

type wbClient interface {
	UploadWBCard(ctx context.Context, wbPayload entities.WBCardUploadPayload, apiKey string) (*entities.WBCardUploadResponse, error)
	UploadMediaFiles(ctx context.Context, apiKey string, nmID string, files []entities.WBClientMediaFile) ([]entities.WBMediaUploadResult, error)
	SaveMediaByLinks(ctx context.Context, apiKey string, payload entities.WBSaveMediaPayload) (*entities.WBMediaGenericResponse, error)
	GetCardList(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error)
}

type WbService struct {
	wbApiGetCardListMaxAttempts int
	wbClient                    wbClient
}

func NewWbService(wbApiGetCardListMaxAttempts int, wbClient wbClient) *WbService {
	return &WbService{
		wbApiGetCardListMaxAttempts: wbApiGetCardListMaxAttempts,
		wbClient:                    wbClient,
	}
}

func (wbs *WbService) AddMedia(ctx context.Context, req *entities.ProductCard) ([]*entities.WbMediaUploadIndividualResponse, *entities.WbMediaSaveByLinksResponse, error) {
	var protoMediaUploadResponses []*entities.WbMediaUploadIndividualResponse
	var protoMediaSaveResponse *entities.WbMediaSaveByLinksResponse

	// If no media operations are requested, return early.
	if len(req.GetWbMediaToUploadFiles()) == 0 && len(req.GetWbMediaToSaveLinks()) == 0 {
		return nil, nil, nil
	}

	apiKey := req.GetWbApiKey()
	vendorCode := req.GetVendorCode()

	if apiKey == "" {
		log.Printf("Wildberries API key not provided, skipping media operations for vendor code %s.", vendorCode)
		return nil, nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("wildberries API key is required for media operations"))
	}

	if vendorCode == "" {
		log.Printf("Vendor code not provided, skipping media operations.") // Should be caught by initial validation, but good to check.
		return nil, nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("vendor_code is required for Wildberries media operations"))
	}

	var foundNmID int

	const retryDelay = 5 * time.Second

	for attempt := 1; attempt <= wbs.wbApiGetCardListMaxAttempts; attempt++ {
		log.Printf("Attempt %d to find nmID for vendor code %s", attempt, vendorCode)
		filterWithPhoto := 0 // Only cards with no photos
		listReqPayload := entities.WBGetCardListRequest{
			Settings: entities.WBGetCardListRequestSettings{
				Filter: &entities.WBGetCardListRequestFilter{
					WithPhoto: &filterWithPhoto,
				},
			},
		}

		wbCardsResp, err := wbs.wbClient.GetCardList(ctx, apiKey, listReqPayload)
		if err != nil {
			log.Printf("Error getting card list from WB (attempt %d) for vendor code %s: %v", attempt, vendorCode, err)
			metrics.AppExternalAPIErrorsTotal.WithLabelValues("wb_get_card_list").Inc()
			if attempt == wbs.wbApiGetCardListMaxAttempts {
				return nil, nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to get card list from WB after %d attempts for vendor code %s: %w", wbs.wbApiGetCardListMaxAttempts, vendorCode, err))
			}
			time.Sleep(retryDelay)
			continue
		}

		for _, card := range wbCardsResp.Cards {
			fmt.Printf("Card: %s - %d\n", card.VendorCode, card.NmID)
			if card.VendorCode == vendorCode {
				foundNmID = card.NmID
				log.Printf("Found nmID %d for vendor code %s", foundNmID, vendorCode)
				break
			}
		}

		if foundNmID != 0 {
			break // Found nmID, exit retry loop
		}

		if attempt < wbs.wbApiGetCardListMaxAttempts {
			log.Printf("Card with vendor code %s not found in attempt %d. Retrying in %v...", vendorCode, attempt, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if foundNmID == 0 {
		log.Printf("Failed to find nmID for vendor code %s after %d attempts.", vendorCode, wbs.wbApiGetCardListMaxAttempts)
		return nil, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("card with vendor code '%s' not found on Wildberries after %d attempts", vendorCode, wbs.wbApiGetCardListMaxAttempts))
	}

	// Handle file uploads
	if len(req.GetWbMediaToUploadFiles()) > 0 {
		log.Printf("Attempting to upload %d media files to Wildberries for nmID %d.", len(req.GetWbMediaToUploadFiles()), foundNmID)
		clientMediaFiles := make([]entities.WBClientMediaFile, len(req.WbMediaToUploadFiles))
		for i, f := range req.WbMediaToUploadFiles {
			metrics.AppWBMediaOperationsTotal.WithLabelValues("upload_file").Inc()
			clientMediaFiles[i] = entities.WBClientMediaFile{
				Filename:    f.Filename,
				Content:     f.Content,
				PhotoNumber: f.PhotoNumber,
			}
		}

		uploadResults, err := wbs.wbClient.UploadMediaFiles(ctx, apiKey, fmt.Sprintf("%d", foundNmID), clientMediaFiles)
		if err != nil {
			// This error is for the whole operation, e.g., context cancellation before starting.
			metrics.AppWBMediaOperationErrorsTotal.WithLabelValues("upload_file_batch").Inc() // A general error for the batch
			log.Printf("Overall error calling UploadMediaFiles for nmID %d: %v", foundNmID, err)
			// We might still have partial results in uploadResults, or it might be nil.
			// For now, we'll process any results returned.
		}

		for _, res := range uploadResults {
			protoResp := &entities.WbMediaUploadIndividualResponse{
				PhotoNumber: res.PhotoNumber,
			}
			if res.Response != nil {
				jsonBytes, jsonErr := json.Marshal(res.Response)
				// Check res.Response.Error as well
				if jsonErr != nil {
					log.Printf("Error marshalling WB media upload response for photo %d, nmID %d: %v", res.PhotoNumber, foundNmID, jsonErr)
					errMsg := fmt.Sprintf("Failed to marshal WB media upload response: %v", jsonErr)
					protoResp.ErrorMessage = &errMsg
				} else {
					s := string(jsonBytes)
					protoResp.ResponseJson = &s
				}
				if res.Response.Error {
					metrics.AppWBMediaOperationErrorsTotal.WithLabelValues("upload_file").Inc()
				}
			}
			if res.Error != nil {
				errMsg := res.Error.Error()
				metrics.AppWBMediaOperationErrorsTotal.WithLabelValues("upload_file").Inc()
				log.Printf("Error uploading media file (photo %d) for nmID %d: %s", res.PhotoNumber, foundNmID, errMsg)
				if protoResp.ErrorMessage != nil && *protoResp.ErrorMessage != "" { // Append if marshalling error already occurred
					combined := *protoResp.ErrorMessage + "; " + errMsg
					protoResp.ErrorMessage = &combined
				} else {
					protoResp.ErrorMessage = &errMsg
				}
			}
			protoMediaUploadResponses = append(protoMediaUploadResponses, protoResp)
		}
	}

	// Handle save media by links
	if len(req.GetWbMediaToSaveLinks()) > 0 {
		metrics.AppWBMediaOperationsTotal.WithLabelValues("save_by_link").Inc()
		log.Printf("Attempting to save %d media links to Wildberries for nmID %d.", len(req.GetWbMediaToSaveLinks()), foundNmID)
		payload := entities.WBSaveMediaPayload{
			NmID: foundNmID,
			Data: req.WbMediaToSaveLinks,
		}
		saveResp, saveErr := wbs.wbClient.SaveMediaByLinks(ctx, apiKey, payload)

		protoSaveResp := &entities.WbMediaSaveByLinksResponse{}
		if saveResp != nil {
			if saveResp.Error {
				metrics.AppWBMediaOperationErrorsTotal.WithLabelValues("save_by_link").Inc()
			}
			jsonBytes, jsonErr := json.Marshal(saveResp)
			if jsonErr != nil {
				log.Printf("Error marshalling WB media save by links response: %v", jsonErr)
				errMsg := fmt.Sprintf("Failed to marshal WB media save by links response: %v", jsonErr)
				protoSaveResp.ErrorMessage = &errMsg
			} else {
				s := string(jsonBytes)
				protoSaveResp.ResponseJson = &s
			}
		}
		if saveErr != nil {
			errMsg := saveErr.Error()
			metrics.AppWBMediaOperationErrorsTotal.WithLabelValues("save_by_link").Inc()
			log.Printf("Error during WB media save by links operation: %s", errMsg)
			if protoSaveResp.ErrorMessage != nil && *protoSaveResp.ErrorMessage != "" {
				combined := *protoSaveResp.ErrorMessage + "; " + errMsg
				protoSaveResp.ErrorMessage = &combined
			} else {
				protoSaveResp.ErrorMessage = &errMsg
			}
		}
		if protoSaveResp.ResponseJson != nil || protoSaveResp.ErrorMessage != nil {
			protoMediaSaveResponse = protoSaveResp
		}
	}

	return protoMediaUploadResponses, protoMediaSaveResponse, nil
}

func (wbs *WbService) CreateCard(ctx context.Context, req *entities.ProductCard, aiGeneretedContent *entities.CardCraftAiGeneratedContent) (*string, *string, *bool, error) {
	var wbApiResponseJSON *string
	var wbPreparedRequestJSON *string
	var wbRequestAttempted *bool

	// Safely prepare WBDimensions, defaulting to zero values if request dimensions are nil.
	wbDimensions := entities.WBDimensions{}
	if req.Dimensions != nil {
		wbDimensions.Length = req.Dimensions.Length
		wbDimensions.Width = req.Dimensions.Width
		wbDimensions.Height = req.Dimensions.Height
		wbDimensions.WeightBrutto = req.Dimensions.WeightBrutto
	}

	// Prepare WB payload (common for all WB-related scenarios)
	wbVariant := entities.WBVariant{
		VendorCode:  req.VendorCode,
		Brand:       req.Brand,
		Title:       aiGeneretedContent.Title,       // Use Title from CardCraftAiAPIResponse
		Description: aiGeneretedContent.Description, // Use Description from CardCraftAiAPIResponse
		Dimensions:  wbDimensions,
		Sizes:       make([]entities.WBSize, len(req.Sizes)),
	}
	for i, s := range req.Sizes {
		// Determine price for WildBerries - prefer WB-specific price, fallback to general price
		var price int
		if s.WbPrice != nil {
			price = int(*s.WbPrice)
		} else {
			price = s.Price
		}

		wbVariant.Sizes[i] = entities.WBSize{
			TechSize: s.TechSize,
			WbSize:   s.WbSize,
			Price:    price,
			Skus:     s.Skus,
		}
	}

	var subjectIDValue int32
	if aiGeneretedContent.SubjectID != nil {
		subjectIDValue = *aiGeneretedContent.SubjectID
	} else {
		// If SubjectID is nil, it's a problem for WB. Log it.
		// The WB API call will likely fail or use 0. This fix prevents the panic.
		log.Printf("Error in handleWildberriesIntegration: CardCraftAI API response has a nil SubjectID. WB card creation might fail or use 0 for SubjectID.")
		// subjectIDValue remains 0 by default for int.
	}
	wbPayload := entities.WBCardUploadPayload{entities.WBCardRequestItem{
		SubjectID: subjectIDValue,
		Variants:  []entities.WBVariant{wbVariant},
	}}

	// Determine if an actual API call to Wildberries will be attempted
	attemptAPICall := req.GetWb() && req.GetWbApiKey() != ""
	wbRequestAttempted = &attemptAPICall

	if attemptAPICall {
		// Scenario: wb=true AND API key is provided. Make the API call.
		log.Printf("Attempting to upload card to Wildberries with provided API key.")
		wbResp, wbUploadErr := wbs.wbClient.UploadWBCard(ctx, wbPayload, req.GetWbApiKey())
		var responseStringToStore string

		if wbUploadErr != nil {
			log.Printf("Error uploading card to Wildberries: %v", wbUploadErr)
			metrics.AppExternalAPIErrorsTotal.WithLabelValues("wb_card_upload").Inc()
			errorResponse := entities.WBCardUploadResponse{
				Error:     true,
				ErrorText: wbUploadErr.Error(),
			}
			errBytes, err := json.Marshal(errorResponse)
			if err != nil {
				log.Printf("Error marshalling WB client error response: %v", err)
				responseStringToStore = fmt.Sprintf("{\"error\":true,\"errorText\":\"Client error uploading to WB and failed to marshal error: %s\"}", wbUploadErr.Error())
			} else {
				responseStringToStore = string(errBytes)
			}
			wbApiResponseJSON = &responseStringToStore
			return wbApiResponseJSON, wbPreparedRequestJSON, wbRequestAttempted, fmt.Errorf("WB card upload failed: %w", wbUploadErr)
		} else {
			log.Printf("Successfully called Wildberries API. Response received.")
			if wbResp != nil && wbResp.Error {
				metrics.AppExternalAPIErrorsTotal.WithLabelValues("wb_card_upload").Inc()
				respBytes, marshalErr := json.Marshal(wbResp)
				if marshalErr != nil {
					log.Printf("Error marshalling WB API response: %v", marshalErr)
					errorText := "Failed to marshal WB API response."
					if wbResp.ErrorText != "" {
						errorText = wbResp.ErrorText
					}
					responseStringToStore = fmt.Sprintf("{\"error\":true,\"errorText\":\"%s\",\"additionalErrors\":\"Marshalling of WB response failed: %s\"}", errorText, marshalErr.Error())
				} else {
					responseStringToStore = string(respBytes)
				}
				wbApiResponseJSON = &responseStringToStore
				return wbApiResponseJSON, wbPreparedRequestJSON, wbRequestAttempted, fmt.Errorf("WB API returned error: %s", wbResp.ErrorText)
			}
			respBytes, marshalErr := json.Marshal(wbResp)
			if marshalErr != nil {
				log.Printf("Error marshalling WB API response: %v", marshalErr)
				errorText := "Failed to marshal WB API response."
				if wbResp != nil && wbResp.ErrorText != "" {
					errorText = wbResp.ErrorText
				}
				responseStringToStore = fmt.Sprintf("{\"error\":true,\"errorText\":\"%s\",\"additionalErrors\":\"Marshalling of WB response failed: %s\"}", errorText, marshalErr.Error())
			} else {
				responseStringToStore = string(respBytes)
			}
		}
		wbApiResponseJSON = &responseStringToStore
	} else {
		// Scenario: API call will NOT be made.
		// This happens if wb=false, OR if wb=true but no API key is provided.
		// In this case, populate wb_prepared_request_json.

		if req.GetWb() { // wb=true, but API key was empty (handled by !attemptAPICall)
			log.Printf("wb=true, but API key not provided. Populating wb_prepared_request_json.")
		} else { // wb=false
			log.Printf("wb=false. Populating wb_prepared_request_json.")
		}

		preparedBytes, err := json.Marshal(wbPayload)
		if err != nil {
			log.Printf("Error marshalling WB prepared request: %v", err)
			errMsg := fmt.Sprintf("{\"error\":true,\"errorText\":\"Failed to marshal prepared WB request: %s\"}", err.Error())
			wbPreparedRequestJSON = &errMsg
		} else {
			jsonStr := string(preparedBytes)
			wbPreparedRequestJSON = &jsonStr
		}
		// No API call made, so API response JSON is empty.
		emptyStr := ""
		wbApiResponseJSON = &emptyStr
	}

	return wbApiResponseJSON, wbPreparedRequestJSON, wbRequestAttempted, nil
}
