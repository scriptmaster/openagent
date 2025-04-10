# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy only necessary files first to leverage Docker cache
# If you add go.mod/go.sum later, uncomment these lines:
# COPY go.mod go.sum ./
# RUN go mod download

COPY agent.go .

# Build the Go binary statically linked (optional but good for minimal images)
# CGO_ENABLED=0 disables CGo, forcing static linking for standard libs
# -ldflags="-s -w" strips debugging symbols and info, reducing binary size
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/agent agent.go

# Stage 2: Create the minimal runtime image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies:
# - ca-certificates: For HTTPS communication (e.g., if agent needs to curl websites)
# - bash: Often useful for shell commands the agent might run (sh is default, bash is richer)
# - curl: A common tool the agent might need to use
RUN apk update && apk add --no-cache ca-certificates bash curl

# Create the data directory where the agent will work and volume will be mounted
# Use 'nobody' user/group which exists in Alpine and is non-root
RUN mkdir -p /app/data && chown nobody:nogroup /app/data

# Copy the built Go binary from the builder stage
COPY --from=builder /app/agent /app/agent

# Copy the HTML template
COPY index.html /app/index.html

# Set permissions for copied files (optional, root owns by default)
# RUN chown nobody:nogroup /app/agent /app/index.html

# Switch to non-root user for better security.
# Commands executed by the agent will run as this user.
# Ensure this user has write permissions to /app/data (done above)
USER nobody:nogroup

# Declare the volume mount point (actual mapping done in docker-compose)
VOLUME /app/data

# Expose the port the Go application listens on
EXPOSE 8080

# Set default environment variables (can be overridden in docker-compose)
# host.docker.internal is for Docker Desktop (Mac/Win)
# For Linux, this might need to be changed or network_mode: host used
ENV OLLAMA_URL=http://host.docker.internal:11434
ENV OLLAMA_MODEL=llama3:8b
ENV PORT=8080

# Command to run when the container starts
CMD ["/app/agent"]

