package main

import (
	"api/app/domain/entities"
	"encoding/json"
	"fmt"
)

func main() {
	// Test 1: Empty images slice
	fmt.Println("=== Test 1: Empty images slice ===")
	item1 := entities.OzonProductImportItem{
		Name:                  "Test Product",
		OfferID:               "test123",
		DescriptionCategoryID: 123,
		TypeID:                456,
		Price:                 "100",
		Vat:                   "0",
		Images:                []string{}, // Empty slice
	}

	json1, _ := json.Marshal(item1)
	fmt.Printf("JSON with empty images: %s\n\n", string(json1))

	// Test 2: Nil images slice
	fmt.Println("=== Test 2: Nil images slice ===")
	item2 := entities.OzonProductImportItem{
		Name:                  "Test Product",
		OfferID:               "test123",
		DescriptionCategoryID: 123,
		TypeID:                456,
		Price:                 "100",
		Vat:                   "0",
		Images:                nil, // Nil slice
	}

	json2, _ := json.Marshal(item2)
	fmt.Printf("JSON with nil images: %s\n\n", string(json2))

	// Test 3: With actual images
	fmt.Println("=== Test 3: With actual images ===")
	item3 := entities.OzonProductImportItem{
		Name:                  "Test Product",
		OfferID:               "test123",
		DescriptionCategoryID: 123,
		TypeID:                456,
		Price:                 "100",
		Vat:                   "0",
		Images:                []string{"http://example.com/image1.jpg", "http://example.com/image2.jpg"},
	}

	json3, _ := json.Marshal(item3)
	fmt.Printf("JSON with images: %s\n\n", string(json3))

	// Test 4: Full payload
	fmt.Println("=== Test 4: Full payload ===")
	payload := entities.OzonProductImportRequest{
		Items: []entities.OzonProductImportItem{item3},
	}

	jsonPayload, _ := json.Marshal(payload)
	fmt.Printf("Full payload JSON: %s\n", string(jsonPayload))
}
