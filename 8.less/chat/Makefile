.PHONY: generate build test lint run

CURDIR=$(shell pwd)
LOCAL_BIN:=$(CURDIR)/bin
PROTO_DIR := proto
GO_OUT_DIR := pkg
PROTO_FILE := chat/v1/chat.proto
DOCKER_IMAGE_NAME := chat-server
DOCKER_CONTAINER_NAME := chat-server-container

install-deps:
	@echo "Installing dependencies..."
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	@GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v0.10.1
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.15.2
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.15.2
	@GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3

generate:
	@echo "Generating Go code from protobuf..."
	@mkdir -p ./$(GO_OUT_DIR)/chat/v1
	@PATH="$(LOCAL_BIN):$(PATH)" protoc \
    		-I=./$(PROTO_DIR) \
    		-I=./proto \
    		-I=./proto/google/api \
    		-I=./proto/protoc-gen-openapiv2/options \
    		-I=./proto/validate \
    		--go_out=./$(GO_OUT_DIR) --go_opt=paths=source_relative \
    		--go-grpc_out=./$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
    		--grpc-gateway_out=./$(GO_OUT_DIR) --grpc-gateway_opt=paths=source_relative \
    		--validate_out="lang=go,paths=source_relative:./$(GO_OUT_DIR)" \
    		--openapiv2_out=./$(GO_OUT_DIR) \
    		./$(PROTO_DIR)/$(PROTO_FILE)


build: generate
	@echo "Building..."
	@go build -o bin/chat cmd/server/main.go

test:
	@echo "Running tests..."
	@go test ./...

lint:
	@echo "Running linter..."
	@golangci-lint run

run: generate
	@echo "Running server..."
	@go run cmd/server/main.go

docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE_NAME) .

docker-run:
	@echo "Running Docker container..."
	@docker run --name $(DOCKER_CONTAINER_NAME) --rm -p 50051:50051 -p 8080:8080 $(DOCKER_IMAGE_NAME) $(ARGS)

docker-stop:
	@echo "Stopping Docker container..."
	@docker stop $(DOCKER_CONTAINER_NAME) || true

.DEFAULT_GOAL := build