#!/usr/bin/env bash
set -euo pipefail

echo "🚀 Starting deployment..."

# Change to project directory
cd "$(dirname "$0")"

echo "📦 Pulling latest code..."
git pull origin main

echo "🔨 Building application..."
go build -o storage-api ./cmd/server
go build -o backfill-webp ./cmd/backfill-webp

echo "🗄️  Running migrations..."
./migrate.sh

echo "🔄 Restarting service..."
sudo systemctl restart storage-api

# Wait a moment for service to start
sleep 2

# Check service status
if sudo systemctl is-active --quiet storage-api; then
    echo "✅ Service is running successfully"

    # Test health endpoint
    if curl -f http://localhost:8080/health/live >/dev/null 2>&1; then
        echo "✅ Health check passed"
    else
        echo "⚠️  Warning: Health check failed"
    fi
else
    echo "❌ Service failed to start"
    sudo systemctl status storage-api
    exit 1
fi

echo "🎉 Deployment complete!"


