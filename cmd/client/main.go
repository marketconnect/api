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

	// Create a request with the 5 required input fields
	req := connect.NewRequest(&apiv1.ProductRequest{
		Title:           "Wireless Bluetooth Headphones",
		Description:     "High-quality wireless headphones with noise cancellation and long battery life.",
		GenerateContent: true,
		Ozon:            true,
		Wb:              false,
	})

	// Set headers if needed
	req.Header().Set("User-Agent", "ConnectRPC-Client/1.0")

	// Make the request
	log.Printf("Making request to %s", serverURL)
	res, err := client.GetProductCard(context.Background(), req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	// Print the comprehensive response from Python API
	log.Printf("Response received:")
	log.Printf("  Title: %s", res.Msg.Title)
	log.Printf("  Description: %s", res.Msg.Description)
	log.Printf("  Attributes: %v", res.Msg.Attributes)
	log.Printf("  Parent ID: %d", res.Msg.ParentId)
	log.Printf("  Subject ID: %d", res.Msg.SubjectId)
	log.Printf("  Type ID: %d", res.Msg.TypeId)
	log.Printf("  Root ID: %d", res.Msg.RootId)
	log.Printf("  Sub ID: %d", res.Msg.SubId)
	log.Printf("  Keywords: %v", res.Msg.Keywords)

	// Print response headers
	log.Printf("Response headers:")
	for key, values := range res.Header() {
		for _, value := range values {
			log.Printf("  %s: %s", key, value)
		}
	}
}
