#!/bin/bash

# Test Ozon integration with image upload

# Create a simple test image (1x1 pixel PNG)
echo "Creating test image..."
# This creates a minimal valid PNG file
echo -e "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A\x00\x00\x00\x0D\x49\x48\x44\x52\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90\x77\x53\xDE\x00\x00\x00\x0C\x49\x44\x41\x54\x08\x99\x01\x01\x00\x00\xFF\xFF\x00\x00\x00\x02\x00\x01\x73\x75\x01\x18\x00\x00\x00\x00\x49\x45\x4E\x44\xAE\x42\x60\x82" > test_image.png

# Base64 encode the image
IMAGE_BASE64=$(base64 -w0 test_image.png)

echo "Testing Ozon integration with image upload..."

# Test with image files - using localhost for local testing
curl -X POST "http://localhost:8080/api.v1.ProductService/Create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_api_key" \
  -d '{
    "product_title": "Test Badge with Image",
    "product_description": "A test badge for Ozon with image upload",
    "generate_content": false,
    "ozon": true,
    "wb": false,
    "translate": false,
    "vendor_code": "badge_with_image",
    "dimensions": {
      "depth": 10,
      "width": 10,
      "height": 10,
      "weight": 10,
      "dimension_unit": "mm",
      "weight_unit": "g"
    },
    "brand": "Prank Bank",
    "wb_media_to_upload_files": [
      {
        "content": "'$IMAGE_BASE64'",
        "filename": "test_badge.png",
        "photo_number": 1
      }
    ],
    "wb_media_to_save_links": [
      "https://example.com/test_image.jpg"
    ],
    "ozon_api_client_id": "test_client_id",
    "ozon_api_key": "test_api_key"
  }' | jq .

echo -e "\n\nTest completed. Check the logs for debug information about image processing."

# Clean up
rm -f test_image.png 