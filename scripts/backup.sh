#!/usr/bin/env bash
# Sauvegarde SQLite + pièces jointes — Revues
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

info() { echo -e "${GREEN}==>${NC} $1"; }
fail() { echo -e "${RED}FAIL:${NC} $1" >&2; exit 1; }

DB_PATH="${REVUES_DATABASE_PATH:-data/revues.db}"
ATTACHMENTS_DIR="${REVUES_ATTACHMENTS_DIR:-data/attachments}"
BACKUP_ROOT="${REVUES_BACKUP_DIR:-backups}"
RETENTION_DAYS="${REVUES_BACKUP_RETENTION_DAYS:-90}"

if ! command -v sqlite3 >/dev/null 2>&1; then
  fail "sqlite3 absent — installer le client SQLite (ex. apt install sqlite3)"
fi

if [[ ! -f "$DB_PATH" ]]; then
  fail "Base introuvable : $DB_PATH (REVUES_DATABASE_PATH)"
fi

timestamp="$(date -u +%Y-%m-%dT%H%M%SZ)"
dest="${BACKUP_ROOT}/${timestamp}"
mkdir -p "$dest"

info "Dump SQL → ${dest}/revues.sql"
sqlite3 "$DB_PATH" ".dump" > "${dest}/revues.sql"

if [[ -d "$ATTACHMENTS_DIR" ]] && [[ -n "$(ls -A "$ATTACHMENTS_DIR" 2>/dev/null || true)" ]]; then
  info "Archive attachments → ${dest}/attachments.tar.gz"
  tar -czf "${dest}/attachments.tar.gz" -C "$(dirname "$ATTACHMENTS_DIR")" "$(basename "$ATTACHMENTS_DIR")"
else
  info "Répertoire attachments absent ou vide — attachments.tar.gz ignoré"
fi

{
  echo "created_utc=${timestamp}"
  echo "database_path=${DB_PATH}"
  echo "attachments_dir=${ATTACHMENTS_DIR}"
  echo "hostname=$(hostname 2>/dev/null || echo unknown)"
  echo "revues_sql_bytes=$(wc -c < "${dest}/revues.sql")"
  if [[ -f "${dest}/attachments.tar.gz" ]]; then
    echo "attachments_tar_bytes=$(wc -c < "${dest}/attachments.tar.gz")"
  fi
} > "${dest}/manifest.txt"

info "Purge des sauvegardes de plus de ${RETENTION_DAYS} jours dans ${BACKUP_ROOT}"
if [[ -d "$BACKUP_ROOT" ]]; then
  find "$BACKUP_ROOT" -mindepth 1 -maxdepth 1 -type d -mtime +"${RETENTION_DAYS}" -print -exec rm -rf {} +
fi

info "Sauvegarde terminée : ${dest}"
echo "${dest}"
