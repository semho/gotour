.PHONY: generate build test lint run fmt migrate-up migrate-down migrate-create

CURDIR=$(shell pwd)
LOCAL_BIN:=$(CURDIR)/bin
PROTO_DIR := proto
GO_OUT_DIR := pkg
PROTO_FILE := chat/v1/chat.proto
DOCKER_IMAGE_NAME := chat-server
DOCKER_CONTAINER_NAME := chat-server-container
MIGRATIONS_DIR:= migrations

install-deps:
	@echo "Installing dependencies..."
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	@GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v0.10.1
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.15.2
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.15.2
	@GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3
	@GOBIN=$(LOCAL_BIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest


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
	@go test -v ./...

fmt:
	@echo "Running goimports..."
	@goimports -w .

lint: fmt
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

migrate-create:
	@echo "Creating migration files..."
	@if [ -z "$(name)" ]; then echo "Please specify migration name by running 'make migrate-create name=your_migration_name'"; exit 1; fi
	@mkdir -p $(MIGRATIONS_DIR)
	@$(LOCAL_BIN)/migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	@echo "Running migrations up..."
	@if [ -z "$(DB_DSN)" ]; then echo "Please specify DB_DSN"; exit 1; fi
	@$(LOCAL_BIN)/migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" up

migrate-down:
	@echo "Running migrations down..."
	@if [ -z "$(DB_DSN)" ]; then echo "Please specify DB_DSN"; exit 1; fi
	@$(LOCAL_BIN)/migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down

migrate-force:
	@echo "Forcing migration version..."
	@if [ -z "$(DB_DSN)" ]; then echo "Please specify DB_DSN"; exit 1; fi
	@if [ -z "$(version)" ]; then echo "Please specify version"; exit 1; fi
	@$(LOCAL_BIN)/migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" force $(version)


.DEFAULT_GOAL := build