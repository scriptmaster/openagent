#!/bin/bash
# Script to be executed on the remote server for deployment

set -e # Exit immediately if a command exits with a non-zero status.

REMOTE_DIR="/root/github.com/openagent" # Define project directory
MASTER_ENV_PATH="/var/www/configs/openagent/.env" # Path to the master config
PROJECT_ENV_PATH="$REMOTE_DIR/.env" # Path to the project's .env

# --- Safety Check --- 
if [ ! -f "$MASTER_ENV_PATH" ]; then
    echo "Error: Master environment file not found at $MASTER_ENV_PATH" >&2
    exit 1
fi

cd "$REMOTE_DIR" || exit 1

# --- Update Repository --- 
echo 'Updating repository...'
git fetch origin
git reset --hard origin/main

# --- Copy Master Environment File --- 
echo "Copying master environment file from $MASTER_ENV_PATH to $PROJECT_ENV_PATH..."
cp -f "$MASTER_ENV_PATH" "$PROJECT_ENV_PATH"
if [ $? -ne 0 ]; then
    echo "Error: Failed to copy master environment file." >&2
    exit 1
fi
echo "Master environment file copied."

# --- Graceful Container Management --- 
echo 'Managing containers gracefully...'

# Check if PostgreSQL container is running
POSTGRES_CONTAINER=$(docker ps --filter "name=postgres-service" --format "{{.Names}}")
if [ -n "$POSTGRES_CONTAINER" ]; then
    echo "PostgreSQL container is running, preserving it..."
    # Build first to minimize downtime
    echo "Building new openagent-app image..."
    docker compose build openagent-app
    
    # Stop and start immediately
    echo "Stopping and starting openagent-service..."
    docker compose stop openagent-service
    docker compose rm -f openagent-service
    docker compose up -d openagent-service
else
    echo "PostgreSQL container not found, will start fresh..."
    # Stop all services if PostgreSQL isn't running
    docker compose down
    docker compose up -d --build
fi

# --- Verify Deployment --- 
echo 'Verifying deployment...'
sleep 3 # Give services time to start

# Check if app is running
APP_CONTAINER=$(docker ps --filter "name=openagent-service" --format "{{.Names}}")
if [ -z "$APP_CONTAINER" ]; then
    echo "Error: OpenAgent container failed to start" >&2
    exit 1
fi

echo "Deployment completed successfully!" 