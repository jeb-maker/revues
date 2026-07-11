#!/usr/bin/env bash
# Crée les issues GitHub + PRs empilées de l'épique Organisations.
# Prérequis : gh auth login
set -euo pipefail

export PATH="${HOME}/bin:${PATH}"
REPO="${GITHUB_REPOSITORY:-jeb-maker/revues}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! gh auth status >/dev/null 2>&1; then
  echo "Authentification requise : gh auth login" >&2
  exit 1
fi

issue_num() {
  local title="$1"
  gh issue list --repo "$REPO" --search "$title in:title" --state open --json number -q '.[0].number'
}

create_pr() {
  local base="$1" head="$2" title="$3" issue_n="$4"
  if gh pr list --repo "$REPO" --head "$head" --json number -q '.[0].number' 2>/dev/null | grep -qE '^[0-9]+$'; then
    echo "PR déjà existante pour $head — skip"
    return 0
  fi
  gh pr create --repo "$REPO" --base "$base" --head "$head" --title "$title" --body "Closes #${issue_n}"
}

echo "==> Création des issues (skip si déjà créées)..."
if ! gh issue list --repo "$REPO" --search "[Epic] Organisations multi-tenant" --json number -q '.[0].number' | grep -qE '^[0-9]+$'; then
  "$ROOT/scripts/create-organization-issues.sh" | tee /tmp/org-issues-created.txt
else
  echo "Épique déjà présente — skip create-organization-issues.sh"
fi

echo "==> Création des PRs empilées..."
create_pr main cursor/issue-organizations-schema-f21b \
  "[data] Schéma organizations + organization_members + store" \
  "$(issue_num '[data] Schéma organizations')"

create_pr cursor/issue-organizations-schema-f21b cursor/issue-organizations-session-f21b \
  "[auth] Organisation active en session + middleware" \
  "$(issue_num '[auth] Organisation active en session')"

create_pr cursor/issue-organizations-session-f21b cursor/issue-organizations-onboarding-f21b \
  "[auth][ui] Création org self-service + sélecteur multi-org" \
  "$(issue_num '[auth][ui] Création org self-service')"

create_pr cursor/issue-organizations-onboarding-f21b cursor/issue-organizations-scoping-f21b \
  "[data][core] Scoper projets et entités métier par organization_id" \
  "$(issue_num '[data][core] Scoper projets')"

create_pr cursor/issue-organizations-scoping-f21b cursor/issue-organizations-project-invite-f21b \
  "[core] Invitation projet → adhésion org induite" \
  "$(issue_num '[core] Invitation projet')"

create_pr cursor/issue-organizations-project-invite-f21b cursor/issue-organizations-whitelist-f21b \
  "[admin][auth] Whitelist globale → membres / invitations org" \
  "$(issue_num '[admin][auth] Whitelist globale')"

create_pr cursor/issue-organizations-whitelist-f21b cursor/issue-organizations-switcher-f21b \
  "[ui] Switcher organisation + invitations en attente" \
  "$(issue_num '[ui] Switcher organisation')"

echo ""
echo "Terminé. Merger dans l'ordre 1→7."
echo "Follow-up recommandé : issue dédiée pour mettre à jour docs/RBAC.md"
