.PHONY: all build proto test clean up down logs

# Default target
all: build

# Build all application binaries
build: proto
	@echo "Building job-server..."
	@go build -o bin/job-server ./cmd/job-server
	@echo "Building qgjob CLI..."
	@go build -o bin/qgjob ./cmd/qgjob
	@echo "Build complete."

# Generate Go code from protobuf definitions
proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        api/proto/job_service.proto

# Run the integration test suite
test:
	@echo "Running integration test suite..."
	@./scripts/test_integration.sh

# Clean up build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin
	@rm -f job-server.log

# Start dependent services (Postgres, Redis)
up:
	@echo "Starting services..."
	@docker-compose up -d

# Stop dependent services
down:
	@echo "Stopping services..."
	@docker-compose down

# View logs from services
logs:
	@docker-compose logs -f
