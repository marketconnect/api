package services

import (
	"api/app/domain/entities"
	"context"
)

type tokenCounterClient interface {
	GetSessionData(ctx context.Context, sessionID string) (*entities.SessionData, error)
}

type balanceStorage interface {
	GetBalance(ctx context.Context, apiKey string) (int, error)
	SetBalance(ctx context.Context, apiKey string, balance int) error
	GetTokenCost(ctx context.Context, tokenType string) (int, error)
}

type TokenBillingService struct {
	counterClient tokenCounterClient
	storage       balanceStorage
}

func NewTokenBillingService(counterClient tokenCounterClient, storage balanceStorage) *TokenBillingService {
	return &TokenBillingService{counterClient: counterClient, storage: storage}
}

func (s *TokenBillingService) UpdateBalanceForSession(ctx context.Context, apiKey, sessionID string) error {
	if apiKey == "" || sessionID == "" {
		return nil
	}
	data, err := s.counterClient.GetSessionData(ctx, sessionID)
	if err != nil {
		return err
	}
	inputCost, err := s.storage.GetTokenCost(ctx, "input")
	if err != nil {
		return err
	}
	outputCost, err := s.storage.GetTokenCost(ctx, "output")
	if err != nil {
		return err
	}
	totalCost := data.TotalPromptTokens*inputCost + data.TotalCompletionTokens*outputCost
	balance, err := s.storage.GetBalance(ctx, apiKey)
	if err != nil {
		return err
	}
	return s.storage.SetBalance(ctx, apiKey, balance-totalCost)
}
