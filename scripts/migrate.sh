#!/usr/bin/env bash
# Apply or roll back SQL migrations against the SyncGuard database.
# Usage: ./scripts/migrate.sh [up|down]
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="${DIR}/../migrations"

DB_HOST="${SYNCGUARD_DB_HOST:-localhost}"
DB_PORT="${SYNCGUARD_DB_PORT:-5432}"
DB_NAME="${SYNCGUARD_DB_NAME:-syncguard}"
DB_USER="${SYNCGUARD_DB_USER:-syncguard}"
DB_PASS="${SYNCGUARD_DB_PASSWORD:-changeme}"

export PGPASSWORD="${DB_PASS}"
PSQL="psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME}"

direction="${1:-up}"

case "${direction}" in
  up)
    echo "Applying migrations..."
    for f in "${MIGRATIONS_DIR}"/*.up.sql; do
      echo "  → ${f##*/}"
      ${PSQL} -f "${f}"
    done
    echo "Done."
    ;;
  down)
    echo "Rolling back last migration..."
    # Apply down scripts in reverse order
    for f in $(ls "${MIGRATIONS_DIR}"/*.down.sql | sort -r); do
      echo "  ← ${f##*/}"
      ${PSQL} -f "${f}"
      break  # only roll back one at a time
    done
    echo "Done."
    ;;
  *)
    echo "Usage: $0 [up|down]" >&2
    exit 1
    ;;
esac
