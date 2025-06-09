.PHONY: help install generate build run test clean lint format

help:
	@echo "Available targets:"
	@echo "  install   - Install required tools and dependencies"
	@echo "  generate  - Generate code from protobuf schema"
	@echo "  build     - Build server and client binaries"
	@echo "  run       - Run the server"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean generated files and binaries"
	@echo "  lint      - Lint protobuf files"
	@echo "  format    - Format protobuf files"

install:
	go mod tidy
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest

generate:
	buf generate

build:
	mkdir -p bin
	go build -o bin/server cmd/server/main.go
	go build -o bin/client cmd/client/main.go

run:
	go run cmd/server/main.go

test:
	go test ./...

clean:
	rm -rf gen/
	rm -rf bin/

lint:
	buf lint

format:
	buf format -w 

