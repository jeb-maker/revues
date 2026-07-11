#!/usr/bin/env bash
# Crée l'épique et les 7 issues « Organisations » sur GitHub.
# Prérequis : gh auth login
set -euo pipefail

REPO="${GITHUB_REPOSITORY:-jeb-maker/revues}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC="$ROOT/docs/issues/organizations-epic.md"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI requis : https://cli.github.com/" >&2
  exit 1
fi

gh label create "vague-4" --repo "$REPO" --color "0E8A16" --description "Organisations multi-tenant" 2>/dev/null || true

extract_issue() {
  local n="$1"
  sed -n "/^## Issue ${n} /,/^---$/p" "$SPEC" | head -n -1 | tail -n +2
}

EPIC_BODY="$(cat <<EOF
## Contexte

Revues v1 est mono-instance. Introduire une couche **Organisation** au-dessus des **Projets** pour le self-service B2B.

Spec : [docs/issues/organizations-epic.md](https://github.com/jeb-maker/revues/blob/main/docs/issues/organizations-epic.md)

## Issues filles

_(Compléter après création)_

## Hors scope

- PostgreSQL / multi-tenant infra
- Vérification domaine email automatique
EOF
)"

echo "Création épique..."
EPIC_URL=$(gh issue create --repo "$REPO" \
  --title "[Epic] Organisations multi-tenant self-service" \
  --label "epic" --label "vague-4" --label "area:data" --label "area:auth" \
  --body "$EPIC_BODY")
EPIC_NUM=$(echo "$EPIC_URL" | grep -oE '[0-9]+$')
echo "Épique #$EPIC_NUM"

declare -a ISSUE_URLS

create_issue() {
  local title="$1"
  shift
  local -a labels=("$@")
  local label_args=()
  for l in "${labels[@]}"; do
    label_args+=(--label "$l")
  done
  gh issue create --repo "$REPO" --title "$title" "${label_args[@]}" --body "$body"
}

declare -a TITLES=(
  "[data] Schéma organizations + organization_members + store"
  "[auth] Organisation active en session + middleware"
  "[auth][ui] Création org self-service + sélecteur multi-org"
  "[data][core] Scoper projets et entités métier par organization_id"
  "[core] Invitation projet → adhésion org induite"
  "[admin][auth] Whitelist globale → membres / invitations org"
  "[ui] Switcher organisation + invitations en attente"
)

declare -a LABEL_SETS=(
  "area:data vague-4"
  "area:auth vague-4"
  "area:auth area:ui vague-4"
  "area:data area:core vague-4"
  "area:core vague-4"
  "area:admin area:auth vague-4"
  "area:ui vague-4"
)

for i in "${!TITLES[@]}"; do
  n=$((i + 1))
  body="$(extract_issue "$n")

  body="$body

---
Épique parente : #$EPIC_NUM"

  if [[ "$n" -gt 1 ]]; then
    body="$body
Bloqué par : issue précédente (voir spec)."
  fi

  label_args=()
  for l in ${LABEL_SETS[$i]}; do
    label_args+=(--label "$l")
  done

  url=$(gh issue create --repo "$REPO" --title "${TITLES[$i]}" "${label_args[@]}" --body "$body")
  ISSUE_URLS+=("$url")
  num=$(echo "$url" | grep -oE '[0-9]+$')
  echo "Issue #$num : ${TITLES[$i]}"
done

TASK_LIST=""
for url in "${ISSUE_URLS[@]}"; do
  num=$(echo "$url" | grep -oE '[0-9]+$')
  TASK_LIST="${TASK_LIST}- [ ] #${num}"$'\n'
done

gh issue edit "$EPIC_NUM" --repo "$REPO" --body "$(cat <<EOF
## Contexte

Revues v1 est mono-instance. Introduire une couche **Organisation** au-dessus des **Projets** pour le self-service B2B.

Spec : [docs/issues/organizations-epic.md](https://github.com/jeb-maker/revues/blob/main/docs/issues/organizations-epic.md)

## Issues filles

${TASK_LIST}

## Hors scope

- PostgreSQL / multi-tenant infra
- Vérification domaine email automatique
EOF
)"

echo ""
echo "Terminé. Épique : https://github.com/$REPO/issues/$EPIC_NUM"
echo "Déléguer Issue 1 avec le prompt dans docs/issues/organizations-epic.md"
