# Laulo

A monorepo containing the Laulo media storage platform - a lightweight, multi-tenant storage system built with Go and Next.js, designed to run on Raspberry Pi.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-16-black?style=flat&logo=next.js)](https://nextjs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Monorepo Structure

```
laulo/
├── apps/
│   ├── api/          # Go backend (storage-api)
│   └── web/          # Next.js frontend
├── package.json      # Root workspace config
└── pnpm-workspace.yaml
```

## Quick Start

```bash
# Install dependencies
pnpm install

# Run web frontend
pnpm dev:web

# Run API (from apps/api directory)
cd apps/api
go run ./cmd/server
```

## Table of Contents

- [Description](#description)
- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [Running the Application](#running-the-application)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Deployment](#deployment)
- [Project Structure](#project-structure)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Description

Storage Pi is a RESTful API service for managing storage files in a multi-tenant environment. It provides household-based organization with user authentication, built with performance and simplicity in mind. The service is optimized to run efficiently on Raspberry Pi hardware while maintaining production-ready features.

## Features

- **Multi-tenant Architecture**: Household-based organization with user management
- **RESTful API**: Clean HTTP endpoints for storage operations
- **PostgreSQL Backend**: Reliable data persistence with connection pooling
- **Health Checks**: Application and database health monitoring endpoints
- **Database Migrations**: Version-controlled schema management using goose
- **Docker Support**: Easy deployment with Docker Compose
- **Lightweight**: Optimized for Raspberry Pi hardware
- **Development Mode**: Simple header-based authentication for local development

## Architecture

```mermaid
graph LR
    Client[HTTP Client]
    API[Storage Pi API]
    DB[(PostgreSQL)]

    Client -->|REST Requests| API
    API -->|pgx Pool| DB

    subgraph storage_pi [Storage Pi Server]
        API
    end

    subgraph database [Database Layer]
        DB
    end
```

### Data Model

The system uses a multi-tenant model organized by households:

- **Households**: Top-level organizational units
- **Users**: Belong to households with role-based access
- **Storage Items**: Stored and organized within household contexts

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go**: Version 1.24.0 or higher ([Download](https://golang.org/dl/))
- **PostgreSQL**: Version 16 or higher ([Download](https://www.postgresql.org/download/))
- **goose**: Database migration tool
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
- **Docker & Docker Compose** (optional, for containerized deployment)
  ```bash
  # On Raspberry Pi OS
  curl -sSL https://get.docker.com | sh
  sudo usermod -aG docker $USER
  ```

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/galosandoval/laulo.git
cd laulo
```

### 2. Install Dependencies

```bash
# Install Node.js dependencies (for web frontend)
pnpm install

# Install Go dependencies (for API)
cd apps/api
go mod download
```

### 3. Build the Applications

```bash
# Build web frontend
pnpm build:web

# Build API
cd apps/api
go build -o storage-api ./cmd/server
```

## Configuration

The application is configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `ADDR` | Server listen address | `:8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://storageapp:change_me_now@localhost:5432/storage_db?sslmode=disable` |

### Setting Environment Variables

Create a `.env` file in the `api` directory:

```bash
# api/.env
ADDR=:8080
DATABASE_URL=postgres://storageapp:your_secure_password@localhost:5432/storage_db?sslmode=disable
```

**Security Note**: Always change the default password in production environments.

## Database Setup

### Option 1: Docker Compose (Recommended)

Start PostgreSQL using the provided Docker Compose configuration:

```bash
cd infra
docker-compose up -d
```

This will:
- Start PostgreSQL 16 in a container
- Expose it on port 5432
- Create persistent storage volume
- Set up health checks

### Option 2: Manual PostgreSQL Installation

If you prefer a native installation:

```bash
# On Raspberry Pi OS
sudo apt update
sudo apt install postgresql-16

# Create database and user
sudo -u postgres psql
```

```sql
CREATE USER storageapp WITH PASSWORD 'change_me_now';
CREATE DATABASE storage_db OWNER storageapp;
GRANT ALL PRIVILEGES ON DATABASE storage_db TO storageapp;
```

### Running Migrations

Apply database migrations using the provided script:

```bash
chmod +x migrate.sh
./migrate.sh
```

Or manually with goose CLI:

```bash
cd api
export DATABASE_URL="postgres://storageapp:change_me_now@localhost:5432/storage_db?sslmode=disable"
goose -dir migrations postgres "$DATABASE_URL" up
```

## Running the Application

### Local Development

```bash
cd api
export DATABASE_URL="postgres://storageapp:change_me_now@localhost:5432/storage_db?sslmode=disable"
./storage-api
```

The API will start on `http://localhost:8080` (or the port specified in `ADDR`).

### Using systemd (Production)

Create a systemd service file for automatic startup:

```bash
sudo nano /etc/systemd/system/storage-api.service
```

```ini
[Unit]
Description=Storage API Service
After=network.target postgresql.service

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/storage-api
EnvironmentFile=/home/pi/storage-api/.env
ExecStart=/home/pi/storage-api/storage-api
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable storage-api
sudo systemctl start storage-api
sudo systemctl status storage-api
```

## API Documentation

### Endpoints

#### Health Check

Check if the API is running:

```bash
GET /health
```

**Example:**

```bash
curl http://localhost:8080/health
```

**Response:**

```json
{
  "status": "ok",
  "time": "2026-01-05T12:34:56Z"
}
```

#### Database Health Check

Verify database connectivity:

```bash
GET /health/db
```

**Example:**

```bash
curl http://localhost:8080/health/db
```

**Response (Success):**

```json
{
  "status": "db_ok"
}
```

**Response (Failure):**

```json
{
  "status": "db_unhealthy",
  "error": "connection timeout"
}
```

#### Get Current User

Retrieve user information (development mode):

```bash
GET /v1/me
```

**Headers:**
- `X-Dev-User`: External user identifier (development mode only)

**Example:**

```bash
curl -H "X-Dev-User: dev-alice" http://localhost:8080/v1/me
```

**Response:**

```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "householdId": "223e4567-e89b-12d3-a456-426614174111",
    "externalSub": "dev-alice",
    "email": "alice@example.com",
    "role": "admin",
    "createdAt": "2026-01-04T10:30:00Z"
  }
}
```

## Development

### Development Mode Authentication

The API currently uses a simplified authentication mechanism for development:

- Set the `X-Dev-User` header with any identifier (e.g., `dev-alice`)
- The API will look up the user by `external_sub` in the database
- Users must exist in the database before authentication

**Example:**

```bash
# Create a test user (using psql)
psql $DATABASE_URL -c "
  INSERT INTO users (id, household_id, external_sub, email, role)
  VALUES (
    gen_random_uuid(),
    (SELECT id FROM households LIMIT 1),
    'dev-alice',
    'alice@example.com',
    'admin'
  );"

# Test authentication
curl -H "X-Dev-User: dev-alice" http://localhost:8080/v1/me
```

### Running Tests

```bash
cd api
go test ./...
```

### Code Formatting

```bash
cd api
go fmt ./...
```

## Deployment

For automated deployment setup with GitHub Actions, see **[DEPLOYMENT.md](./DEPLOYMENT.md)** for detailed instructions.

### Quick Deployment

```bash
# Manual deployment on Pi
cd /home/pi/storage-api
./deploy.sh
```

### Raspberry Pi Considerations

#### Hardware Recommendations

- **Model**: Raspberry Pi 4 (4GB+ RAM recommended)
- **Storage**: External SSD for better I/O performance
- **Network**: Ethernet connection for stable connectivity

#### Performance Tuning

1. **PostgreSQL Configuration** (`postgresql.conf`):

```conf
# Optimize for Raspberry Pi
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
max_wal_size = 1GB
```

2. **Connection Pooling**: The application uses `pgxpool` with sensible defaults. Adjust pool size if needed based on concurrent load.

#### Security Considerations

1. **Change Default Credentials**: Update PostgreSQL password from default values
2. **Firewall Configuration**: Use `ufw` to restrict access
   ```bash
   sudo ufw allow 8080/tcp
   sudo ufw enable
   ```
3. **SSL/TLS**: Place a reverse proxy (nginx/caddy) in front of the API for HTTPS
4. **Authentication**: Replace development auth with proper OAuth2/OIDC in production

### Monitoring

Monitor the service with systemd:

```bash
# View logs
sudo journalctl -u storage-api -f

# Check status
sudo systemctl status storage-api

# Restart service
sudo systemctl restart storage-api
```

## Project Structure

```
laulo/
├── apps/
│   ├── api/                     # Go backend
│   │   ├── cmd/server/          # API entry point
│   │   ├── internal/            # Private application code
│   │   │   ├── config/          # Configuration loading
│   │   │   ├── models/          # Data models
│   │   │   ├── database/        # Database queries
│   │   │   ├── handlers/        # HTTP handlers
│   │   │   └── server/          # HTTP server setup
│   │   ├── migrations/          # SQL migration files
│   │   ├── go.mod               # Go module definition
│   │   └── docker-compose.yml   # PostgreSQL container
│   └── web/                     # Next.js frontend
│       ├── app/                 # Next.js app router
│       ├── components/          # React components
│       ├── hooks/               # Custom React hooks
│       ├── lib/                 # Utilities and API client
│       └── package.json         # Web dependencies
├── .github/workflows/           # GitHub Actions
├── package.json                 # Root workspace config
├── pnpm-workspace.yaml          # pnpm workspace definition
└── README.md                    # This file
```

## Troubleshooting

### Database Connection Issues

**Problem**: `db connect error: connection refused`

**Solution**:
1. Verify PostgreSQL is running:
   ```bash
   sudo systemctl status postgresql
   # Or for Docker
   docker ps | grep postgres
   ```
2. Check `DATABASE_URL` is correct
3. Verify network connectivity: `pg_isready -h localhost`

### Port Already in Use

**Problem**: `bind: address already in use`

**Solution**:
1. Check what's using port 8080:
   ```bash
   sudo lsof -i :8080
   ```
2. Either stop the other service or change `ADDR` to use a different port

### Migration Failures

**Problem**: `goose: no such file or directory`

**Solution**:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### Permission Denied on Raspberry Pi

**Problem**: Cannot execute binary

**Solution**:
```bash
chmod +x api/storage-api
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Commit your changes: `git commit -am 'Add new feature'`
4. Push to the branch: `git push origin feature/your-feature`
5. Submit a pull request

### Code Style

- Follow Go conventions and `gofmt` formatting
- Write clear commit messages
- Add tests for new features
- Update documentation as needed

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Note**: This project is optimized for Raspberry Pi but can run on any system with Go and PostgreSQL support.

For questions or issues, please open an issue on the repository.

