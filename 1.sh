#!/bin/bash

# Encode file properly (no newlines!)
BASE64_CONTENT_1=$(base64 -w 0 photo1.jpg)
BASE64_CONTENT_2=$(base64 -w 0 photo2.jpg)

# Create valid JSON
cat > payload.json <<EOF
{
  "product_title": "Значок Сарказм, Prank Bank, в ассортименте",
  "product_description": "Значок Сарказм, Prank Bank, с ироничной надписью - лучший способ поднять настроение себе и окружающим! Приблизительный размер: 4х6 см. Состав: фанера (берёза), картон, ламинационная плёнка, фурнитура. Товар представлен в ассортименте.",
  "wb_api_key":  "$WB_TOKEN",
  "ozon_api_key": "$OZON_API_KEY",
  "ozon_api_client_id": "$OZON_CLIENT_ID",
  "ozon": true,
  "dimensions": {
    "length": 10,
    "width": 10,
    "height": 10,
    "weight_brutto": 10,
    "depth": 10,
    "dimension_unit": "mm",
    "weight": 10,
    "weight_unit": "g"
  },
  "wb": true,
  "vendor_code": "badge_2",
  "sizes": [
    {
      "tech_size": "0",
      "wb_size": "Универсальный",
      "price": 15000,
      "wb_price": 12000,
      "ozon_price": 18000,
      "skus": ["4627180440022"]
    }
  ],
  "wb_media_to_upload_files": [
    {
      "content": "$BASE64_CONTENT_1",
      "filename": "photo1.jpg",
      "photo_number": 2
    },
    {
      "content": "$BASE64_CONTENT_2",
      "filename": "photo2.jpg",
      "photo_number": 1
    }
  ]
}
EOF

# Send it
curl -X POST https://api.mcpconnect.ru/api.v1.ProductService/Create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test" \
  -d @payload.json
