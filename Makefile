.PHONY: all build run test clean db-up db-down migrate-up migrate-down

# Build variables
BINARY_NAME=mono-e-mono
BUILD_DIR=bin

all: build

# Build the application
build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run the application
run:
	@go run ./cmd/server

# Run tests
test:
	@go test -v ./...

# Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)

# Start database
db-up:
	@docker-compose up -d postgres redis

# Stop database
db-down:
	@docker-compose down

# Run migrations up (requires golang-migrate CLI)
migrate-up:
	@migrate -path internal/db/migrations -database "$${DATABASE_URL}" up

# Run migrations down (requires golang-migrate CLI)
migrate-down:
	@migrate -path internal/db/migrations -database "$${DATABASE_URL}" down

# Generate sqlc code
sqlc:
	@cd internal/db && sqlc generate

# Development: start db and run server
dev: db-up
	@sleep 2
	@go run ./cmd/server

# Tidy dependencies
tidy:
	@go mod tidy
