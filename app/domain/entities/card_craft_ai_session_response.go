package entities

// CardCraftAiSessionResponse represents the session response from the v1/sessions endpoint
type CardCraftAiSessionResponse struct {
	SessionID string `json:"session_id"`
	AppName   string `json:"app_name"`
	UserID    string `json:"user_id"`
}
