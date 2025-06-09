#!/bin/bash

echo "Testing Ozon validation..."

# Test 1: Missing dimensions
echo "Test 1: Missing dimensions"
curl -X POST http://45.141.76.230:8080/api.v1.ProductService/Create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test" \
  -d '{
    "product_title": "Test Product",
    "product_description": "Test Description",
    "ozon": true,
    "vendor_code": "TEST123",
    "ozon_api_key": "test_key",
    "ozon_api_client_id": "test_client"
  }' | jq .

echo -e "\n\n"

# Test 2: Zero dimensions
echo "Test 2: Zero dimensions"
curl -X POST http://45.141.76.230:8080/api.v1.ProductService/Create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test" \
  -d '{
    "product_title": "Test Product", 
    "product_description": "Test Description",
    "ozon": true,
    "vendor_code": "TEST123",
    "ozon_api_key": "test_key",
    "ozon_api_client_id": "test_client",
    "dimensions": {
      "length": 0,
      "width": 0,
      "height": 0,
      "weight_brutto": 0
    },
    "sizes": [{"price": 100}]
  }' | jq .

echo -e "\n\n"

# Test 3: Missing API credentials
echo "Test 3: Missing API key"
curl -X POST http://45.141.76.230:8080/api.v1.ProductService/Create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test" \
  -d '{
    "product_title": "Test Product",
    "product_description": "Test Description", 
    "ozon": true,
    "vendor_code": "TEST123",
    "ozon_api_client_id": "test_client",
    "dimensions": {
      "length": 100,
      "width": 50,
      "height": 30,
      "weight_brutto": 250
    },
    "sizes": [{"price": 100}]
  }' | jq .

echo -e "\n\n"

# Test 4: Valid Ozon request
echo "Test 4: Valid Ozon request"
curl -X POST http://45.141.76.230:8080/api.v1.ProductService/Create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test" \
  -d '{
    "product_title": "Test Product",
    "product_description": "Test Description",
    "ozon": true,
    "vendor_code": "TEST123", 
    "ozon_api_key": "test_key",
    "ozon_api_client_id": "test_client",
    "dimensions": {
      "length": 100,
      "width": 50,
      "height": 30,
      "weight_brutto": 250
    },
    "sizes": [{"price": 1500}]
  }' | jq . 