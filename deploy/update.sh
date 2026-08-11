#!/usr/bin/env bash
# Revues — mise à jour incrémentale (Docker)
# Usage (sur le VPS) : cd /opt/revues && bash deploy/update.sh [branche]
#
# git pull (fast-forward) → docker compose up -d --build → healthcheck.
set -euo pipefail

# Hors arbre (ex. /usr/local/sbin) : fixer REVUES_APP_DIR=/opt/revues
if [[ -n "${REVUES_APP_DIR:-}" ]]; then
  APP_DIR="$REVUES_APP_DIR"
elif [[ "$(basename "$(dirname "$0")")" == "deploy" ]]; then
  APP_DIR="$(cd "$(dirname "$0")/.." && pwd)"
else
  APP_DIR="${REVUES_APP_DIR:-/opt/revues}"
fi
BRANCH="${1:-main}"
HEALTH_URL="${REVUES_HEALTH_URL:-http://127.0.0.1:8088/healthz}"

cd "$APP_DIR"

echo "=== Revues — mise à jour (Docker) ==="
echo "    Répertoire : ${APP_DIR}"
echo "    Branche    : ${BRANCH}"
echo ""

if ! command -v docker >/dev/null 2>&1; then
  echo "ERREUR : docker introuvable." >&2
  exit 1
fi
if [[ ! -f "$APP_DIR/.env" ]]; then
  echo "ERREUR : ${APP_DIR}/.env manquant." >&2
  exit 1
fi

echo "[1/3] Git pull (${BRANCH})..."
git fetch origin
git checkout "$BRANCH"
git pull --ff-only origin "$BRANCH"
echo "    Commit courant : $(git log --oneline -1)"

echo ""
echo "[2/3] docker compose up -d --build..."
docker compose --project-directory "$APP_DIR" up -d --build --remove-orphans

echo ""
echo "[3/3] Healthcheck ${HEALTH_URL}..."
ok=0
for _ in $(seq 1 30); do
  if curl -sf "$HEALTH_URL" >/dev/null; then
    ok=1
    break
  fi
  sleep 2
done
if [[ "$ok" -ne 1 ]]; then
  echo "ERREUR : healthcheck KO après 60s." >&2
  docker compose --project-directory "$APP_DIR" ps
  docker compose --project-directory "$APP_DIR" logs --tail=80 app
  exit 1
fi

echo "=== Revues — mise à jour OK ($(git rev-parse --short HEAD)) ==="
