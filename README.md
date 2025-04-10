# OpenAgent

OpenAgent is a web-based authentication and database management system with OTP (One-Time Password) email authentication.

## Project Structure

```
openagent/
├── auth/            # Authentication related code
├── models/          # Database models 
├── tpl/             # HTML templates
│   ├── login.html   # Login page with OTP authentication
│   ├── index.html   # Main dashboard page
│   ├── admin.html   # Admin dashboard 
│   └── ...
├── static/          # Static assets (CSS, JS, images)
│   ├── css/
│   ├── js/
│   └── img/
├── data/            # Data storage (SQLite DB, other files)
├── *.go             # Main Go source files
├── go.mod           # Go module file
├── go.sum           # Go module checksums
├── Dockerfile       # Docker build instructions
├── docker-compose.yml # Docker Compose configuration
├── Makefile         # Build and deployment automation
└── README.md        # This file
```

## Setup and Usage

### Prerequisites

- Go 1.18+ with CGO enabled
- Docker and Docker Compose (for containerized deployment)
- SQLite (for local development without PostgreSQL)
- PostgreSQL (optional, for production)

### Local Development

1. Initialize the project:
   ```
   make init
   ```

2. Test the application:
   ```
   make test
   ```

3. Build the application:
   ```
   make build
   ```

4. Run locally (uses SQLite by default):
   ```
   cd go && CGO_ENABLED=1 go run .
   ```

### Docker Deployment

1. Build the Docker image:
   ```
   make docker-build
   ```

2. Run with Docker:
   ```
   make docker-run
   ```

3. Stop the containers:
   ```
   make docker-stop
   ```

### Remote Deployment

1. Edit the Makefile to set your remote server details:
   ```
   REMOTE_USER := your-username
   REMOTE_HOST := your-server.com
   REMOTE_DIR := /path/to/deployment
   ```

2. Deploy to the remote server:
   ```
   make deploy
   ```

This will copy all necessary files to the remote server and restart the Docker containers.

## Configuration

The application can be configured using environment variables or a `.env` file. Key configuration options:

- `PORT`: HTTP server port (default: 8080)
- `DATA_DIR`: Directory for data storage (default: ./data or /app/data in Docker)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL connection details
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`: Email server settings

## Email OTP Authentication

The system uses email-based One-Time Passwords for authentication:
1. User enters their email
2. System sends a 6-digit OTP to their email
3. User enters the OTP to authenticate
4. System creates a session for the user
