# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy module files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code from root and subdirectories
COPY . .

# Build the Go binary for package main, including all its files
RUN go build -ldflags="-s -w" -o /app/server .

# Stage 2: Create the runtime image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies:
# - ca-certificates: For HTTPS communication
# - bash: For shell commands
# - curl: Common utility
RUN apk update && apk add --no-cache ca-certificates bash curl

# Create the data directory (if still needed for other runtime data)
RUN mkdir -p /app/data/sql && chmod -R 755 /app/data # Ensure data/sql exists

# Copy the built Go binary from the builder stage
COPY --from=builder /app/server /app/server

# Copy HTML templates and static assets
COPY tpl/ /app/tpl/
COPY static/ /app/static/

# Copy SQL files from data/sql to /app/data/sql/
COPY data/sql/ /app/data/sql/

# Copy .env file to app root
COPY .env /app/.env

# Declare the volume mount point (if needed for runtime data, keep it)
VOLUME /app/data

# Expose the port the application listens on
EXPOSE 8800

# Set default environment variables
ENV PORT=8800
ENV DATA_DIR=/app/data
ENV SQL_DIR=/app/data/sql # Point SQL loader to the new image location

# Command to run when the container starts
CMD ["/app/server"] 