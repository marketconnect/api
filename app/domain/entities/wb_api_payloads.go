package entities

// WBDimensions represents product dimensions for the Wildberries API.
type WBDimensions struct {
	Length       *int32   `json:"length"`
	Width        *int32   `json:"width"`
	Height       *int32   `json:"height"`
	WeightBrutto *float64 `json:"weightBrutto"`

	// Ozon specific fields
	Depth         *int32 `json:"depth,omitempty"`         // Ozon depth (can be same as length)
	DimensionUnit string `json:"dimensionUnit,omitempty"` // Ozon dimension unit (e.g., "mm")
	Weight        *int32 `json:"weight,omitempty"`        // Ozon weight (can be same as weight_brutto)
	WeightUnit    string `json:"weightUnit,omitempty"`    // Ozon weight unit (e.g., "g")
}

// WBSize represents product size information for the Wildberries API.
type WBSize struct {
	TechSize string   `json:"techSize"`
	WbSize   string   `json:"wbSize,omitempty"`
	Price    int      `json:"price,omitempty"`
	Skus     []string `json:"skus"`
}

// WBCharacteristic represents a product characteristic for the Wildberries API.
type WBCharacteristic struct {
	ID    int         `json:"id"`
	Value interface{} `json:"value"`
}

// WBVariant represents a product variant for the Wildberries API.
type WBVariant struct {
	VendorCode      string             `json:"vendorCode"`
	Brand           string             `json:"brand,omitempty"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	Dimensions      WBDimensions       `json:"dimensions"`
	Sizes           []WBSize           `json:"sizes,omitempty"`
	Characteristics []WBCharacteristic `json:"characteristics,omitempty"`
}

// WBCardRequestItem represents a single item in the card creation request for Wildberries API.
type WBCardRequestItem struct {
	SubjectID int32       `json:"subjectID"`
	Variants  []WBVariant `json:"variants"`
}

// WBCardUploadPayload is the type for the entire request payload for Wildberries card upload.
type WBCardUploadPayload []WBCardRequestItem

// WBCardUploadResponse represents the response from the Wildberries card upload API.
type WBCardUploadResponse struct {
	Data             interface{} `json:"data"` // Can be null or an object
	Error            bool        `json:"error"`
	ErrorText        string      `json:"errorText"`
	AdditionalErrors interface{} `json:"additionalErrors"` // Can be object or string
}

// WBSaveMediaPayload is the request body for saving media by links.
type WBSaveMediaPayload struct {
	NmID int      `json:"nmId"`
	Data []string `json:"data"`
}

// WBMediaGenericResponse is a generic response structure for WB media operations.
// It mirrors WBCardUploadResponse structure as per OpenAPI spec.
type WBMediaGenericResponse struct {
	Data             interface{} `json:"data"`
	Error            bool        `json:"error"`
	ErrorText        string      `json:"errorText"`
	AdditionalErrors interface{} `json:"additionalErrors"`
}

// WBClientMediaFile represents a file to be uploaded by the WBClient.
type WBClientMediaFile struct {
	Filename    string
	Content     []byte
	PhotoNumber int32 // Corresponds to X-Photo-Number
}

// WBMediaUploadResult holds the outcome of a single file upload attempt.
type WBMediaUploadResult struct {
	PhotoNumber int32 // The X-Photo-Number used for this upload
	Response    *WBMediaGenericResponse
	Error       error
}

// WBGetCardListRequestFilter defines the filter for listing cards.
type WBGetCardListRequestFilter struct {
	TextSearch            string   `json:"textSearch,omitempty"`
	WithPhoto             *int     `json:"withPhoto,omitempty"` // -1 for all, 0 for no photo, 1 for with photo
	AllowedCategoriesOnly *bool    `json:"allowedCategoriesOnly,omitempty"`
	TagIDs                []int    `json:"tagIDs,omitempty"`
	ObjectIDs             []int    `json:"objectIDs,omitempty"`
	Brands                []string `json:"brands,omitempty"`
	ImtID                 *int     `json:"imtID,omitempty"`
}

// WBGetCardListRequestCursor defines the cursor for paginating card lists.
type WBGetCardListRequestCursor struct {
	Limit     int     `json:"limit"`
	UpdatedAt *string `json:"updatedAt,omitempty"`
	NmID      *int    `json:"nmID,omitempty"`
}

// WBGetCardListRequestSort defines sorting for card lists.
type WBGetCardListRequestSort struct {
	Ascending bool `json:"ascending"`
}

// WBGetCardListRequestSettings defines the settings for the card list request.
type WBGetCardListRequestSettings struct {
	Sort   *WBGetCardListRequestSort   `json:"sort,omitempty"`
	Filter *WBGetCardListRequestFilter `json:"filter,omitempty"`
	Cursor WBGetCardListRequestCursor  `json:"cursor"`
}

// WBGetCardListRequest is the request payload for /content/v2/get/cards/list.
type WBGetCardListRequest struct {
	Settings WBGetCardListRequestSettings `json:"settings"`
}

// WBCardDefinition represents a card as returned by /content/v2/get/cards/list.
type WBCardDefinition struct {
	NmID       int    `json:"nmID"`
	VendorCode string `json:"vendorCode"`
	// Add other fields if necessary, e.g., Title, SubjectName for debugging
}

// WBGetCardListResponseCursorData is the cursor part of the response from /content/v2/get/cards/list.
type WBGetCardListResponseCursorData struct {
	UpdatedAt *string `json:"updatedAt,omitempty"`
	NmID      *int    `json:"nmID,omitempty"`
	Total     int     `json:"total"`
}

// WBGetCardListResponse is the response from /content/v2/get/cards/list.
type WBGetCardListResponse struct {
	Cards  []WBCardDefinition              `json:"cards"`
	Cursor WBGetCardListResponseCursorData `json:"cursor"`
	// Standard error fields are not part of the 200 response schema for this endpoint.
	// Errors are typically handled via HTTP status codes or a different error response structure for 4xx/5xx.
}
