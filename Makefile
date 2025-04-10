# OpenAgent Makefile

# Configuration variables - update these for your environment
REMOTE_USER := root
REMOTE_HOST := in.msheriff.com
REMOTE_DIR := /root/github.com/openagent
REMOTE_CMD := "cd $(REMOTE_DIR) && docker-compose down && docker-compose build && docker-compose up -d"

# Binary name
BINARY_NAME := openagent

# Include environment variables from .env file
include .env
export

.PHONY: all test build clean deploy test-psql migrations

all: test build

# Test PSQL connection
test-psql:
	@echo "Testing PostgreSQL connection..."
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c "SELECT version();" >/dev/null 2>&1 || (echo "Error: PostgreSQL connection failed!" && exit 1)
	@echo "PostgreSQL connection successful."

# Apply database migrations
migrations:
	@echo "Applying database migrations..."
	@# Get the current highest migration number
	@LAST_APPLIED=$$(grep -oP "MIGRATION_START=\K\d+" .env 2>/dev/null || echo "0"); \
	HIGHEST_APPLIED=$$LAST_APPLIED; \
	for file in migrations/[0-9][0-9][0-9]_*.sql; do \
		NUM=$$(echo $${file} | grep -oP "migrations/\K\d+"); \
		if [ "$$NUM" -gt "$$LAST_APPLIED" ]; then \
			echo "Applying $${file}..."; \
			PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $${file} || exit 1; \
			if [ "$$NUM" -gt "$$HIGHEST_APPLIED" ]; then \
				HIGHEST_APPLIED=$$NUM; \
			fi; \
		else \
			echo "Skipping $${file} (already applied)"; \
		fi; \
	done; \
	if [ "$$HIGHEST_APPLIED" -gt "$$LAST_APPLIED" ]; then \
		echo "Updating MIGRATION_START to $$HIGHEST_APPLIED in .env"; \
		sed -i.bak "/MIGRATION_START=/d" .env || true; \
		echo "MIGRATION_START=$$HIGHEST_APPLIED" >> .env; \
		rm -f .env.bak 2>/dev/null || true; \
	fi
	@echo "Migrations complete!"

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
	PID=$$! && \
	sleep 5 && \
	curl -s http://localhost:8800/ > /dev/null && \
	echo "Test successful, server is responding!" || echo "Test failed, server not responding"; \
	kill -9 $$PID 2>/dev/null || true
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
