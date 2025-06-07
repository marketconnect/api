package usecases

import "context"

type balanceStorage interface {
	GetBalance(ctx context.Context, apiKey string) (int, error)
}

type GetBalanceUsecase struct {
	storage balanceStorage
}

func NewGetBalanceUsecase(storage balanceStorage) *GetBalanceUsecase {
	return &GetBalanceUsecase{storage: storage}
}

func (uc *GetBalanceUsecase) GetBalance(ctx context.Context, apiKey string) (int, error) {
	return uc.storage.GetBalance(ctx, apiKey)
}
