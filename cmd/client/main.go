package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"

	apiv1 "api/gen/api/v1"
	"api/gen/api/v1/apiv1connect"
)

func main() {
	// Get server URL from environment variable or use default
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	// Create the client
	client := apiv1connect.NewProductServiceClient(
		http.DefaultClient,
		serverURL,
	)

	// Create a request
	req := connect.NewRequest(&apiv1.ProductRequest{
		ProductTitle:       "Wireless Bluetooth Headphones",
		ProductDescription: "High-quality wireless headphones with noise cancellation and long battery life.",
	})

	// Set headers if needed
	req.Header().Set("User-Agent", "ConnectRPC-Client/1.0")

	// Make the request
	log.Printf("Making request to %s", serverURL)
	res, err := client.GetProductCard(context.Background(), req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	// Print the response
	log.Printf("Response received:")
	log.Printf("  Title: %s", res.Msg.Title)
	log.Printf("  Description: %s", res.Msg.Description)
	log.Printf("  Generate Content: %t", res.Msg.GenerateContent)
	log.Printf("  Ozon: %t", res.Msg.Ozon)
	log.Printf("  WB: %t", res.Msg.Wb)

	// Print response headers
	log.Printf("Response headers:")
	for key, values := range res.Header() {
		for _, value := range values {
			log.Printf("  %s: %s", key, value)
		}
	}
}
