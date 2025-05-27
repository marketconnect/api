# ConnectRPC Product API

This is a ConnectRPC proxy server for the Python CardCraftAI service. It provides a REST API with one endpoint that:
- **Receives** JSON input with 5 fields: `title`, `description`, `generate_content`, `ozon`, `wb`
- **Makes requests** to the Python CardCraftAI service
- **Returns** the comprehensive response from CardCraftAI

## Features

- **ConnectRPC Protocol**: Supports Connect, gRPC, and gRPC-Web protocols
- **Type-Safe**: Generated from Protocol Buffer schema
- **HTTP/2 Support**: Built-in HTTP/2 support with h2c
- **Python CardCraftAI Integration**: Proxy server for CardCraftAI service
- **Health Check**: Includes a `/health` endpoint for monitoring

## Project Structure

```
.
├── proto/api/v1/product.proto    # Protocol Buffer schema
├── gen/api/v1/                   # Generated Go code
├── cmd/server/main.go            # Server implementation
├── cmd/client/main.go            # Client example
├── buf.yaml                      # Buf configuration
├── buf.gen.yaml                  # Code generation configuration
└── go.mod                        # Go module dependencies
```

## Prerequisites

- Go 1.24.2 or later
- buf CLI tool
- protoc-gen-go plugin
- protoc-gen-connect-go plugin

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd api
```

2. Install dependencies:
```bash
go mod tidy
```

3. Install required tools:
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
```

4. Generate code from protobuf schema:
```bash
buf generate
```

## Running the Server

### Environment Variables

- `PYTHON_API_URL`: URL of the Python CardCraftAI endpoint (default: `http://localhost:8000/v1/product_card_comprehensive`)
- `PORT`: Port to run the server on (default: `8080`)

### Start the server:

```bash
go run cmd/server/main.go
```

Or with custom configuration:
```bash
PYTHON_API_URL=http://your-cardcraft-api:8000/v1/product_card_comprehensive PORT=9090 go run cmd/server/main.go
```

## API Usage

### ConnectRPC Endpoint

The API exposes one endpoint:
- **Service**: `api.v1.ProductService`
- **Method**: `GetProductCard`
- **URL**: `/api.v1.ProductService/GetProductCard`

### Request Format (Input)

```json
{
  "title": "Product Title",
  "description": "Product Description",
  "generate_content": true,
  "ozon": true,
  "wb": false
}
```

### Response Format (Output)

Returns the comprehensive response from Python CardCraftAI:

```json
{
  "title": "Optimized Product Title",
  "description": "Optimized Product Description",
  "attributes": {"key": "value"},
  "parent_id": 123,
  "subject_id": 456,
  "type_id": 789,
  "root_id": 101,
  "sub_id": 112,
  "keywords": ["keyword1", "keyword2"]
}
```

### Using cURL (Connect Protocol)

```bash
curl \
  --header "Content-Type: application/json" \
  --data '{
    "title": "Wireless Headphones", 
    "description": "High-quality audio device",
    "generate_content": true,
    "ozon": true,
    "wb": false
  }' \
  http://localhost:8080/api.v1.ProductService/GetProductCard
```

### Using grpcurl (gRPC Protocol)

```bash
grpcurl \
  -plaintext \
  -d '{
    "title": "Wireless Headphones", 
    "description": "High-quality audio device",
    "generate_content": true,
    "ozon": true,
    "wb": false
  }' \
  localhost:8080 \
  api.v1.ProductService/GetProductCard
```

### Using the Go Client

```bash
go run cmd/client/main.go
```

Or with custom server URL:
```bash
SERVER_URL=http://localhost:9090 go run cmd/client/main.go
```

## Python CardCraftAI Integration

The ConnectRPC proxy server:

1. **Receives ConnectRPC request** with 5 input fields
2. **Extracts** `title` and `description` from the input
3. **Makes HTTP POST request** to CardCraftAI with:
   ```json
   {
     "product_title": "title_from_input",
     "product_description": "description_from_input"
   }
   ```
4. **Returns** the comprehensive CardCraftAI response directly

### Expected CardCraftAI Response:
```json
{
  "title": "Optimized Title",
  "attributes": {"key": "value"},
  "description": "Optimized Description",
  "parent_id": 123,
  "subject_id": 456,
  "type_id": 789,
  "root_id": 101,
  "sub_id": 112,
  "keywords": ["keyword1", "keyword2"]
}
```

## Health Check

The server includes a health check endpoint:

```bash
curl http://localhost:8080/health
```

## Development

### Regenerating Code

If you modify the protobuf schema, regenerate the code:

```bash
buf generate
```

### Building

```bash
go build -o server cmd/server/main.go
go build -o client cmd/client/main.go
```

### Testing

```bash
go test ./...
```

Use the demo script: `./demo.sh`

## Protocol Support

This API supports three protocols:

1. **Connect Protocol**: Simple HTTP-based protocol (default)
2. **gRPC Protocol**: Standard gRPC protocol
3. **gRPC-Web Protocol**: gRPC protocol for web browsers

Clients can choose which protocol to use when making requests.

## Error Handling

The API returns appropriate ConnectRPC error codes:
- `CodeInternal`: For internal server errors
- `CodeUnavailable`: When the Python CardCraftAI API is unreachable
- `CodeInvalidArgument`: For malformed requests

## License

This project is licensed under the Apache 2.0 License. 