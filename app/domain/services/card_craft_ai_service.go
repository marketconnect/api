package services

import (
	"api/app/domain/entities"
	"api/metrics"
	"context"
)

type cardCraftAiClient interface {
	GetCardContent(ctx context.Context, sessionID string, productCard entities.ProductCard) (*entities.CardCraftAiGeneratedContent, error)
	GetSessionID(ctx context.Context) (string, error)
}

type CardCraftAiService struct {
	cardCraftAiClient cardCraftAiClient
}

func NewCardCraftAiService(cardCraftAiClient cardCraftAiClient) *CardCraftAiService {
	return &CardCraftAiService{
		cardCraftAiClient: cardCraftAiClient,
	}
}

func (c *CardCraftAiService) GetCardContent(ctx context.Context, req entities.ProductCard) (*entities.CardCraftAiGeneratedContent, error) {
	sessionID, err := c.cardCraftAiClient.GetSessionID(ctx)
	if err != nil {
		metrics.AppExternalAPIErrorsTotal.WithLabelValues("card_craft_ai_session").Inc()
		return nil, err
	}

	cardCraftAiAPIResponse, err := c.cardCraftAiClient.GetCardContent(ctx, sessionID, req)
	if err != nil {
		metrics.AppExternalAPIErrorsTotal.WithLabelValues("card_craft_ai_content").Inc()
		return nil, err
	}

	return cardCraftAiAPIResponse, nil
}
