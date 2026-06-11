#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi

if ! command -v goose >/dev/null 2>&1; then
  echo "goose is not installed; run: go install github.com/pressly/goose/v3/cmd/goose@v3.24.1" >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec goose -dir "$ROOT_DIR/database/migrations" postgres "$DATABASE_URL" up
