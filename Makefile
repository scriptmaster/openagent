# OpenAgent Makefile

# Configuration variables - update these for your environment
REMOTE_USER := root
REMOTE_HOST := in.msheriff.com
REMOTE_DIR := /root/github.com/openagent
REMOTE_CMD := "cd $(REMOTE_DIR) && docker-compose down && docker-compose build && docker-compose up -d"

# Binary name
BINARY_NAME := openagent

.PHONY: all test build clean deploy test-psql

all: test build

# Test PSQL connection
test-psql:
	@echo "Testing PostgreSQL connection..."
	@psql -c "SELECT version();" || { echo "PostgreSQL connection failed! Ensure psql is installed and properly configured."; exit 1; }
	@echo "PostgreSQL connection successful!"

# Build the application
build: test-psql
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "Testing local build..."
	go run . & \
	sleep 5 && \
	curl -s http://localhost:8800/ > /dev/null && \
	echo "Test successful, server is responding!" || echo "Test failed, server not responding"
	@pkill $(BINARY_NAME) || true
	@echo "Test completed"

# Clean build files
clean:
	@echo "Cleaning build files..."
	rm -f $(BINARY_NAME)
	go clean

# Deploy to remote server
deploy:
	@echo "Deploying to $(REMOTE_HOST)..."
	@echo "1. Copying files to remote server..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "mkdir -p $(REMOTE_DIR)/tpl $(REMOTE_DIR)/static $(REMOTE_DIR)/data"
	scp -r *.go *.mod *.sum .env Dockerfile docker-compose.yml $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/
	scp -r tpl/* $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/tpl/
	scp -r static/* $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/static/
	
	@echo "2. Restarting containers on remote server..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) $(REMOTE_CMD)
	
	@echo "Deployment complete!"

# Build for production with Docker
docker-build:
	@echo "Building Docker image..."
	docker-compose build

# Run with Docker
docker-run:
	@echo "Starting with Docker..."
	docker-compose up -d
	@echo "Services started! Check logs with: docker-compose logs -f"

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# Create data directory for SQLite
init:
	@echo "Creating required directories..."
	mkdir -p data tpl static models auth
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
