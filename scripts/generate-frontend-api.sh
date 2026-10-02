#!/usr/bin/env bash
# Génère le client TypeScript front depuis api/openapi/openapi.yaml.
# Voir docs/FRONTEND.md / docs/API.md — toolchain WP-005.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SPEC="api/openapi/openapi.yaml"
OUT="frontend/src/lib/api/schema.d.ts"

if [[ ! -f "$SPEC" ]]; then
  echo "missing OpenAPI spec: $SPEC" >&2
  exit 1
fi

if [[ ! -f frontend/package.json ]]; then
  echo "missing frontend/package.json" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUT")"

echo "generating ${OUT} from ${SPEC}..."
(
  cd frontend
  if [[ ! -d node_modules/openapi-typescript ]]; then
    echo "openapi-typescript missing — run npm ci in frontend/" >&2
    exit 1
  fi
  npx openapi-typescript "../${SPEC}" -o "src/lib/api/schema.d.ts"
)

echo "ok: ${OUT}"
echo "Wrappers non générés : frontend/src/lib/api/client.ts"
