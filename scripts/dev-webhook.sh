#!/usr/bin/env bash
# dev-webhook.sh — Start the PBOM webhook listener with smee proxy for local development.
#
# Prerequisites:
#   npm install -g smee-client
#
# Required environment variables:
#   SMEE_URL             — Your smee.io channel URL (get one at https://smee.io/new)
#   PBOM_WEBHOOK_SECRET  — The webhook secret configured in GitHub org settings
#   GITHUB_TOKEN         — A GitHub PAT with repo and read:org scopes
#
# Usage:
#   export SMEE_URL=https://smee.io/YourChannelHere
#   export PBOM_WEBHOOK_SECRET=your-secret
#   export GITHUB_TOKEN=ghp_...
#   ./scripts/dev-webhook.sh

set -euo pipefail

: "${SMEE_URL:?Set SMEE_URL to your smee.io channel URL (get one at https://smee.io/new)}"
: "${PBOM_WEBHOOK_SECRET:?Set PBOM_WEBHOOK_SECRET to the webhook secret from GitHub org settings}"
: "${GITHUB_TOKEN:?Set GITHUB_TOKEN to a GitHub PAT with repo and read:org scopes}"

STORAGE_DIR="${PBOM_STORAGE_DIR:-./pbom-data}"
ADDR="${PBOM_WEBHOOK_ADDR:-:8080}"

# Build the binary
echo "==> Building PBOM CLI..."
go build -o bin/pbom ./cmd/pbom

# Create storage directory
mkdir -p "$STORAGE_DIR"

# Start smee proxy in background
echo "==> Starting smee proxy: $SMEE_URL -> http://localhost${ADDR}/webhook"
smee -u "$SMEE_URL" --target "http://localhost${ADDR}/webhook" &
SMEE_PID=$!
trap "echo '==> Stopping smee proxy...'; kill $SMEE_PID 2>/dev/null" EXIT

# Give smee a moment to connect
sleep 1

# Start webhook listener
echo "==> Starting webhook listener on $ADDR (storage: $STORAGE_DIR)"
echo ""
./bin/pbom webhook \
  --addr "$ADDR" \
  --secret "$PBOM_WEBHOOK_SECRET" \
  --token "$GITHUB_TOKEN" \
  --storage-dir "$STORAGE_DIR"
