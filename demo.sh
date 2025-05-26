#!/bin/bash

echo "ConnectRPC Product API Demo"
echo "=========================="

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "Server is not running. Please start it with:"
    echo "  go run cmd/server/main.go"
    echo ""
    echo "Or in another terminal:"
    echo "  make run-server"
    exit 1
fi

echo "✓ Server is running"
echo ""

echo "Testing the API with cURL (Connect Protocol):"
echo "----------------------------------------------"

curl -v \
  --header "Content-Type: application/json" \
  --data '{
    "product_title": "Wireless Bluetooth Headphones", 
    "product_description": "High-quality wireless headphones with noise cancellation and long battery life."
  }' \
  http://localhost:8080/api.v1.ProductService/GetProductCard

echo ""
echo ""
echo "Note: This will fail if the Python API is not running."
echo "To test without the Python API, you can modify the server to return mock data." 