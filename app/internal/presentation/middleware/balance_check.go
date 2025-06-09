package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const MinRequiredBalance = 10

type balanceChecker interface {
	GetBalance(ctx context.Context, apiKey string) (int, error)
}

type apiKeyExtractor func(header http.Header) (string, error)

type BalanceCheckMiddleware struct {
	balanceChecker balanceChecker
	extractor      apiKeyExtractor
}

func NewBalanceCheckMiddleware(balanceChecker balanceChecker, extractor apiKeyExtractor) *BalanceCheckMiddleware {
	return &BalanceCheckMiddleware{
		balanceChecker: balanceChecker,
		extractor:      extractor,
	}
}

// isBalanceEndpoint checks if the request is for balance-related endpoints
func (m *BalanceCheckMiddleware) isBalanceEndpoint(r *http.Request) bool {
	path := r.URL.Path

	// Allow balance endpoints
	if strings.HasSuffix(path, "/GetBalance") ||
		path == "/balance" ||
		path == "/balance-by-token" {
		return true
	}

	// Allow payment endpoints (all payment-related paths)
	if strings.HasPrefix(path, "/payment/") {
		return true
	}

	// Allow metrics endpoint
	if path == "/metrics" {
		return true
	}

	return false
}

// CheckBalance middleware function
func (m *BalanceCheckMiddleware) CheckBalance(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip balance check for balance endpoints and other excluded paths
		if m.isBalanceEndpoint(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract API key from request
		apiKey, err := m.extractor(r.Header)
		if err != nil {
			http.Error(w, "Unauthorized: missing or invalid API key", http.StatusUnauthorized)
			return
		}

		// Check balance
		balance, err := m.balanceChecker.GetBalance(r.Context(), apiKey)
		if err != nil {
			http.Error(w, "Failed to check balance", http.StatusInternalServerError)
			return
		}

		// Reject if balance is less than minimum required
		if balance < MinRequiredBalance {
			http.Error(w, fmt.Sprintf("Insufficient balance: %d. Minimum required: %d", balance, MinRequiredBalance), http.StatusPaymentRequired)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}
