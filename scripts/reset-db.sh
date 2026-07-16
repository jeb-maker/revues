#!/usr/bin/env bash
# Supprime la base SQLite locale et laisse goose recréer au prochain démarrage.
# Usage: ./scripts/reset-db.sh [--seed]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DB="${REVUES_DATABASE_PATH:-$ROOT/data/revues.db}"

rm -f "$DB" "${DB}-wal" "${DB}-shm"
echo "Base supprimée : $DB"

if [[ "${1:-}" == "--seed" ]]; then
  echo "Seed..."
  (cd "$ROOT" && go run ./cmd/seed/main.go)
else
  echo "Relancer : go run ./cmd/revues  (migrations appliquées au démarrage)"
  echo "Ou seed : ./scripts/reset-db.sh --seed"
fi
