package entities

// OzonProductAttributeValue represents a value for a product attribute.
type OzonProductAttributeValue struct {
	DictionaryValueID int64  `json:"dictionary_value_id,omitempty"`
	Value             string `json:"value,omitempty"`
}

// OzonProductAttribute represents a product attribute for Ozon.
type OzonProductAttribute struct {
	ComplexID int64                       `json:"complex_id"`
	ID        int64                       `json:"id"`
	Values    []OzonProductAttributeValue `json:"values"`
}

// OzonComplexAttribute represents a complex attribute for Ozon (e.g., for video).
type OzonComplexAttribute struct {
	Attributes []OzonProductAttribute `json:"attributes"`
}

// OzonProductImportItem represents a single product item for Ozon's v3/product/import.
type OzonProductImportItem struct {
	Attributes            []OzonProductAttribute `json:"attributes"`
	Barcode               string                 `json:"barcode,omitempty"`
	DescriptionCategoryID int64                  `json:"description_category_id"`
	ColorImage            string                 `json:"color_image,omitempty"`
	ComplexAttributes     []OzonComplexAttribute `json:"complex_attributes,omitempty"`
	CurrencyCode          string                 `json:"currency_code,omitempty"` // Default: RUB
	Depth                 int64                  `json:"depth"`
	DimensionUnit         string                 `json:"dimension_unit,omitempty"` // e.g., mm
	Height                int64                  `json:"height"`
	Images                []string               `json:"images,omitempty"`
	Images360             []string               `json:"images360,omitempty"`
	Name                  string                 `json:"name"`
	OfferID               string                 `json:"offer_id"`
	OldPrice              string                 `json:"old_price,omitempty"`
	PdfList               []string               `json:"pdf_list,omitempty"` // Assuming URLs
	Price                 string                 `json:"price"`
	PrimaryImage          string                 `json:"primary_image,omitempty"`
	Vat                   string                 `json:"vat"` // e.g., "0.1" for 10%
	Weight                int64                  `json:"weight"`
	WeightUnit            string                 `json:"weight_unit,omitempty"` // e.g., g
	Width                 int64                  `json:"width"`
	// NewDescriptionCategoryID int64 `json:"new_description_category_id,omitempty"`
	// TypeID int64 `json:"type_id,omitempty"`
	// Promotions []interface{} `json:"promotions,omitempty"` // Define if needed
}

// OzonProductImportRequest is the request body for POST /v3/product/import.
type OzonProductImportRequest struct {
	Items []OzonProductImportItem `json:"items"`
}

// OzonProductImportResponseResult is the result part of the Ozon product import response.
type OzonProductImportResponseResult struct {
	TaskID int64 `json:"task_id"`
}

// OzonProductImportResponse is the response from POST /v3/product/import.
type OzonProductImportResponse struct {
	Result OzonProductImportResponseResult `json:"result"`
}

// OzonErrorDetail represents a detail in Ozon's error response.
type OzonErrorDetail struct {
	TypeURL string `json:"typeUrl"` // Note: Ozon's actual error structure might differ.
	Value   string `json:"value"`
}

// OzonError represents an error response from Ozon API.
type OzonError struct {
	Code    interface{}       `json:"code"` // Can be string or int
	Message string            `json:"message"`
	Details []OzonErrorDetail `json:"details,omitempty"`
}
