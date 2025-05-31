#!/bin/bash

# Encode file properly (no newlines!)
BASE64_CONTENT_1=$(base64 -w 0 photo1.jpg)
BASE64_CONTENT_2=$(base64 -w 0 photo2.jpg)

# Create valid JSON
cat > payload.json <<EOF
{
  "product_title": "Значок Сарказм, Prank Bank, в ассортименте",
  "product_description": "Значок Сарказм, Prank Bank, с ироничной надписью – лучший способ поднять настроение себе и окружающим! Приблизительный размер: 4х6 см. Состав: фанера (берёза), картон, ламинационная плёнка, фурнитура. Товар представлен в ассортименте.",
  "wb_api_key":  "$WB_TOKEN",
  "wb": true,
  "vendor_code": "testCode",
  "wb_media_to_upload_files": [
    {
      "content": "$BASE64_CONTENT_1",
      "filename": "photo1.jpg",
      "photo_number": 1
    },
    {
      "content": "$BASE64_CONTENT_2",
      "filename": "photo2.jpg",
      "photo_number": 2
    }
  ]
}
EOF

# Send it
curl -X POST http://localhost:8080/api.v1.CreateProductCardService/CreateProductCard \
  -H "Content-Type: application/json" \
  -d @payload.json
