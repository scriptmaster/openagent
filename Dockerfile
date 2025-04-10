# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files
COPY *.go ./
COPY models/ ./models/
COPY auth/ ./auth/

# Build the Go binary without CGO
RUN go build -ldflags="-s -w" -o /app/server .

# Stage 2: Create the runtime image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies:
# - ca-certificates: For HTTPS communication
# - bash: For shell commands
# - curl: Common utility
RUN apk update && apk add --no-cache ca-certificates bash curl

# Create the data directory where the application will store data
RUN mkdir -p /app/data && chmod 755 /app/data

# Copy the built Go binary from the builder stage
COPY --from=builder /app/server /app/server

# Copy HTML templates and static assets
COPY tpl/ /app/tpl/
COPY static/ /app/static/

# Copy .env file if it exists (will be overridden by environment variables)
COPY .env* /app/

# Declare the volume mount point
VOLUME /app/data

# Expose the port the application listens on
EXPOSE 8800

# Set default environment variables
ENV PORT=8800
ENV DATA_DIR=/app/data

# Command to run when the container starts
CMD ["/app/server"]

