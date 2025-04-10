# Go OpenAgent Makefile

# Configuration variables - update these for your environment
REMOTE_USER := root
REMOTE_HOST := in.msheriff.com
REMOTE_DIR := /github.com/openagent
REMOTE_CMD := "cd $(REMOTE_DIR) && docker-compose down && docker-compose build && docker-compose up -d"

# Binary name
BINARY_NAME := openagent

# Default environment settings
export CGO_ENABLED := 1

.PHONY: all test build clean deploy

all: test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	cd go && go build -o ../$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests with CGO enabled for SQLite support..."
	cd go && CGO_ENABLED=1 go test -v ./...
	@echo "Testing local build with SQLite database..."
	cd go && CGO_ENABLED=1 go run . & \
	sleep 5 && \
	curl -s http://localhost:8080/ > /dev/null && \
	echo "Test successful, server is responding!" || echo "Test failed, server not responding"
	@pkill $(BINARY_NAME) || true
	@echo "Test completed"

# Clean build files
clean:
	@echo "Cleaning build files..."
	rm -f $(BINARY_NAME)
	cd go && go clean

# Deploy to remote server
deploy:
	@echo "Deploying to $(REMOTE_HOST)..."
	@echo "1. Copying files to remote server..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "mkdir -p $(REMOTE_DIR)"
	scp -r go/* $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/
	
	@echo "2. Restarting containers on remote server..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) $(REMOTE_CMD)
	
	@echo "Deployment complete!"

# Build for production with Docker
docker-build:
	@echo "Building Docker image..."
	cd go && docker-compose build

# Run with Docker
docker-run:
	@echo "Starting with Docker..."
	cd go && docker-compose up -d
	@echo "Services started! Check logs with: docker-compose logs -f"

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	cd go && docker-compose down

# Create data directory for SQLite
init:
	@echo "Creating data directory..."
	mkdir -p go/data
	@echo "Initialization complete!"

help:
	@echo "Available commands:"
	@echo "  make build      - Build the application"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make deploy     - Deploy to remote server (update REMOTE_* variables first)"
	@echo "  make docker-build - Build with Docker"
	@echo "  make docker-run - Run with Docker"
	@echo "  make docker-stop - Stop Docker containers"
	@echo "  make init       - Initialize directories"
	@echo "  make help       - Show this help"
