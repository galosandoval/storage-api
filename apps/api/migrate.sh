#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(dirname "$0")"

# Try to load .env from api directory first, then project root
if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a
  source "$SCRIPT_DIR/.env"
  set +a
elif [ -f "$SCRIPT_DIR/../../.env" ]; then
  set -a
  source "$SCRIPT_DIR/../../.env"
  set +a
else
  echo "No .env file found. Make sure DATABASE_URL is set."
fi

goose -dir "$SCRIPT_DIR/migrations" postgres "$DATABASE_URL" up