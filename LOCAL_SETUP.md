# Local Development Setup

This guide helps you set up and test the storage API on your local Mac.

## Prerequisites

- ✅ Go installed (verify with `go version`)
- ✅ Docker installed (for PostgreSQL)
- ✅ goose migration tool installed (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- ✅ `$HOME/go/bin` in your PATH

## Local Setup Steps

### 1. Environment Variables

Create a `.env` file in the project root (already done if migrations worked):

```bash
POSTGRES_USER=storageapp
POSTGRES_PASSWORD=change_me_now
POSTGRES_DB=storage_db
DATABASE_URL=postgres://storageapp:change_me_now@localhost:5432/storage_db?sslmode=disable
```

### 2. Start PostgreSQL

```bash
docker-compose up -d
```

Verify it's running:
```bash
docker ps
# Should show storage-postgres container
```

### 3. Run Migrations

```bash
./migrate.sh
```

Expected output: `goose: no migrations to run` or `OK <timestamp>`

### 4. Build the Application

```bash
go build -o storage-api ./cmd/server
```

This creates a `storage-api` binary in the project root.

### 5. Run the API Server

```bash
./storage-api
```

Expected output:
```
API listening on :8080
```

### 6. Test the Endpoints

In another terminal, test the health endpoints:

```bash
# Basic health check
curl http://localhost:8080/health
# Expected: {"status":"ok","time":"2026-01-07T..."}

# Database health check
curl http://localhost:8080/health/db
# Expected: {"status":"db_ok"}
```

### 7. Test with Environment Variables

You can override the default port or database URL:

```bash
# Run on different port
ADDR=:3000 ./storage-api

# Or source the .env file
source .env && ./storage-api
```

## Troubleshooting

### PostgreSQL Connection Issues

```bash
# Check if PostgreSQL is running
docker ps

# Check PostgreSQL logs
docker logs storage-postgres

# Restart PostgreSQL
docker-compose down
docker-compose up -d
```

### Migration Errors

```bash
# Check migration status
goose -dir migrations postgres "$DATABASE_URL" status

# Rollback last migration if needed
goose -dir migrations postgres "$DATABASE_URL" down
```

### Build Errors

```bash
# Update dependencies
go mod tidy
go mod download

# Clean build cache
go clean -cache
```

## Next Steps

Once local testing is complete, proceed to [GITHUB_ACTIONS_SETUP.md](GITHUB_ACTIONS_SETUP.md) to configure automated deployment to your Raspberry Pi.
