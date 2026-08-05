#!/bin/bash
set -e

DATA_DIR="${TELBOT_DATA_DIR:-/data}"

# Auto-load .env from persistent volume if it exists
if [ -f "$DATA_DIR/.env" ]; then
  echo "📄 Loading .env from $DATA_DIR/.env"
  set -a
  . "$DATA_DIR/.env"
  set +a
fi

# Ensure persistent directory exists for sessions
mkdir -p "$DATA_DIR"

echo "🚀 Starting telbot with args: $@"
exec /app/telbot "$@"
