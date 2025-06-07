package presentation

import (
	"context"
	"encoding/json"
	"net/http"
)

type balanceUsecase interface {
	GetBalance(ctx context.Context, apiKey string) (int, error)
}

type BalanceHandler struct {
	usecase balanceUsecase
}

func NewBalanceHandler(uc balanceUsecase) *BalanceHandler {
	return &BalanceHandler{usecase: uc}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusBadRequest)
		return
	}
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		http.Error(w, "invalid Authorization header", http.StatusBadRequest)
		return
	}
	apiKey := authHeader[len(bearerPrefix):]

	balance, err := h.usecase.GetBalance(r.Context(), apiKey)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}
