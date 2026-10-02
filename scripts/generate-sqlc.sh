#!/usr/bin/env bash
# Génère le package Go typé depuis internal/store/queries via sqlc.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v sqlc >/dev/null 2>&1; then
  echo "sqlc introuvable — installer : https://docs.sqlc.dev/en/latest/overview/install.html" >&2
  echo "Exemple : go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0" >&2
  exit 1
fi

sqlc generate -f sqlc.yaml
echo "OK — code généré dans internal/store/sqlc/"
