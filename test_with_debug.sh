#!/bin/bash

echo "=== Building server with debug logging ==="
go build -o main cmd/server/main.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "=== Starting server in background ==="
export CARD_CRAFT_AI_API_URL=localhost
export TOKEN_COUNTER_API_URL=localhost  
export PG_DATABASE=test
export PG_USER=test
export PG_PASSWORD=test
export TINKOFF_SECRET_KEY=test
export TINKOFF_TERMINAL_KEY=test
export FILE_STORAGE_UPLOAD_DIR=./debug_uploads

# Create upload directory
mkdir -p ./debug_uploads
chmod 755 ./debug_uploads

# Start server in background and capture logs
./main > server.log 2>&1 &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"
echo "Waiting for server to start..."
sleep 3

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "Server failed to start! Logs:"
    cat server.log
    exit 1
fi

echo "=== Testing image upload with debug logging ==="

# Create a simple test image
echo -e "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A\x00\x00\x00\x0D\x49\x48\x44\x52\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90\x77\x53\xDE\x00\x00\x00\x0C\x49\x44\x41\x54\x08\x99\x01\x01\x00\x00\xFF\xFF\x00\x00\x00\x02\x00\x01\x73\x75\x01\x18\x00\x00\x00\x00\x49\x45\x4E\x44\xAE\x42\x60\x82" > test_image.png

# Base64 encode the image
IMAGE_BASE64=$(base64 -w0 test_image.png)

# Make the test request
curl -X POST "http://localhost:8080/api.v1.ProductService/Create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_api_key" \
  -d '{
    "product_title": "Test Badge Debug",
    "product_description": "A test badge for debugging image upload",
    "generate_content": false,
    "ozon": true,
    "wb": false,
    "translate": false,
    "vendor_code": "badge_debug",
    "dimensions": {
      "depth": 10,
      "width": 10,
      "height": 10,
      "weight": 10,
      "dimension_unit": "mm",
      "weight_unit": "g"
    },
    "brand": "Debug Brand",
    "wb_media_to_upload_files": [
      {
        "content": "'$IMAGE_BASE64'",
        "filename": "debug_image.png",
        "photo_number": 1
      }
    ],
    "wb_media_to_save_links": [
      "https://example.com/debug_image.jpg"
    ],
    "ozon_api_client_id": "debug_client_id",
    "ozon_api_key": "debug_api_key"
  }' > response.json

echo -e "\n=== API Response ==="
cat response.json | jq . 2>/dev/null || cat response.json

echo -e "\n=== Server Logs (Debug Info) ==="
cat server.log | grep -E "\[OZON DEBUG\]|\[FILE UPLOAD\]|\[FILE STORAGE\]" | tail -20

echo -e "\n=== Uploaded Files ==="
ls -la ./debug_uploads/ 2>/dev/null || echo "No upload directory found"

echo -e "\n=== Stopping server ==="
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\n=== Full Server Logs ==="
cat server.log

# Cleanup
rm -f test_image.png response.json main
rm -rf ./debug_uploads 