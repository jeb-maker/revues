#!/usr/bin/env bash
# auto-deploy.sh — déploiement automatique de la prod depuis origin/main.
#
# Idempotent et sûr :
#   - ne déploie QUE si origin/main a du nouveau (sinon skip, pas de restart) ;
#   - FAST-FORWARD ONLY : si main a divergé du HEAD prod, on s'abstient ;
#   - flock : pas deux déploiements concurrents ;
#   - délègue à deploy/update.sh (git pull + docker compose build).
#
# Installation cron (VPS, toutes les 30 min) :
#   ( crontab -l 2>/dev/null | grep -v revues-auto-deploy; \
#     echo '*/30 * * * * /usr/local/sbin/revues-auto-deploy.sh' ) | crontab -

set -uo pipefail

APP_DIR="${REVUES_APP_DIR:-/opt/revues}"
LOG="${REVUES_DEPLOY_LOG:-/var/log/revues-deploy.log}"
LOCK="${REVUES_DEPLOY_LOCK:-/run/lock/revues-auto-deploy.lock}"

# Résoudre le script update : copie hors-arbre ou version dans le dépôt.
UPDATE_SCRIPT="${REVUES_UPDATE_SCRIPT:-}"
if [[ -z "$UPDATE_SCRIPT" ]]; then
  if [[ -x /usr/local/sbin/revues-update.sh ]]; then
    UPDATE_SCRIPT=/usr/local/sbin/revues-update.sh
  else
    UPDATE_SCRIPT="$APP_DIR/deploy/update.sh"
  fi
fi

cd "$APP_DIR" || exit 1
exec >>"$LOG" 2>&1

exec 9>"$LOCK"
if ! flock -n 9; then
  echo "[$(date -Iseconds)] deploy déjà en cours, skip."
  exit 0
fi

echo "=== [$(date -Iseconds)] auto-deploy : vérification ==="

if [[ ! -d "$APP_DIR/.git" ]]; then
  echo "  ERREUR : ${APP_DIR} n'est pas un dépôt git."
  exit 1
fi

if ! git fetch -q origin main; then
  echo "  fetch KO — abandon."
  exit 1
fi

LOCAL="$(git rev-parse HEAD)"
REMOTE="$(git rev-parse origin/main)"

if [[ "$LOCAL" = "$REMOTE" ]]; then
  echo "  déjà à jour ($LOCAL), rien à faire."
  exit 0
fi

if ! git merge-base --is-ancestor "$LOCAL" "$REMOTE"; then
  echo "  !! origin/main n'est PAS un descendant du HEAD prod (divergence)."
  echo "     Deploy automatique abandonné — intervention manuelle requise."
  exit 1
fi

echo "  nouveau : ${LOCAL:0:9} -> ${REMOTE:0:9} ($(git rev-list --count "$LOCAL".."$REMOTE") commit(s)). Déploiement..."

if bash "$UPDATE_SCRIPT" main; then
  echo "=== [$(date -Iseconds)] auto-deploy : OK ($(git rev-parse --short HEAD)) ==="
else
  rc=$?
  echo "=== [$(date -Iseconds)] auto-deploy : ÉCHEC (exit=$rc) ==="
  exit "$rc"
fi
