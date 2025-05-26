# ConnectRPC Product API

This is a ConnectRPC API that retrieves product information with the following fields:
- `title` (string)
- `description` (string) 
- `generate_content` (boolean)
- `ozon` (boolean)
- `wb` (boolean)

The API acts as a proxy to a Python API endpoint and transforms the response into the required format.

## Features

- **ConnectRPC Protocol**: Supports Connect, gRPC, and gRPC-Web protocols
- **Type-Safe**: Generated from Protocol Buffer schema
- **HTTP/2 Support**: Built-in HTTP/2 support with h2c
- **Python API Integration**: Makes requests to your Python API endpoint
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

- `PYTHON_API_URL`: URL of the Python API endpoint (default: `http://localhost:8000/v1/product_card_comprehensive`)
- `PORT`: Port to run the server on (default: `8080`)

### Start the server:

```bash
go run cmd/server/main.go
```

Or with custom configuration:
```bash
PYTHON_API_URL=http://your-python-api:8000/v1/product_card_comprehensive PORT=9090 go run cmd/server/main.go
```

## API Usage

### ConnectRPC Endpoint

The API exposes one endpoint:
- **Service**: `api.v1.ProductService`
- **Method**: `GetProductCard`
- **URL**: `/api.v1.ProductService/GetProductCard`

### Request Format

```json
{
  "product_title": "Product Title",
  "product_description": "Product Description"
}
```

### Response Format

```json
{
  "title": "Optimized Product Title",
  "description": "Optimized Product Description", 
  "generate_content": true,
  "ozon": true,
  "wb": false
}
```

### Using cURL (Connect Protocol)

```bash
curl \
  --header "Content-Type: application/json" \
  --data '{"product_title": "Wireless Headphones", "product_description": "High-quality audio device"}' \
  http://localhost:8080/api.v1.ProductService/GetProductCard
```

### Using grpcurl (gRPC Protocol)

```bash
grpcurl \
  -plaintext \
  -d '{"product_title": "Wireless Headphones", "product_description": "High-quality audio device"}' \
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

## Python API Integration

The ConnectRPC server makes POST requests to your Python API with the following structure:

### Request to Python API:
```json
{
  "product_title": "Product Title",
  "product_description": "Product Description"
}
```

### Expected Response from Python API:
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

### Business Logic

The ConnectRPC API transforms the Python API response as follows:
- `title`: Direct mapping from Python API
- `description`: Direct mapping from Python API  
- `generate_content`: Always set to `true` (can be customized)
- `ozon`: Set to `true` if `type_id`, `root_id`, and `sub_id` are present
- `wb`: Set to `true` if `subject_id` and `parent_id` are present

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

You can test the API without the Python backend by modifying the server to return mock data, or by setting up a mock HTTP server.

## Protocol Support

This API supports three protocols:

1. **Connect Protocol**: Simple HTTP-based protocol (default)
2. **gRPC Protocol**: Standard gRPC protocol
3. **gRPC-Web Protocol**: gRPC protocol for web browsers

Clients can choose which protocol to use when making requests.

## Error Handling

The API returns appropriate ConnectRPC error codes:
- `CodeInternal`: For internal server errors
- `CodeUnavailable`: When the Python API is unreachable
- `CodeInvalidArgument`: For malformed requests

## License

This project is licensed under the Apache 2.0 License. 