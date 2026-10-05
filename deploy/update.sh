#!/usr/bin/env bash
# Revues — mise à jour incrémentale (Docker + image GHCR)
# Usage (sur le VPS) : cd /opt/revues && bash deploy/update.sh [branche]
#
# git pull (fast-forward) → pull image sha-<commit> (GHCR) → up -d → healthcheck.
# Fallback : REVUES_DEPLOY_BUILD=1 force un build local (sans attendre Actions).
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
# Aligné sur docker/metadata-action format=short (7).
SHA_LEN="${REVUES_IMAGE_SHA_LEN:-7}"
PULL_ATTEMPTS="${REVUES_IMAGE_PULL_ATTEMPTS:-30}"
PULL_SLEEP_SEC="${REVUES_IMAGE_PULL_SLEEP_SEC:-20}"

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

SHORT_SHA="$(git rev-parse --short="${SHA_LEN}" HEAD)"
export REVUES_IMAGE="${REVUES_IMAGE:-ghcr.io/jeb-maker/revues}"
export REVUES_IMAGE_TAG="${REVUES_IMAGE_TAG:-sha-${SHORT_SHA}}"

echo ""
if [[ "${REVUES_DEPLOY_BUILD:-}" == "1" ]]; then
  echo "[2/3] docker compose up -d --build (REVUES_DEPLOY_BUILD=1)..."
  docker compose --project-directory "$APP_DIR" up -d --build --remove-orphans
else
  echo "[2/3] Pull ${REVUES_IMAGE}:${REVUES_IMAGE_TAG} puis up -d..."
  pulled=0
  for i in $(seq 1 "$PULL_ATTEMPTS"); do
    if docker compose --project-directory "$APP_DIR" pull app; then
      pulled=1
      break
    fi
    echo "    image pas encore sur GHCR (tentative ${i}/${PULL_ATTEMPTS}), nouvel essai dans ${PULL_SLEEP_SEC}s..."
    sleep "$PULL_SLEEP_SEC"
  done
  if [[ "$pulled" -ne 1 ]]; then
    echo "ERREUR : impossible de pull ${REVUES_IMAGE}:${REVUES_IMAGE_TAG} après $((PULL_ATTEMPTS * PULL_SLEEP_SEC))s." >&2
    echo "         Vérifier Actions « Docker » ou relancer avec REVUES_DEPLOY_BUILD=1." >&2
    exit 1
  fi
  docker compose --project-directory "$APP_DIR" up -d --remove-orphans --no-build
fi

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

# La SPA doit être servie : sans build embarqué, le binaire répond 503 « SPA non construite ».
SPA_URL="${REVUES_SPA_CHECK_URL:-${HEALTH_URL%/healthz}/login}"
if ! curl -sf "$SPA_URL" | grep -q '_app/'; then
  echo "ERREUR : ${SPA_URL} ne sert pas la SPA SvelteKit (stub ou 503) — image sans frontend/build ?" >&2
  docker compose --project-directory "$APP_DIR" logs --tail=40 app
  exit 1
fi

echo "=== Revues — mise à jour OK ($(git rev-parse --short HEAD) → ${REVUES_IMAGE}:${REVUES_IMAGE_TAG}) ==="
