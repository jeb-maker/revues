#!/usr/bin/env bash
# Génère / met à jour .env production sur l'hôte (ne pas committer).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/.env"
BASE_URL="${REVUES_BASE_URL:-https://revues.betafly.ovh}"

umask 077

if [[ -f "${ENV_FILE}" ]]; then
	echo "Conservé existant: ${ENV_FILE}"
	exit 0
fi

SESSION_SECRET="$(openssl rand -base64 48)"
ENCRYPTION_KEY="$(openssl rand -base64 32)"

cat >"${ENV_FILE}" <<EOF
# Généré sur l'hôte $(date -u +%Y-%m-%dT%H:%M:%SZ) — ne pas committer
REVUES_HOST_PORT=8088
REVUES_BASE_URL=${BASE_URL}
REVUES_SESSION_SECRET=${SESSION_SECRET}
REVUES_ENCRYPTION_KEY=${ENCRYPTION_KEY}
REVUES_GITHUB_CLIENT_ID=
REVUES_GITHUB_CLIENT_SECRET=
REVUES_BOOTSTRAP_ADMIN_EMAIL=
EOF

chmod 600 "${ENV_FILE}"
echo "Créé ${ENV_FILE} (secrets générés). Remplir OAuth GitHub + bootstrap admin."
