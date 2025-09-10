# OpenAgent Makefile

# Configuration variables - update these for your environment
REMOTE_USER := root
REMOTE_HOST := in.msheriff.com
REMOTE_DIR := /root/github.com/openagent
DEPLOY_PATH := $(REMOTE_DIR)/cicd
REMOTE_CMD := "cd $(DEPLOY_PATH) && chmod +x deploy.sh && ./deploy.sh $(VERSION) $(LATEST_COMMIT)" # Command to run deploy.sh in DEPLOY_PATH
GIT_REMOTE := origin
GIT_BRANCH := main
BACKUP_DIR := /root/github.com/openagent_backup

# Binary name
BINARY_NAME := openagent

# Include environment variables from .env file
include .env
export

.PHONY: all test build clean deploy deploy-git deploy-scp test-psql fix-remote stop migrations

all: test build

# Test PSQL connection
test-psql:
	@echo "Testing PostgreSQL connection..."
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c "SELECT version();" >/dev/null 2>&1 || (echo "Error: PostgreSQL connection failed!" && exit 1)
	@echo "PostgreSQL connection successful."

# Build the application
build: test-psql
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .

# Manually apply database migrations (Original Method)
migrations:
	@echo "Applying migrations manually..."
	# Ensure the Go binary exists
	@[ -f $(BINARY_NAME) ] || go build -o $(BINARY_NAME) .
	# Execute the migration command within your app (assuming a command or flag exists)
	# Example: ./$(BINARY_NAME) --migrate 
	# OR, if migrations are applied on startup based on MIGRATION_START:
	@echo "Running application to apply migrations based on MIGRATION_START..."
	./$(BINARY_NAME)
	@echo "Migrations check/apply completed."

# Run tests (includes running migrations first if needed)
test:
	make stop
	# Optionally run migrations before testing if tests depend on schema
	# make migrations
	@echo "Running go mod tidy..."
	go mod tidy
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

deploy: deploy-scp
# deploy: deploy-git

# Deploy to server using git commit hash for versioning
deploy-git: build
	@echo "Deploying to server using Git..."
	@LATEST_COMMIT=$$(git rev-parse HEAD);
	@echo "Latest commit: $${LATEST_COMMIT}"
	@VERSION=$$(git describe --tags --always --dirty)
	@echo "Deploying version: $${VERSION}"
	# Ensure remote directory exists and has proper permissions
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "mkdir -p $(DEPLOY_PATH) && chmod 755 $(DEPLOY_PATH)"
	# Ensure deploy.sh has execute permissions
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(DEPLOY_PATH) && chmod +x deploy.sh"
	# Execute the remote command
	# IMPORTANT: Ensure .env file is securely managed on the remote server in $(DEPLOY_PATH)
	# IMPORTANT: Assumes deploy.sh knows docker-compose.yml is in $(DEPLOY_PATH)
	ssh $(REMOTE_USER)@$(REMOTE_HOST) $(REMOTE_CMD)
	@echo "Deployment complete."

# Deploy using SCP (legacy method)
deploy-scp:
	@echo "Deploying using SCP to $(REMOTE_HOST)..."
	
	# Copy Dockerfile from root, docker-compose.yml from cicd/
	scp Dockerfile $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/
	
	# scp cicd/docker-compose.yml $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/cicd/ # Copy to cicd subdir on remote
	scp docker-compose.yml $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/ # Copy to cicd subdir on remote
	
	# Use rsync for other files, excluding .git, .env, etc.
	rsync -avz --exclude '.git' --exclude '.env' --exclude '.idea' --exclude '.vscode' --exclude 'node_modules' --exclude '*.log' ./ $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/
	
	@echo "Executing remote build and run..."
	# IMPORTANT: Ensure .env file is securely managed on the remote server in $(REMOTE_DIR)
	# Run docker compose using the compose file in cicd/
	
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_DIR) && docker compose -f docker-compose.yml up -d --build"
	@echo "SCP Deployment complete!"

# Docker commands
# Use -f to specify the compose file location in cicd/
docker-build:
	@echo "Building Docker images using cicd/docker-compose.yml..."
	docker compose -f cicd/docker-compose.yml build

docker-run: docker-build
	@echo "Starting Docker containers using cicd/docker-compose.yml..."
	docker compose -f cicd/docker-compose.yml up -d

docker-stop:
	@echo "Stopping Docker containers using cicd/docker-compose.yml..."
	docker compose -f cicd/docker-compose.yml down

# Create data directory for SQLite
init:
	@echo "Creating required directories..."
	mkdir -p data tpl static models auth
	@echo "Initialization complete!"

update-deps:
	go clean -modcache
	go mod tidy

# Fix remote repository by backing up untracked files and performing clean pull
fix-remote:
	@echo "Fixing remote repository at $(REMOTE_HOST)..."
	@echo "1. Creating backup directory..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "mkdir -p $(BACKUP_DIR)"
	
	@echo "2. Backing up potentially modified files (excluding .env)..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_DIR) && \
		mv Dockerfile $(BACKUP_DIR)/ 2>/dev/null || true && \
		mv cicd/docker-compose.yml $(BACKUP_DIR)/cicd/ 2>/dev/null || true && \
		mv *.go $(BACKUP_DIR)/ 2>/dev/null || true && \
		mv go.* $(BACKUP_DIR)/ 2>/dev/null || true && \
		mv -f static/* $(BACKUP_DIR)/static/ 2>/dev/null || true && \
		mv -f tpl/* $(BACKUP_DIR)/tpl/ 2>/dev/null || true" 
	
	@echo "3. Performing clean pull..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_DIR) && \
		git fetch $(GIT_REMOTE) && \
		git reset --hard $(GIT_REMOTE)/$(GIT_BRANCH) && \
		git clean -fdx" # Use -x to remove ignored files too (like local .env, deploy.sh)
	
	@echo "4. Restoring required files from backup (excluding .env)..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(BACKUP_DIR) && \
		cp -f Dockerfile $(REMOTE_DIR)/ 2>/dev/null || true && \
		mkdir -p $(REMOTE_DIR)/cicd && cp -f cicd/docker-compose.yml $(REMOTE_DIR)/cicd/ 2>/dev/null || true && \
		cp -f *.go $(REMOTE_DIR)/ 2>/dev/null || true && \
		cp -f go.* $(REMOTE_DIR)/ 2>/dev/null || true && \
		cp -rf static/* $(REMOTE_DIR)/static/ 2>/dev/null || true && \
		cp -rf tpl/* $(REMOTE_DIR)/tpl/ 2>/dev/null || true"
	
	@echo "5. Running standard deployment steps..."
	make deploy # Rerun the normal deploy which copies deploy.sh and runs it
	
	@echo "Fix complete! Backup files are stored in $(BACKUP_DIR) on the remote server."

# Stop all running openagent processes
stop:
	@echo "Stopping all openagent.exe processes..."
	@while true; do \
		PID=$$(tasklist /NH /FI "IMAGENAME eq openagent.exe" 2>nul | awk '{print $$2}' | head -n 1); \
		if [ -z "$$PID" ] || [ "$$PID" = "No" ]; then \
			echo "All openagent.exe processes stopped."; \
			break; \
		fi; \
		echo "Stopping PID: $$PID"; \
		taskkill /PID $$PID /F /T >nul 2>&1 || true; \
		sleep 1; \
	done

help:
	@echo "Available commands:"
	@echo "  make build      - Build the application"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make migrations - Apply database migrations manually (runs the app)"
	@echo "  make deploy     - Deploy using Git (uses deploy.sh on remote)"
	@echo "  make deploy-git - Deploy using Git"
	@echo "  make deploy-scp - Deploy using SCP (legacy, uses deploy.sh on remote)"
	@echo "  make fix-remote - Fix remote repository issues (excluding .env)"
	@echo "  make docker-build - Build with Docker"
	@echo "  make docker-run - Run with Docker"
	@echo "  make docker-stop - Stop Docker containers"
	@echo "  make init       - Initialize directories"
	@echo "  make help       - Show this help"
	@echo "  make update-deps - Update dependencies"
	@echo "  make stop       - Stop all running openagent processes"
