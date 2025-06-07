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

	// Create card in Wildberries
	wbApiResponseJSON, wbPreparedRequestJSON, wbRequestAttempted, wbErr := uc.wbService.CreateCard(ctx, &req, cardCraftAiGeneratedContent)
	createProductCardResult.WbApiResponseJson = wbApiResponseJSON
	createProductCardResult.WbPreparedRequestJson = wbPreparedRequestJSON
	createProductCardResult.WbRequestAttempted = wbRequestAttempted

	// Create card in Ozon
	ozonApiResponseJSON, ozonRequestAttempted, ozonErr := uc.ozonService.CreateCard(ctx, &req, cardCraftAiGeneratedContent)
	createProductCardResult.OzonApiResponseJson = ozonApiResponseJSON
	createProductCardResult.OzonRequestAttempted = ozonRequestAttempted

	// Log errors but don't stop execution (marketplace integrations are independent)
	if wbErr != nil {
		log.Printf("Error in Wildberries card creation: %v", wbErr)
	}
	if ozonErr != nil {
		log.Printf("Error in Ozon card creation: %v", ozonErr)
	}

	// Handle media uploads and saves - only if WB card creation was attempted and successful
	var shouldAttemptMedia bool = false
	if wbRequestAttempted != nil && *wbRequestAttempted && wbErr == nil {
		shouldAttemptMedia = true
		log.Printf("WB card creation successful, proceeding with media operations")
	} else if wbRequestAttempted != nil && *wbRequestAttempted && wbErr != nil {
		log.Printf("WB card creation failed, skipping media operations: %v", wbErr)
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
