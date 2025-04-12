# OpenAgent Makefile

# Configuration variables - update these for your environment
REMOTE_USER := root
REMOTE_HOST := in.msheriff.com
REMOTE_DIR := /root/github.com/openagent
REMOTE_CMD := cd $(REMOTE_DIR) && \
	echo 'Checking for old go-go-agent container...' && \
	docker stop go-go-agent || true && \
	docker rm go-go-agent || true && \
	docker stop openagent-service || true && \
	docker rm openagent-service || true && \
	docker compose down --remove-orphans && \
	echo 'Checking port 8800 after down...' && \
	fuser -k -n tcp 8800 || echo 'Port 8800 appears free or fuser failed.' && \
	sleep 3 && \
	docker compose build --no-cache && \
	docker compose up -d
GIT_REMOTE := origin
GIT_BRANCH := main
BACKUP_DIR := /root/github.com/openagent_backup

# Binary name
BINARY_NAME := openagent

# Include environment variables from .env file
include .env
export

.PHONY: all test build clean deploy deploy-git deploy-scp test-psql migrations fix-remote

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

# Deploy using Git (push and pull)
deploy: deploy-git

# Deploy using Git (push and pull)
deploy-git:
	@echo "Deploying using Git to $(REMOTE_HOST)..."
	@echo "1. Pushing changes to $(GIT_REMOTE)/$(GIT_BRANCH)..."
	git push $(GIT_REMOTE) $(GIT_BRANCH)
	
	@echo "2. Pulling changes and restarting containers on remote server..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_DIR) && git pull $(GIT_REMOTE) $(GIT_BRANCH) && $(REMOTE_CMD)"
	
	@echo "Deployment complete!"

# Deploy using SCP (legacy method)
deploy-scp:
	@echo "Deploying to $(REMOTE_HOST) using SCP..."
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
	docker compose build

# Run with Docker
docker-run:
	@echo "Starting with Docker..."
	docker compose up -d
	@echo "Services started! Check logs with: docker compose logs -f"

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	docker compose down

# Create data directory for SQLite
init:
	@echo "Creating required directories..."
	mkdir -p data tpl static models auth
	@echo "Initialization complete!"

# Fix remote repository by backing up untracked files and performing clean pull
fix-remote:
	@echo "Fixing remote repository at $(REMOTE_HOST)..."
	@echo "1. Creating backup directory..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "mkdir -p $(BACKUP_DIR)"
	
	@echo "2. Backing up untracked files..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_DIR) && \
		mv .env $(BACKUP_DIR)/ && \
		mv Dockerfile $(BACKUP_DIR)/ && \
		mv docker-compose.yml $(BACKUP_DIR)/ && \
		mv *.go $(BACKUP_DIR)/ 2>/dev/null || true && \
		mv go.* $(BACKUP_DIR)/ 2>/dev/null || true && \
		mv -f static/* $(BACKUP_DIR)/static/ 2>/dev/null || true && \
		mv -f tpl/* $(BACKUP_DIR)/tpl/ 2>/dev/null || true"
	
	@echo "3. Performing clean pull..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_DIR) && \
		git fetch $(GIT_REMOTE) && \
		git reset --hard $(GIT_REMOTE)/$(GIT_BRANCH) && \
		git clean -fd"
	
	# @echo "4. Restoring backup files..."
	# ssh $(REMOTE_USER)@$(REMOTE_HOST) "cd $(BACKUP_DIR) && \
	# 	cp -f .env $(REMOTE_DIR)/ && \
	# 	cp -f Dockerfile $(REMOTE_DIR)/ && \
	# 	cp -f docker-compose.yml $(REMOTE_DIR)/ && \
	# 	cp -f *.go $(REMOTE_DIR)/ 2>/dev/null || true && \
	# 	cp -f go.* $(REMOTE_DIR)/ 2>/dev/null || true && \
	# 	cp -rf static/* $(REMOTE_DIR)/static/ 2>/dev/null || true && \
	# 	cp -rf tpl/* $(REMOTE_DIR)/tpl/ 2>/dev/null || true"
	
	@echo "5. Restarting containers..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) $(REMOTE_CMD)
	
	@echo "Fix complete! Backup files are stored in $(BACKUP_DIR) on the remote server."

help:
	@echo "Available commands:"
	@echo "  make build      - Build the application"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make deploy     - Deploy using Git (default)"
	@echo "  make deploy-git - Deploy using Git"
	@echo "  make deploy-scp - Deploy using SCP (legacy)"
	@echo "  make fix-remote - Fix remote repository issues"
	@echo "  make docker-build - Build with Docker"
	@echo "  make docker-run - Run with Docker"
	@echo "  make docker-stop - Stop Docker containers"
	@echo "  make init       - Initialize directories"
	@echo "  make help       - Show this help"
