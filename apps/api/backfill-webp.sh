#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(dirname "$0")"

# Load environment variables
if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a
  source "$SCRIPT_DIR/.env"
  set +a
elif [ -f "$SCRIPT_DIR/../../.env" ]; then
  set -a
  source "$SCRIPT_DIR/../../.env"
  set +a
else
  echo "No .env file found. Make sure DATABASE_URL and MEDIA_PATH are set."
fi

echo "Running WebP backfill..."
echo "MEDIA_PATH: ${MEDIA_PATH:-/mnt/storage/media}"

cd "$SCRIPT_DIR"

# Use compiled binary if available, otherwise fall back to go run
if [ -f "./backfill-webp" ]; then
  ./backfill-webp
else
  go run ./cmd/backfill-webp/main.go
fi
