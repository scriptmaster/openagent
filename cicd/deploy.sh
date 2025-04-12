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

# --- Cleanup Old/Conflicting Containers --- 
echo 'Checking for old go-go-agent container...'
docker stop go-go-agent > /dev/null 2>&1 || true
docker rm go-go-agent > /dev/null 2>&1 || true

echo 'Finding and stopping container using host port 8800...'
CONTAINER_ID=$(docker ps --filter "publish=8800" --format "{{.ID}}")
if [ -n "$CONTAINER_ID" ]; then
    echo "Stopping and removing container $CONTAINER_ID using port 8800..."
    docker stop "$CONTAINER_ID" > /dev/null 2>&1 || true
    docker rm "$CONTAINER_ID" > /dev/null 2>&1 || true
else
    echo 'No container found using host port 8800.'
fi

# --- Docker Compose Operations --- 
docker compose down --remove-orphans

echo 'Checking port 8800 after down with fuser...'
fuser -k -n tcp 8800 || echo 'Port 8800 appears free.'
sleep 3

echo 'Building and starting services...'
docker compose build --no-cache
docker compose up -d

echo "Remote deployment steps completed successfully." 