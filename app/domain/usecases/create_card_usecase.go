package usecases

import (
	"context"

	"api/app/domain/entities"
	"api/metrics" // For accessing Prometheus metrics
	"log"
)

type wbService interface {
	CreateCard(ctx context.Context, req *entities.ProductCard, ccaApiResponse *entities.CardCraftAiGeneratedContent) (*string, *string, *bool, error)
	AddMedia(ctx context.Context, req *entities.ProductCard) ([]*entities.WbMediaUploadIndividualResponse, *entities.WbMediaSaveByLinksResponse, error)
}

type ozonService interface {
	CreateCard(ctx context.Context, req *entities.ProductCard, ccaApiResponse *entities.CardCraftAiGeneratedContent) (*string, *bool, error)
}

type cardCraftAiService interface {
	GetCardContent(ctx context.Context, cardCraftAiAPIRequest entities.ProductCard) (*entities.CardCraftAiGeneratedContent, error)
}

type tokenBillingService interface {
	UpdateBalanceForSession(ctx context.Context, apiKey, sessionID string) error
}

type CreateCardUsecase struct {
	cardCraftAiService  cardCraftAiService
	wbService           wbService
	ozonService         ozonService
	tokenBillingService tokenBillingService
}

func NewCreateCardUsecase(cardCraftAiService cardCraftAiService, wbService wbService, ozonService ozonService, tokenBillingService tokenBillingService) *CreateCardUsecase {
	return &CreateCardUsecase{
		cardCraftAiService:  cardCraftAiService,
		wbService:           wbService,
		ozonService:         ozonService,
		tokenBillingService: tokenBillingService,
	}
}

func (uc *CreateCardUsecase) CreateProductCard(ctx context.Context, apiKey string, req entities.ProductCard) (*entities.CreateProductCardResult, error) {

	var createProductCardResult entities.CreateProductCardResult

	// Generate content for the card (sujects and optionaly seo content: title, description, attributes)
	cardCraftAiGeneratedContent, err := uc.cardCraftAiService.GetCardContent(ctx, req)
	if err != nil {
		metrics.AppExternalAPIErrorsTotal.WithLabelValues("card_craft_ai_content").Inc()
		return nil, err
	}
	createProductCardResult.CardCraftAiGeneratedContent = cardCraftAiGeneratedContent

	if err := uc.tokenBillingService.UpdateBalanceForSession(ctx, apiKey, cardCraftAiGeneratedContent.SessionID); err != nil {
		log.Printf("failed to update balance: %v", err)
	}

	metrics.AppCardCreationsTotal.Inc() // Core content generation successful

	// Create cards in WB and Ozon in parallel since they are independent
	type wbResult struct {
		apiResponseJSON     *string
		preparedRequestJSON *string
		requestAttempted    *bool
		err                 error
	}

	type ozonResult struct {
		apiResponseJSON  *string
		requestAttempted *bool
		err              error
	}

	wbChan := make(chan wbResult, 1)
	ozonChan := make(chan ozonResult, 1)

	// Create card in Wildberries (parallel)
	go func() {
		wbApiResponseJSON, wbPreparedRequestJSON, wbRequestAttempted, wbErr := uc.wbService.CreateCard(ctx, &req, cardCraftAiGeneratedContent)
		wbChan <- wbResult{
			apiResponseJSON:     wbApiResponseJSON,
			preparedRequestJSON: wbPreparedRequestJSON,
			requestAttempted:    wbRequestAttempted,
			err:                 wbErr,
		}
	}()

	// Create card in Ozon (parallel)
	go func() {
		log.Printf("Starting Ozon card creation for product: %s", req.ProductTitle)
		log.Printf("Ozon enabled: %t, ClientID: %s, ApiKey length: %d", req.Ozon, req.OzonApiClientId, len(req.OzonApiKey))

		ozonApiResponseJSON, ozonRequestAttempted, ozonErr := uc.ozonService.CreateCard(ctx, &req, cardCraftAiGeneratedContent)

		log.Printf("Ozon card creation completed - attempted: %v, error: %v", ozonRequestAttempted, ozonErr)
		if ozonApiResponseJSON != nil {
			log.Printf("Ozon response JSON: %s", *ozonApiResponseJSON)
		} else {
			log.Printf("Ozon response JSON is nil")
		}

		ozonChan <- ozonResult{
			apiResponseJSON:  ozonApiResponseJSON,
			requestAttempted: ozonRequestAttempted,
			err:              ozonErr,
		}
	}()

	// Wait for both operations to complete
	wbRes := <-wbChan
	ozonRes := <-ozonChan

	// Set WB results
	createProductCardResult.WbApiResponseJson = wbRes.apiResponseJSON
	createProductCardResult.WbPreparedRequestJson = wbRes.preparedRequestJSON
	createProductCardResult.WbRequestAttempted = wbRes.requestAttempted

	// Set Ozon results
	createProductCardResult.OzonApiResponseJson = ozonRes.apiResponseJSON
	createProductCardResult.OzonRequestAttempted = ozonRes.requestAttempted

	// Log errors but don't stop execution (marketplace integrations are independent)
	if wbRes.err != nil {
		log.Printf("Error in Wildberries card creation: %v", wbRes.err)
	}
	if ozonRes.err != nil {
		log.Printf("Error in Ozon card creation: %v", ozonRes.err)
	}

	// Handle media uploads and saves - only if WB card creation was attempted and successful
	var shouldAttemptMedia bool = false
	if wbRes.requestAttempted != nil && *wbRes.requestAttempted && wbRes.err == nil {
		shouldAttemptMedia = true
		log.Printf("WB card creation successful, proceeding with media operations")
	} else if wbRes.requestAttempted != nil && *wbRes.requestAttempted && wbRes.err != nil {
		log.Printf("WB card creation failed, skipping media operations: %v", wbRes.err)
	} else {
		log.Printf("WB card creation was not attempted, skipping media operations")
	}

	if shouldAttemptMedia {
		wbMediaUploadResponses, wbMediaSaveResponse, mediaErr := uc.wbService.AddMedia(ctx, &req)
		if mediaErr != nil {
			log.Printf("Error in Wildberries media operations: %v", mediaErr)
			// Don't return error here, media operations are not critical for the main flow
		} else {
			// Handle upload responses - could be nil if no uploads were attempted
			for _, response := range wbMediaUploadResponses {
				createProductCardResult.WbMediaUploadResponses = append(createProductCardResult.WbMediaUploadResponses, &entities.WbMediaUploadIndividualResponse{
					PhotoNumber:  response.PhotoNumber,
					ResponseJson: response.ResponseJson,
					ErrorMessage: response.ErrorMessage,
				})
			}

			// Handle save response - could be nil if no save by links were attempted
			if wbMediaSaveResponse != nil {
				createProductCardResult.WbMediaSaveResponse = &entities.WbMediaSaveByLinksResponse{
					ResponseJson: wbMediaSaveResponse.ResponseJson,
					ErrorMessage: wbMediaSaveResponse.ErrorMessage,
				}
			} else {
				// Initialize with empty response if nil
				createProductCardResult.WbMediaSaveResponse = &entities.WbMediaSaveByLinksResponse{}
			}
		}
	}

	// Initialize empty responses when media operations are skipped or failed
	if !shouldAttemptMedia || createProductCardResult.WbMediaSaveResponse == nil {
		createProductCardResult.WbMediaSaveResponse = &entities.WbMediaSaveByLinksResponse{}
	}

	return &createProductCardResult, nil
}
