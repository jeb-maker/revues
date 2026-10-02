#!/usr/bin/env bash
# Génère le code serveur Go depuis api/openapi/openapi.yaml (oapi-codegen).
# Voir docs/API.md — toolchain WP-002.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export PATH="$(go env GOPATH)/bin:${PATH}"

OAPI_CODEGEN_VERSION="${OAPI_CODEGEN_VERSION:-v2.4.1}"
SPEC="api/openapi/openapi.yaml"
CFG="api/openapi/oapi-codegen.yaml"
OUT="internal/api/v1/oapi.gen.go"

if [[ ! -f "$SPEC" ]]; then
  echo "missing OpenAPI spec: $SPEC" >&2
  exit 1
fi
if [[ ! -f "$CFG" ]]; then
  echo "missing oapi-codegen config: $CFG" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUT")"

if ! command -v oapi-codegen >/dev/null 2>&1; then
  echo "installing oapi-codegen ${OAPI_CODEGEN_VERSION}..."
  go install "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}"
fi

echo "generating ${OUT} from ${SPEC}..."
oapi-codegen -config "$CFG" "$SPEC"

# Ensure package directory is gofmt-clean.
gofmt -w "$OUT"

echo "ok: ${OUT}"
