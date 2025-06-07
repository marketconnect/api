package entities

type CreateProductCardResult struct {
	CardCraftAiGeneratedContent *CardCraftAiGeneratedContent
	OzonApiResponseJson         *string
	OzonRequestAttempted        *bool
	WbApiResponseJson           *string
	WbPreparedRequestJson       *string
	WbRequestAttempted          *bool
	WbMediaUploadResponses      []*WbMediaUploadIndividualResponse
	WbMediaSaveResponse         *WbMediaSaveByLinksResponse
}

type WbMediaUploadIndividualResponse struct {
	PhotoNumber  int32
	ResponseJson *string
	ErrorMessage *string
}

type WbMediaSaveByLinksResponse struct {
	ResponseJson *string
	ErrorMessage *string
}
