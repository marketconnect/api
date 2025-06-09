package services

import (
	"api/app/domain/entities"
	"api/metrics"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

type ozonClient interface {
	ImportProductsV3(ctx context.Context, clientID, apiKey string, request entities.OzonProductImportRequest) (*entities.OzonProductImportResponse, error)
}

type fileUploadService interface {
	UploadWBMediaFiles(ctx context.Context, wbFiles []*entities.WBClientMediaFile) ([]string, error)
}

type ozonService struct {
	ozonClient        ozonClient
	fileUploadService fileUploadService
}

func NewOzonService(ozonClient ozonClient, fileUploadService fileUploadService) *ozonService {
	return &ozonService{
		ozonClient:        ozonClient,
		fileUploadService: fileUploadService,
	}
}

func (ozs *ozonService) CreateCard(ctx context.Context, req *entities.ProductCard, ccaApiResponse *entities.CardCraftAiGeneratedContent) (*string, *bool, error) {
	var ozonApiResponseJSON *string
	var ozonRequestAttempted *bool

	log.Printf("[OZON DEBUG] Starting CreateCard - ProductTitle: %s", req.ProductTitle)
	log.Printf("[OZON DEBUG] Request flags - Ozon: %t, ApiKey present: %t, ClientID present: %t",
		req.GetOzon(), req.GetOzonApiKey() != "", req.GetOzonApiClientId() != "")

	attemptAPICall := req.GetOzon() && req.GetOzonApiKey() != "" && req.GetOzonApiClientId() != ""
	ozonRequestAttempted = &attemptAPICall

	log.Printf("[OZON DEBUG] Will attempt API call: %t", attemptAPICall)

	if !attemptAPICall {
		if req.GetOzon() {
			log.Printf("[OZON DEBUG] Ozon integration requested but API key or Client ID is missing. Skipping Ozon API call.")
			log.Printf("[OZON DEBUG] Missing - ClientID: %t, ApiKey: %t",
				req.GetOzonApiClientId() == "", req.GetOzonApiKey() == "")
		} else {
			log.Printf("[OZON DEBUG] Ozon integration not requested (ozon=false)")
		}
		// No API call will be made, so response JSON is empty.
		emptyStr := ""
		ozonApiResponseJSON = &emptyStr
		return ozonApiResponseJSON, ozonRequestAttempted, nil
	}

	log.Printf("[OZON DEBUG] Starting validation checks")

	// Validate required fields for Ozon
	if req.GetVendorCode() == "" {
		log.Printf("[OZON DEBUG] Validation failed: vendor_code is missing")
		errMsg := `{"error":true,"errorText":"vendor_code (for offer_id) is required for Ozon integration"}`
		ozonApiResponseJSON = &errMsg
		return ozonApiResponseJSON, ozonRequestAttempted, fmt.Errorf("vendor_code (for offer_id) is required for Ozon integration")
	}
	log.Printf("[OZON DEBUG] VendorCode validation passed: %s", req.GetVendorCode())

	if ccaApiResponse.Title == "" {
		log.Printf("[OZON DEBUG] Validation failed: CardCraftAI title is missing")
		errMsg := `{"error":true,"errorText":"CardCraftAI title (for name) is required for Ozon integration"}`
		ozonApiResponseJSON = &errMsg
		return ozonApiResponseJSON, ozonRequestAttempted, fmt.Errorf("CardCraftAI title (for name) is required for Ozon integration")
	}
	log.Printf("[OZON DEBUG] Title validation passed: %s", ccaApiResponse.Title)

	if ccaApiResponse.SubID == nil {
		log.Printf("[OZON DEBUG] Validation failed: CardCraftAI SubID is missing")
		errMsg := `{"error":true,"errorText":"CardCraftAI SubID (for Ozon description_category_id) is required for Ozon integration"}`
		ozonApiResponseJSON = &errMsg
		return ozonApiResponseJSON, ozonRequestAttempted, fmt.Errorf("CardCraftAI SubID (for Ozon description_category_id) is required for Ozon integration")
	}
	log.Printf("[OZON DEBUG] SubID validation passed: %d", *ccaApiResponse.SubID)

	if ccaApiResponse.TypeID == nil {
		log.Printf("[OZON DEBUG] Validation failed: CardCraftAI TypeID is missing")
		errMsg := `{"error":true,"errorText":"CardCraftAI TypeID (for Ozon type_id) is required for Ozon integration"}`
		ozonApiResponseJSON = &errMsg
		return ozonApiResponseJSON, ozonRequestAttempted, fmt.Errorf("CardCraftAI TypeID (for Ozon type_id) is required for Ozon integration")
	}
	log.Printf("[OZON DEBUG] TypeID validation passed: %d", *ccaApiResponse.TypeID)

	if req.Dimensions == nil || req.Dimensions.Depth == nil || req.Dimensions.Width == nil || req.Dimensions.Height == nil || req.Dimensions.Weight == nil {
		log.Printf("[OZON DEBUG] Validation failed: Ozon dimensions are incomplete")
		log.Printf("[OZON DEBUG] Dimensions present: %t", req.Dimensions != nil)
		if req.Dimensions != nil {
			log.Printf("[OZON DEBUG] Depth: %v, Width: %v, Height: %v, Weight: %v",
				req.Dimensions.Depth, req.Dimensions.Width, req.Dimensions.Height, req.Dimensions.Weight)
		}
		errMsg := `{"error":true,"errorText":"Dimensions (depth, width, height, weight) are required and must be non-zero for Ozon integration"}`
		ozonApiResponseJSON = &errMsg
		return ozonApiResponseJSON, ozonRequestAttempted, fmt.Errorf("dimensions (depth, width, height, weight) are required and must be non-zero for Ozon integration")
	}
	log.Printf("[OZON DEBUG] Dimensions validation passed: %dx%dx%d, weight: %d",
		*req.Dimensions.Depth, *req.Dimensions.Width, *req.Dimensions.Height, *req.Dimensions.Weight)

	log.Printf("[OZON DEBUG] All validations passed, processing images")

	// Process images for Ozon - combine uploaded files and existing links
	var ozonImageURLs []string

	// First, add any existing WbMediaToSaveLinks (already URL links)
	if len(req.GetWbMediaToSaveLinks()) > 0 {
		log.Printf("[OZON DEBUG] Using %d existing image links from WbMediaToSaveLinks", len(req.GetWbMediaToSaveLinks()))
		ozonImageURLs = append(ozonImageURLs, req.GetWbMediaToSaveLinks()...)
	}

	// Second, upload WbMediaToUploadFiles and get URLs
	if len(req.GetWbMediaToUploadFiles()) > 0 {
		log.Printf("[OZON DEBUG] Uploading %d files from WbMediaToUploadFiles to get URLs for Ozon", len(req.GetWbMediaToUploadFiles()))

		uploadedURLs, err := ozs.fileUploadService.UploadWBMediaFiles(ctx, req.GetWbMediaToUploadFiles())
		if err != nil {
			log.Printf("[OZON DEBUG] Warning: Failed to upload some files for Ozon: %v", err)
			// Continue with partial URLs if some uploads succeeded
		}

		if len(uploadedURLs) > 0 {
			log.Printf("[OZON DEBUG] Successfully uploaded %d files, adding URLs to Ozon images", len(uploadedURLs))
			ozonImageURLs = append(ozonImageURLs, uploadedURLs...)
		} else {
			log.Printf("[OZON DEBUG] No files were successfully uploaded")
		}
	}

	log.Printf("[OZON DEBUG] Total images for Ozon: %d", len(ozonImageURLs))
	if len(ozonImageURLs) == 0 {
		log.Printf("[OZON DEBUG] Warning: No images available for Ozon product")
	}

	log.Printf("[OZON DEBUG] Creating Ozon payload")

	// Determine price from sizes if available, otherwise use a default minimum price
	var price string = "100" // Minimum price 100 kopecks = 1 ruble
	if len(req.Sizes) > 0 && req.Sizes[0].Price > 0 {
		price = fmt.Sprintf("%d", req.Sizes[0].Price)
	}

	ozonItem := entities.OzonProductImportItem{
		Name:                  ccaApiResponse.Title,
		OfferID:               req.VendorCode,
		DescriptionCategoryID: int64(*ccaApiResponse.SubID),
		TypeID:                int64(*ccaApiResponse.TypeID),
		Price:                 price,
		Vat:                   "0",
		CurrencyCode:          "RUB",
		Depth:                 int32(*req.Dimensions.Depth),
		Width:                 int32(*req.Dimensions.Width),
		Height:                int32(*req.Dimensions.Height),
		DimensionUnit:         req.Dimensions.DimensionUnit,
		Weight:                int32(*req.Dimensions.Weight),
		WeightUnit:            req.Dimensions.WeightUnit,
		Images:                ozonImageURLs, // Use processed image URLs
		Attributes:            []entities.OzonProductAttribute{},
		ComplexAttributes:     []entities.OzonComplexAttribute{},
	}

	if len(req.Sizes) > 0 && len(req.Sizes[0].Skus) > 0 {
		ozonItem.Barcode = req.Sizes[0].Skus[0]
	}

	// Add required "Название модели" attribute (Model Name)
	ozonItem.Attributes = append(ozonItem.Attributes, entities.OzonProductAttribute{
		ID:        9048, // Required "Название модели (для объединения в одну карточку)"
		ComplexID: 0,
		Values:    []entities.OzonProductAttributeValue{{Value: ccaApiResponse.Title}}, // Use title as model name
	})

	if req.Brand != "" {
		ozonItem.Attributes = append(ozonItem.Attributes, entities.OzonProductAttribute{
			ID:        85, // Standard Ozon ID for "Brand"
			ComplexID: 0,
			Values:    []entities.OzonProductAttributeValue{{Value: req.Brand}},
		})
	}

	ozonPayload := entities.OzonProductImportRequest{Items: []entities.OzonProductImportItem{ozonItem}}

	log.Printf("Attempting to import product to Ozon with ClientID: %s", req.GetOzonApiClientId())
	ozonResp, ozonErr := ozs.ozonClient.ImportProductsV3(ctx, req.GetOzonApiClientId(), req.GetOzonApiKey(), ozonPayload)

	var responseStringToStore string
	if ozonErr != nil {
		log.Printf("Error importing product to Ozon: %v", ozonErr)
		metrics.AppExternalAPIErrorsTotal.WithLabelValues("ozon_product_import").Inc()
		// Use a generic error structure for the JSON string
		errorResponse := map[string]interface{}{"error": true, "errorText": ozonErr.Error()}
		errBytes, _ := json.Marshal(errorResponse) // Ignore marshalling error for error response
		responseStringToStore = string(errBytes)
		ozonApiResponseJSON = &responseStringToStore
		return ozonApiResponseJSON, ozonRequestAttempted, fmt.Errorf("ozon product import failed: %w", ozonErr)
	} else {
		log.Printf("Successfully called Ozon API. Response received.")
		// Ozon's v3/product/import response doesn't have a top-level error field like WB.
		// Errors are typically indicated by non-200 HTTP status, handled by the ozonClient.
		// If specific task-level errors need to be parsed from the response, that logic would go here.
		respBytes, _ := json.Marshal(ozonResp) // Ignore marshalling error for success response
		responseStringToStore = string(respBytes)
	}
	ozonApiResponseJSON = &responseStringToStore
	return ozonApiResponseJSON, ozonRequestAttempted, nil
}
