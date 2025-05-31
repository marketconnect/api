package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"api/app/domain/entities"
	apiv1 "api/gen/api/v1"
	"log"

	"connectrpc.com/connect"
)

type cardCraftAiClient interface {
	GetCardContent(ctx context.Context, sessionID string, cardCraftAiAPIRequest entities.CardCraftAiAPIRequest) (*entities.CardCraftAiAPIResponse, error)
	GetSessionID(ctx context.Context) (string, error)
}

type wbClient interface {
	UploadWBCard(ctx context.Context, wbPayload entities.WBCardUploadPayload, apiKey string) (*entities.WBCardUploadResponse, error)
	UploadMediaFiles(ctx context.Context, apiKey string, nmID string, files []entities.WBClientMediaFile) ([]entities.WBMediaUploadResult, error)
	SaveMediaByLinks(ctx context.Context, apiKey string, payload entities.WBSaveMediaPayload) (*entities.WBMediaGenericResponse, error)
	GetCardList(ctx context.Context, apiKey string, listReq entities.WBGetCardListRequest) (*entities.WBGetCardListResponse, error)
}

type CreateCardUsecase struct {
	cardCraftAiClient           cardCraftAiClient
	wbClient                    wbClient
	wbApiGetCardListMaxAttempts int
}

func NewCreateCardUsecase(cardCraftAiClient cardCraftAiClient, wbClient wbClient, wbApiGetCardListMaxAttempts int) *CreateCardUsecase {
	return &CreateCardUsecase{
		cardCraftAiClient:           cardCraftAiClient,
		wbClient:                    wbClient,
		wbApiGetCardListMaxAttempts: wbApiGetCardListMaxAttempts,
	}
}

func (uc *CreateCardUsecase) CreateProductCard(ctx context.Context, req *connect.Request[apiv1.CreateProductCardRequest]) (*connect.Response[apiv1.CreateProductCardResponse], error) {
	// Validate request: if wb is true, vendor_code must be provided.
	if req.Msg.GetWb() && req.Msg.GetVendorCode() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("vendor_code is required when wb is true"))
	}

	sessionID, err := uc.cardCraftAiClient.GetSessionID(ctx)
	if err != nil {
		return nil, err
	}

	cardCraftAiAPIRequest, err := uc.mapRequestToDomain(req)
	if err != nil {
		return nil, err
	}

	cardCraftAiAPIResponse, err := uc.cardCraftAiClient.GetCardContent(ctx, sessionID, cardCraftAiAPIRequest)
	if err != nil {
		return nil, err
	}

	wbApiResponseJSON, wbPreparedRequestJSON, wbRequestAttempted := uc.handleWildberriesIntegration(ctx, req, cardCraftAiAPIResponse)

	// Handle media uploads and saves
	wbMediaUploadResponses, wbMediaSaveResponse, mediaErr := uc.handleWildberriesMediaOperations(ctx, req)
	if mediaErr != nil {
		log.Printf("Error in Wildberries media operations: %v", mediaErr)
		return nil, mediaErr
	}

	// Map CardCraftAiAPIResponse to CreateProductCardResponse
	createProductCardResponse := &apiv1.CreateProductCardResponse{
		Title:                            cardCraftAiAPIResponse.Title,
		Description:                      cardCraftAiAPIResponse.Description,
		Attributes:                       cardCraftAiAPIResponse.Attributes,
		WbApiResponseJson:                wbApiResponseJSON,
		WbPreparedRequestJson:            wbPreparedRequestJSON,
		WbRequestAttempted:               wbRequestAttempted,
		WbMediaUploadIndividualResponses: wbMediaUploadResponses,
		WbMediaSaveByLinksResponse:       wbMediaSaveResponse,
	}

	// Safely handle pointer fields with nil checks
	if cardCraftAiAPIResponse.ParentID != nil {
		createProductCardResponse.ParentId = *cardCraftAiAPIResponse.ParentID
	}
	if cardCraftAiAPIResponse.SubjectID != nil {
		createProductCardResponse.SubjectId = *cardCraftAiAPIResponse.SubjectID
	}
	if cardCraftAiAPIResponse.TypeID != nil {
		createProductCardResponse.TypeId = *cardCraftAiAPIResponse.TypeID
	}
	if cardCraftAiAPIResponse.RootID != nil {
		createProductCardResponse.RootId = *cardCraftAiAPIResponse.RootID
	}
	if cardCraftAiAPIResponse.SubID != nil {
		createProductCardResponse.SubId = *cardCraftAiAPIResponse.SubID
	}

	// Set the optional string fields
	if cardCraftAiAPIResponse.ParentName != nil {
		createProductCardResponse.ParentName = *cardCraftAiAPIResponse.ParentName
	}
	if cardCraftAiAPIResponse.SubjectName != nil {
		createProductCardResponse.SubjectName = *cardCraftAiAPIResponse.SubjectName
	}
	if cardCraftAiAPIResponse.TypeName != nil {
		createProductCardResponse.TypeName = *cardCraftAiAPIResponse.TypeName
	}
	if cardCraftAiAPIResponse.RootName != nil {
		createProductCardResponse.RootName = *cardCraftAiAPIResponse.RootName
	}
	if cardCraftAiAPIResponse.SubName != nil {
		createProductCardResponse.SubName = *cardCraftAiAPIResponse.SubName
	}

	return connect.NewResponse(createProductCardResponse), nil
}

func (uc *CreateCardUsecase) mapRequestToDomain(req *connect.Request[apiv1.CreateProductCardRequest]) (entities.CardCraftAiAPIRequest, error) {
	var domainSizes []entities.Size
	if req.Msg.Sizes != nil {
		domainSizes = make([]entities.Size, len(req.Msg.Sizes))
		for i, s := range req.Msg.Sizes {
			domainSizes[i] = entities.Size{
				TechSize: s.TechSize,
				WbSize:   s.WbSize,
				Price:    int(s.Price),
				Skus:     s.Skus,
			}
		}
	}

	return entities.CardCraftAiAPIRequest{
		ProductTitle:       req.Msg.ProductTitle,
		ProductDescription: req.Msg.ProductDescription,
		ParentID:           req.Msg.ParentId,
		SubjectID:          req.Msg.SubjectId,
		Translate:          req.Msg.GetTranslate(),
		Ozon:               req.Msg.GetOzon(),
		GenerateContent:    req.Msg.GetGenerateContent(),
	}, nil
}

func (uc *CreateCardUsecase) handleWildberriesIntegration(ctx context.Context, req *connect.Request[apiv1.CreateProductCardRequest], ccaApiResponse *entities.CardCraftAiAPIResponse) (*string, *string, *bool) {
	var wbApiResponseJSON *string
	var wbPreparedRequestJSON *string
	var wbRequestAttempted *bool

	// Safely prepare WBDimensions, defaulting to zero values if request dimensions are nil.
	wbDimensions := entities.WBDimensions{}
	if req.Msg.Dimensions != nil {
		wbDimensions.Length = int(req.Msg.Dimensions.Length)
		wbDimensions.Width = int(req.Msg.Dimensions.Width)
		wbDimensions.Height = int(req.Msg.Dimensions.Height)
		wbDimensions.WeightBrutto = float64(req.Msg.Dimensions.WeightBrutto)
	}

	// Prepare WB payload (common for all WB-related scenarios)
	wbVariant := entities.WBVariant{
		VendorCode:  req.Msg.VendorCode,
		Brand:       req.Msg.Brand,
		Title:       ccaApiResponse.Title,       // Use Title from CardCraftAiAPIResponse
		Description: ccaApiResponse.Description, // Use Description from CardCraftAiAPIResponse
		Dimensions:  wbDimensions,
		Sizes:       make([]entities.WBSize, len(req.Msg.Sizes)),
	}
	for i, s := range req.Msg.Sizes {
		wbVariant.Sizes[i] = entities.WBSize{
			TechSize: s.TechSize,
			WbSize:   s.WbSize,
			Price:    int(s.Price),
			Skus:     s.Skus,
		}
	}

	var subjectIDValue int
	if ccaApiResponse.SubjectID != nil {
		subjectIDValue = int(*ccaApiResponse.SubjectID)
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
	attemptAPICall := req.Msg.GetWb() && req.Msg.GetWbApiKey() != ""
	wbRequestAttempted = &attemptAPICall

	if attemptAPICall {
		// Scenario: wb=true AND API key is provided. Make the API call.
		log.Printf("Attempting to upload card to Wildberries with provided API key.")
		wbResp, wbUploadErr := uc.wbClient.UploadWBCard(ctx, wbPayload, req.Msg.GetWbApiKey())
		var responseStringToStore string

		if wbUploadErr != nil {
			log.Printf("Error uploading card to Wildberries: %v", wbUploadErr)
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
		} else {
			log.Printf("Successfully called Wildberries API. Response: %+v", wbResp)
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

		if req.Msg.GetWb() { // wb=true, but API key was empty (handled by !attemptAPICall)
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

	return wbApiResponseJSON, wbPreparedRequestJSON, wbRequestAttempted
}

func (uc *CreateCardUsecase) handleWildberriesMediaOperations(ctx context.Context, req *connect.Request[apiv1.CreateProductCardRequest]) ([]*apiv1.WBMediaUploadIndividualResponse, *apiv1.WBMediaSaveByLinksResponse, error) {
	var protoMediaUploadResponses []*apiv1.WBMediaUploadIndividualResponse
	var protoMediaSaveResponse *apiv1.WBMediaSaveByLinksResponse

	// If no media operations are requested, return early.
	if len(req.Msg.GetWbMediaToUploadFiles()) == 0 && len(req.Msg.GetWbMediaToSaveLinks()) == 0 {
		return nil, nil, nil
	}

	apiKey := req.Msg.GetWbApiKey()
	vendorCode := req.Msg.GetVendorCode()

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

	for attempt := 1; attempt <= uc.wbApiGetCardListMaxAttempts; attempt++ {
		log.Printf("Attempt %d to find nmID for vendor code %s", attempt, vendorCode)
		filterWithPhoto := 0 // Only cards with no photos
		listReqPayload := entities.WBGetCardListRequest{
			Settings: entities.WBGetCardListRequestSettings{
				Filter: &entities.WBGetCardListRequestFilter{
					WithPhoto: &filterWithPhoto,
				},
			},
		}

		wbCardsResp, err := uc.wbClient.GetCardList(ctx, apiKey, listReqPayload)
		if err != nil {
			log.Printf("Error getting card list from WB (attempt %d) for vendor code %s: %v", attempt, vendorCode, err)
			if attempt == uc.wbApiGetCardListMaxAttempts {
				return nil, nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to get card list from WB after %d attempts for vendor code %s: %w", uc.wbApiGetCardListMaxAttempts, vendorCode, err))
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

		if attempt < uc.wbApiGetCardListMaxAttempts {
			log.Printf("Card with vendor code %s not found in attempt %d. Retrying in %v...", vendorCode, attempt, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if foundNmID == 0 {
		log.Printf("Failed to find nmID for vendor code %s after %d attempts.", vendorCode, uc.wbApiGetCardListMaxAttempts)
		return nil, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("card with vendor code '%s' not found on Wildberries after %d attempts", vendorCode, uc.wbApiGetCardListMaxAttempts))
	}

	// Handle file uploads
	if len(req.Msg.GetWbMediaToUploadFiles()) > 0 {
		log.Printf("Attempting to upload %d media files to Wildberries for nmID %d.", len(req.Msg.GetWbMediaToUploadFiles()), foundNmID)
		clientMediaFiles := make([]entities.WBClientMediaFile, len(req.Msg.WbMediaToUploadFiles))
		for i, f := range req.Msg.WbMediaToUploadFiles {
			clientMediaFiles[i] = entities.WBClientMediaFile{
				Filename:    f.Filename,
				Content:     f.Content,
				PhotoNumber: f.PhotoNumber,
			}
		}

		uploadResults, err := uc.wbClient.UploadMediaFiles(ctx, apiKey, fmt.Sprintf("%d", foundNmID), clientMediaFiles)
		if err != nil {
			// This error is for the whole operation, e.g., context cancellation before starting.
			log.Printf("Overall error calling UploadMediaFiles for nmID %d: %v", foundNmID, err)
			// We might still have partial results in uploadResults, or it might be nil.
			// For now, we'll process any results returned.
		}

		for _, res := range uploadResults {
			protoResp := &apiv1.WBMediaUploadIndividualResponse{
				PhotoNumber: res.PhotoNumber,
			}
			if res.Response != nil {
				jsonBytes, jsonErr := json.Marshal(res.Response)
				if jsonErr != nil {
					log.Printf("Error marshalling WB media upload response for photo %d, nmID %d: %v", res.PhotoNumber, foundNmID, jsonErr)
					errMsg := fmt.Sprintf("Failed to marshal WB media upload response: %v", jsonErr)
					protoResp.ErrorMessage = &errMsg
				} else {
					s := string(jsonBytes)
					protoResp.ResponseJson = &s
				}
			}
			if res.Error != nil {
				errMsg := res.Error.Error()
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
	if len(req.Msg.GetWbMediaToSaveLinks()) > 0 {
		log.Printf("Attempting to save %d media links to Wildberries for nmID %d.", len(req.Msg.GetWbMediaToSaveLinks()), foundNmID)
		payload := entities.WBSaveMediaPayload{
			NmID: foundNmID,
			Data: req.Msg.WbMediaToSaveLinks,
		}
		saveResp, saveErr := uc.wbClient.SaveMediaByLinks(ctx, apiKey, payload)

		protoSaveResp := &apiv1.WBMediaSaveByLinksResponse{}
		if saveResp != nil {
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
