package presentation

import (
	"fmt"
	"net/http"
)

// ExtractAPIKeyFromHeader extracts and validates the API key from Authorization header
// This function can be used by both HTTP and ConnectRPC handlers
func ExtractAPIKeyFromHeader(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	return extractAPIKeyFromAuthHeaderString(authHeader)
}

// extractAPIKeyFromAuthHeaderString extracts and validates the API key from Authorization header string
func extractAPIKeyFromAuthHeaderString(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("invalid authorization header format, expected 'Bearer <token>'")
	}

	apiKey := authHeader[len(bearerPrefix):]
	if apiKey == "" {
		return "", fmt.Errorf("empty API key in authorization header")
	}

	return apiKey, nil
}
