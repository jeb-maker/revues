#!/usr/bin/env bash
# Lance un scénario k6 contre l'API Revues locale.
#
# Usage:
#   ./scripts/load/run.sh read
#   ./scripts/load/run.sh mix
#   ./scripts/load/run.sh runs
#   VUS=50 DURATION=1m BASE_URL=http://127.0.0.1:8080 ./scripts/load/run.sh mix
#
# Prérequis : API démarrée (go run ./cmd/revues), REVUES_DEV_AUTH=0 recommandé.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCENARIO="${1:-read}"
SCRIPT="$ROOT/scripts/load/${SCENARIO}.js"

if [[ ! -f "$SCRIPT" ]]; then
  echo "Scénario inconnu: ${SCENARIO} (attendu: read | mix | runs)" >&2
  exit 1
fi

resolve_k6() {
  if [[ -n "${K6:-}" && -x "$K6" ]]; then
    echo "$K6"
    return
  fi
  if command -v k6 >/dev/null 2>&1; then
    command -v k6
    return
  fi
  if [[ -x "$ROOT/.tools/k6" ]]; then
    echo "$ROOT/.tools/k6"
    return
  fi
  echo "k6 introuvable. Installe-le (https://k6.io/docs/get-started/installation/) ou :" >&2
  echo "  mkdir -p $ROOT/.tools && curl -fsSL \\" >&2
  echo "    https://github.com/grafana/k6/releases/download/v0.57.0/k6-v0.57.0-linux-amd64.tar.gz \\" >&2
  echo "    | tar -xz -C $ROOT/.tools --strip-components=1 k6-v0.57.0-linux-amd64/k6" >&2
  exit 1
}

K6_BIN="$(resolve_k6)"
export BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
export VUS="${VUS:-20}"
export DURATION="${DURATION:-30s}"
export WRITE_RATIO="${WRITE_RATIO:-0.2}"
export RUN_RATIO="${RUN_RATIO:-0.1}"
export LOAD_PASSWORD="${LOAD_PASSWORD:-loadtest-pass-12}"
export SESSION_POOL="${SESSION_POOL:-10}"
if [[ -n "${LOAD_EMAIL:-}" ]]; then
  export LOAD_EMAIL
fi
if [[ -n "${LOAD_EMAIL_PREFIX:-}" ]]; then
  export LOAD_EMAIL_PREFIX
fi

echo "→ k6=$K6_BIN  scenario=$SCENARIO  BASE_URL=$BASE_URL  VUS=$VUS  DURATION=$DURATION  SESSION_POOL=$SESSION_POOL"
exec "$K6_BIN" run "$SCRIPT"
