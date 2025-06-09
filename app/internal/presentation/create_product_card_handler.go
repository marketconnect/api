package presentation

import (
	"api/app/domain/entities"
	apiv1 "api/gen/api/v1"
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
)

type CreateCardUsecase interface {
	CreateProductCard(ctx context.Context, apiKey string, req entities.ProductCard) (*entities.CreateProductCardResult, error)
}

type CreateProductCardHandler struct {
	createCardUsecase CreateCardUsecase
}

func NewCreateProductCardHandler(createCardUsecase CreateCardUsecase) *CreateProductCardHandler {
	return &CreateProductCardHandler{
		createCardUsecase: createCardUsecase,
	}
}

func (h *CreateProductCardHandler) Create(ctx context.Context, req *connect.Request[apiv1.CreateRequest]) (*connect.Response[apiv1.CreateResponse], error) {
	return h.CreateProductCard(ctx, req)
}

func (h *CreateProductCardHandler) CreateProductCard(ctx context.Context, req *connect.Request[apiv1.CreateRequest]) (*connect.Response[apiv1.CreateResponse], error) {
	log.Printf("CreateProductCard request: %+v", req.Msg)

	// Extract API key from Authorization header
	apiKey, err := ExtractAPIKeyFromHeader(req.Header())
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if req.Msg.GetWb() && req.Msg.GetVendorCode() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("vendor_code is required when wb is true"))
	}

	// Validate Ozon required fields
	if req.Msg.GetOzon() {
		// Check for required API credentials
		if req.Msg.GetOzonApiKey() == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ozon_api_key is required when ozon is true"))
		}
		if req.Msg.GetOzonApiClientId() == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ozon_api_client_id is required when ozon is true"))
		}

		// Check for vendor_code
		if req.Msg.GetVendorCode() == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("vendor_code is required when ozon is true"))
		}

		// Check for dimensions - must be present and non-zero
		if req.Msg.Dimensions == nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("dimensions (length, width, height, weight_brutto) are required when ozon is true"))
		}
		if req.Msg.Dimensions.Length <= 0 || req.Msg.Dimensions.Width <= 0 || req.Msg.Dimensions.Height <= 0 || req.Msg.Dimensions.WeightBrutto <= 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("dimensions (length, width, height, weight_brutto) must be greater than zero when ozon is true"))
		}

		// Check for price in sizes
		// if len(req.Msg.Sizes) == 0 || req.Msg.Sizes[0].Price <= 0 {
		// 	return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("price in sizes[0] is required and must be greater than zero when ozon is true"))
		// }
	}

	sizes := make([]*entities.WBSize, len(req.Msg.Sizes))
	for i, s := range req.Msg.Sizes {
		sizes[i] = &entities.WBSize{
			TechSize: s.TechSize,
			WbSize:   s.WbSize,
			Price:    int(s.Price),
			Skus:     s.Skus,
		}
	}

	wbMediaToUploadFiles := make([]*entities.WBClientMediaFile, len(req.Msg.WbMediaToUploadFiles))
	for i, f := range req.Msg.WbMediaToUploadFiles {
		wbMediaToUploadFiles[i] = &entities.WBClientMediaFile{
			Content:     f.Content,
			Filename:    f.Filename,
			PhotoNumber: f.PhotoNumber,
		}
	}

	productCard := entities.ProductCard{
		ProductTitle:         req.Msg.ProductTitle,
		ProductDescription:   req.Msg.ProductDescription,
		ParentId:             req.Msg.ParentId,
		SubjectId:            req.Msg.SubjectId,
		RootId:               req.Msg.RootId,
		SubId:                req.Msg.SubId,
		TypeId:               req.Msg.TypeId,
		GenerateContent:      req.Msg.GetGenerateContent(),
		Ozon:                 req.Msg.GetOzon(),
		Wb:                   req.Msg.GetWb(),
		Translate:            req.Msg.GetTranslate(),
		VendorCode:           req.Msg.VendorCode,
		Dimensions:           createDimensions(req.Msg.Dimensions),
		Brand:                req.Msg.Brand,
		Sizes:                sizes,
		WbApiKey:             req.Msg.WbApiKey,
		WbMediaToUploadFiles: wbMediaToUploadFiles,
		WbMediaToSaveLinks:   req.Msg.WbMediaToSaveLinks,
		OzonApiClientId:      req.Msg.OzonApiClientId,
		OzonApiKey:           req.Msg.OzonApiKey,
	}

	createProductCardResult, err := h.createCardUsecase.CreateProductCard(ctx, apiKey, productCard)
	if err != nil {
		return nil, err
	}

	// Check if CardCraftAiGeneratedContent is nil (defensive programming)
	if createProductCardResult.CardCraftAiGeneratedContent == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("card craft AI generated content is nil"))
	}

	wbMediaUploadIndividualResponses := make([]*apiv1.WBMediaUploadIndividualResponse, len(createProductCardResult.WbMediaUploadResponses))
	for i, response := range createProductCardResult.WbMediaUploadResponses {
		wbMediaUploadIndividualResponses[i] = &apiv1.WBMediaUploadIndividualResponse{
			PhotoNumber:  response.PhotoNumber,
			ResponseJson: response.ResponseJson,
			ErrorMessage: response.ErrorMessage,
		}
	}

	wbMediaSaveByLinksResponse := &apiv1.WBMediaSaveByLinksResponse{
		ResponseJson: createProductCardResult.WbMediaSaveResponse.ResponseJson,
		ErrorMessage: createProductCardResult.WbMediaSaveResponse.ErrorMessage,
	}

	createProductCardResponse := &apiv1.CreateResponse{
		Title:                            createProductCardResult.CardCraftAiGeneratedContent.Title,
		Description:                      createProductCardResult.CardCraftAiGeneratedContent.Description,
		Attributes:                       createProductCardResult.CardCraftAiGeneratedContent.Attributes,
		WbApiResponseJson:                createProductCardResult.WbApiResponseJson,
		WbPreparedRequestJson:            createProductCardResult.WbPreparedRequestJson,
		WbRequestAttempted:               createProductCardResult.WbRequestAttempted,
		WbMediaUploadIndividualResponses: wbMediaUploadIndividualResponses,
		WbMediaSaveByLinksResponse:       wbMediaSaveByLinksResponse,
		OzonApiResponseJson:              createProductCardResult.OzonApiResponseJson,
		OzonRequestAttempted:             createProductCardResult.OzonRequestAttempted,
	}

	// Safely handle pointer fields with nil checks
	if createProductCardResult.CardCraftAiGeneratedContent.ParentID != nil {
		createProductCardResponse.ParentId = *createProductCardResult.CardCraftAiGeneratedContent.ParentID
	}
	if createProductCardResult.CardCraftAiGeneratedContent.SubjectID != nil {
		createProductCardResponse.SubjectId = *createProductCardResult.CardCraftAiGeneratedContent.SubjectID
	}
	if createProductCardResult.CardCraftAiGeneratedContent.TypeID != nil {
		createProductCardResponse.TypeId = *createProductCardResult.CardCraftAiGeneratedContent.TypeID
	}
	if createProductCardResult.CardCraftAiGeneratedContent.RootID != nil {
		createProductCardResponse.RootId = *createProductCardResult.CardCraftAiGeneratedContent.RootID
	}
	if createProductCardResult.CardCraftAiGeneratedContent.SubID != nil {
		createProductCardResponse.SubId = *createProductCardResult.CardCraftAiGeneratedContent.SubID
	}

	// Set the optional string fields
	if createProductCardResult.CardCraftAiGeneratedContent.ParentName != nil {
		createProductCardResponse.ParentName = *createProductCardResult.CardCraftAiGeneratedContent.ParentName
	}
	if createProductCardResult.CardCraftAiGeneratedContent.SubjectName != nil {
		createProductCardResponse.SubjectName = *createProductCardResult.CardCraftAiGeneratedContent.SubjectName
	}
	if createProductCardResult.CardCraftAiGeneratedContent.TypeName != nil {
		createProductCardResponse.TypeName = *createProductCardResult.CardCraftAiGeneratedContent.TypeName
	}
	if createProductCardResult.CardCraftAiGeneratedContent.RootName != nil {
		createProductCardResponse.RootName = *createProductCardResult.CardCraftAiGeneratedContent.RootName
	}
	if createProductCardResult.CardCraftAiGeneratedContent.SubName != nil {
		createProductCardResponse.SubName = *createProductCardResult.CardCraftAiGeneratedContent.SubName
	}

	return &connect.Response[apiv1.CreateResponse]{
		Msg: createProductCardResponse,
	}, nil
}

// createDimensions safely creates WBDimensions handling nil input
func createDimensions(dims *apiv1.Dimensions) *entities.WBDimensions {
	if dims == nil {
		return &entities.WBDimensions{
			Length:       nil,
			Width:        nil,
			Height:       nil,
			WeightBrutto: nil,
		}
	}
	return &entities.WBDimensions{
		Length:       &dims.Length,
		Width:        &dims.Width,
		Height:       &dims.Height,
		WeightBrutto: &dims.WeightBrutto,
	}
}
