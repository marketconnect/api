package presentation

import (
	"context"
	"encoding/json"
	"net/http"

	apiv1 "api/gen/api/v1"

	"connectrpc.com/connect"
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

// GetBalanceHTTP handles HTTP balance requests using Authorization header
func (h *BalanceHandler) GetBalanceHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey, err := ExtractAPIKeyFromHeader(r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	balance, err := h.usecase.GetBalance(r.Context(), apiKey)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

func (h *BalanceHandler) GetBalanceByToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey, err := ExtractAPIKeyFromHeader(r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	balance, err := h.usecase.GetBalance(r.Context(), apiKey)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

// GetBalance implements the ConnectRPC BalanceServiceHandler interface
func (h *BalanceHandler) GetBalance(ctx context.Context, req *connect.Request[apiv1.GetBalanceRequest]) (*connect.Response[apiv1.GetBalanceResponse], error) {
	apiKey, err := ExtractAPIKeyFromHeader(req.Header())
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	balance, err := h.usecase.GetBalance(ctx, apiKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &apiv1.GetBalanceResponse{
		Balance: int32(balance),
	}

	return &connect.Response[apiv1.GetBalanceResponse]{
		Msg: response,
	}, nil
}
