package card_craft_ai

import (
	"api/app/domain/entities"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type CardCraftAiClient struct {
	cardCraftAiAPIURL string
	getSessionURL     string
}

func NewCardCraftAiClient(cardCraftAiAPIURL string) *CardCraftAiClient {
	return &CardCraftAiClient{
		cardCraftAiAPIURL: cardCraftAiAPIURL,
		getSessionURL:     fmt.Sprintf("%s/v1/sessions", cardCraftAiAPIURL),
	}
}

func (c *CardCraftAiClient) GetSessionID(ctx context.Context) (string, error) {
	// Build session URL
	sessionURL := c.getSessionURL
	log.Printf("Requesting session from: %s", sessionURL)

	// Create request for session
	httpReq, err := http.NewRequestWithContext(ctx, "POST", sessionURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create session request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the session request
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call session API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read session response body: %w", err)
	}

	log.Printf("Session API response status: %d, body: %s", resp.StatusCode, string(respBody))

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("session API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse session response
	var sessionResp struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(respBody, &sessionResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal session response: %w", err)
	}

	log.Printf("Successfully got session ID: %s", sessionResp.SessionID)
	return sessionResp.SessionID, nil
}

func (c *CardCraftAiClient) GetCardContent(ctx context.Context, sessionID string, productCard entities.ProductCard) (*entities.CardCraftAiGeneratedContent, error) {
	log.Printf("Got session ID: %s", sessionID)

	cardCraftAiAPIRequest := map[string]interface{}{
		"product_title":       productCard.ProductTitle,
		"product_description": productCard.ProductDescription,
		"session_id":          sessionID,
		"parent_id":           productCard.ParentId,
		"subject_id":          productCard.SubjectId,
		"translate":           productCard.Translate,
		"ozon":                productCard.Ozon,
		"generate_content":    productCard.GenerateContent,
	}
	// Marshal request to JSON
	reqBody, err := json.Marshal(cardCraftAiAPIRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL for /run endpoint
	productCardURL := fmt.Sprintf("%s/run", c.cardCraftAiAPIURL)

	// Make request to Python API with session ID in body
	httpReq, err := http.NewRequestWithContext(ctx, "POST", productCardURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call CardCraftAI API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CardCraftAI API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse CardCraftAI API response
	var cardCraftAiResp entities.CardCraftAiGeneratedContent
	if err := json.Unmarshal(respBody, &cardCraftAiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CardCraftAI API response: %w", err)
	}

	log.Printf("CardCraftAI API response: title=%s, description=%s, parentID=%v, parentName=%v, subjectID=%v, subjectName=%v, typeID=%v, typeName=%v, rootID=%v, rootName=%v, subID=%v, subName=%v",
		cardCraftAiResp.Title,
		cardCraftAiResp.Description,
		safeInt32Value(cardCraftAiResp.ParentID),
		safeStringValue(cardCraftAiResp.ParentName),
		safeInt32Value(cardCraftAiResp.SubjectID),
		safeStringValue(cardCraftAiResp.SubjectName),
		safeInt32Value(cardCraftAiResp.TypeID),
		safeStringValue(cardCraftAiResp.TypeName),
		safeInt32Value(cardCraftAiResp.RootID),
		safeStringValue(cardCraftAiResp.RootName),
		safeInt32Value(cardCraftAiResp.SubID),
		safeStringValue(cardCraftAiResp.SubName))

	// Create ConnectRPC response with the comprehensive data from Python API
	response := &entities.CardCraftAiGeneratedContent{
		Title:       cardCraftAiResp.Title,
		Attributes:  cardCraftAiResp.Attributes,
		Description: cardCraftAiResp.Description,
		ParentID:    cardCraftAiResp.ParentID,
		ParentName:  cardCraftAiResp.ParentName,
		SubjectID:   cardCraftAiResp.SubjectID,
		SubjectName: cardCraftAiResp.SubjectName,
		TypeID:      cardCraftAiResp.TypeID,
		TypeName:    cardCraftAiResp.TypeName,
		RootID:      cardCraftAiResp.RootID,
		RootName:    cardCraftAiResp.RootName,
		SubID:       cardCraftAiResp.SubID,
		SubName:     cardCraftAiResp.SubName,
	}

	return response, nil
}

// safeInt32Value safely dereferences an *int32 pointer
func safeInt32Value(ptr *int32) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

// safeStringValue safely dereferences a *string pointer
func safeStringValue(ptr *string) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}
