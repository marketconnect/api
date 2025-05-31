package entities

// Dimensions represents product dimensions.
type Dimensions struct {
	Length       int     `json:"length"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	WeightBrutto float64 `json:"weight_brutto"`
}

// Size represents product size information.
type Size struct {
	TechSize string   `json:"tech_size"`         // e.g., XL, 45
	WbSize   string   `json:"wb_size,omitempty"` // Russian size, optional
	Price    int      `json:"price,omitempty"`   // Price for this specific size, optional
	Skus     []string `json:"skus"`              // Barcodes for this size
}

// CardCraftAiAPIRequest represents the request structure for the CardCraftAi API
type CardCraftAiAPIRequest struct {
	ProductTitle       string `json:"product_title"`
	ProductDescription string `json:"product_description"`
	ParentID           int32  `json:"parent_id"`
	SubjectID          int32  `json:"subject_id"`
	Translate          bool   `json:"translate"`
	Ozon               bool   `json:"ozon"`
	GenerateContent    bool   `json:"generate_content"`
}
