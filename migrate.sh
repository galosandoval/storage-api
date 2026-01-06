#!/usr/bin/env bash
set -euo pipefail

# Load env vars from .env in project root
set -a
source "$(dirname "$0")/.env"
set +a

goose -dir "$(dirname "$0")/migrations" postgres "$DATABASE_URL" up
