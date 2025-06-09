package usecases

import (
	"context"
	"fmt"
	"time"
)

type updateBalanceStorage interface {
	GetBalance(ctx context.Context, apiKey string) (int, error)
	SetBalance(ctx context.Context, apiKey string, balance int) error
}

type UpdateBalanceUsecase struct {
	storage updateBalanceStorage
}

func NewUpdateBalanceUsecase(storage updateBalanceStorage) *UpdateBalanceUsecase {
	return &UpdateBalanceUsecase{storage: storage}
}

// UpdateBalance increases user balance after successful payment
// For now, we'll add a fixed amount per payment
func (uc *UpdateBalanceUsecase) UpdateBalance(ctx context.Context, userName string, endDate time.Time) (string, int64, []string, error) {
	// Define the amount to add based on subscription
	// This should be configurable or based on payment amount
	amountToAdd := 1000 // Default amount for now

	// Get current balance
	currentBalance, err := uc.storage.GetBalance(ctx, userName)
	if err != nil {
		// If user doesn't exist, start with 0
		currentBalance = 0
	}

	// Add payment amount
	newBalance := currentBalance + amountToAdd

	// Update balance
	err = uc.storage.SetBalance(ctx, userName, newBalance)
	if err != nil {
		return "", 0, nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Return successful update info
	message := fmt.Sprintf("Balance updated for user %s. New balance: %d", userName, newBalance)
	return message, int64(newBalance), []string{"balance_updated"}, nil
}
