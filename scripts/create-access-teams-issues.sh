#!/usr/bin/env bash
# Crée l'épique et les 9 issues « Accès par équipes » sur GitHub.
# Prérequis : gh auth login
set -euo pipefail

REPO="${GITHUB_REPOSITORY:-jeb-maker/revues}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC="$ROOT/docs/issues/access-teams-epic.md"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI requis : https://cli.github.com/" >&2
  exit 1
fi

if [[ ! -f "$SPEC" ]]; then
  echo "Fichier spec introuvable : $SPEC" >&2
  echo "git pull origin main (docs/issues/access-teams-epic.md requis)" >&2
  exit 1
fi

ensure_label() {
  local name="$1" color="$2" desc="$3"
  gh label create "$name" --repo "$REPO" --color "$color" --description "$desc" 2>/dev/null || true
}

echo "==> Labels (création si absents)..."
ensure_label "epic" "5319E7" "Issue épique regroupant plusieurs tâches"
ensure_label "vague-5" "1D76DB" "Accès par équipes et gouvernance org"
ensure_label "area:data" "0075CA" "Schéma, migrations, store"
ensure_label "area:auth" "D93F0B" "OAuth, sessions, RBAC"
ensure_label "area:ui" "FBCA04" "Templates, HTMX, écrans"
ensure_label "area:core" "1D76DB" "Logique métier cœur"
ensure_label "area:admin" "B60205" "Administration"

extract_issue() {
  local n="$1"
  sed -n "/^## Issue ${n}[ —]/,/^---$/p" "$SPEC" | head -n -1 | tail -n +2
}

EPIC_BODY="$(cat <<EOF
## Contexte

Introduire les **équipes** comme chemin principal d'accès collectif aux projets, conformément aux pratiques B2B (groupes + rôles sur ressources). Les tags projet restent **métier** (modèles).

Spec : [docs/issues/access-teams-epic.md](https://github.com/jeb-maker/revues/blob/main/docs/issues/access-teams-epic.md)

## Issues filles

_(Compléter après création)_

## Hors scope

- Tags projet = accès
- SCIM / LDAP
- Expiration automatique des accès
EOF
)"

echo "Création épique..."
if gh issue list --repo "$REPO" --search "[Epic] Accès par équipes in:title" --state all --json number -q '.[0].number' 2>/dev/null | grep -qE '^[0-9]+$'; then
  EPIC_NUM=$(gh issue list --repo "$REPO" --search "[Epic] Accès par équipes in:title" --state all --json number -q '.[0].number')
  echo "Épique déjà existante : #$EPIC_NUM — skip création"
else
  EPIC_URL=$(gh issue create --repo "$REPO" \
    --title "[Epic] Accès par équipes et gouvernance org" \
    --label "epic" --label "vague-5" --label "area:auth" --label "area:data" \
    --body "$EPIC_BODY")
  EPIC_NUM=$(echo "$EPIC_URL" | grep -oE '[0-9]+$')
  echo "Épique #$EPIC_NUM"
fi

declare -a ISSUE_URLS

declare -a TITLES=(
  "[auth] Spec RBAC — équipes, org admin, projets privés"
  "[data] Migration organization_teams + store"
  "[store][auth] ResolveProjectAccess + tests"
  "[auth] Refactor handlers — ResolveProjectAccess"
  "[auth] Org admin — visibilité globale org + TestIDOR"
  "[ui] Admin équipes — CRUD"
  "[ui] Projet — équipes, preview, sources d'accès"
  "[data][auth] Projets privés (visibility)"
  "[admin] Politiques org — délégation lead"
)

declare -a LABEL_SETS=(
  "area:auth vague-5"
  "area:data vague-5"
  "area:auth area:data vague-5"
  "area:auth area:core vague-5"
  "area:auth vague-5"
  "area:ui area:admin vague-5"
  "area:ui area:core vague-5"
  "area:data area:auth vague-5"
  "area:admin area:auth vague-5"
)

declare -a BLOCKED_BY=(
  ""
  "Issue 1 (RBAC spec)"
  "Issue 2"
  "Issue 3"
  "Issue 3"
  "Issue 2"
  "Issues 4 et 6"
  "Issue 3"
  "Issue 7"
)

for i in "${!TITLES[@]}"; do
  n=$((i + 1))
  issue_body="$(extract_issue "$n")"
  if [[ -z "$issue_body" ]]; then
    echo "Erreur: contenu Issue ${n} introuvable dans ${SPEC}" >&2
    exit 1
  fi

  body="${issue_body}

---
Épique parente : #${EPIC_NUM}"

  if [[ -n "${BLOCKED_BY[$i]}" ]]; then
    body="${body}
Bloqué par : ${BLOCKED_BY[$i]} (voir spec)."
  fi

  label_args=()
  for l in ${LABEL_SETS[$i]}; do
    label_args+=(--label "$l")
  done

  if gh issue list --repo "$REPO" --search "${TITLES[$i]} in:title" --state all --json number -q '.[0].number' 2>/dev/null | grep -qE '^[0-9]+$'; then
    num=$(gh issue list --repo "$REPO" --search "${TITLES[$i]} in:title" --state all --json number -q '.[0].number')
    url="https://github.com/$REPO/issues/$num"
    echo "Issue #$num déjà existante — skip : ${TITLES[$i]}"
  else
    url=$(gh issue create --repo "$REPO" --title "${TITLES[$i]}" "${label_args[@]}" --body "$body")
    num=$(echo "$url" | grep -oE '[0-9]+$')
    echo "Issue #$num : ${TITLES[$i]}"
  fi
  ISSUE_URLS+=("$url")
done

TASK_LIST=""
for url in "${ISSUE_URLS[@]}"; do
  num=$(echo "$url" | grep -oE '[0-9]+$')
  TASK_LIST="${TASK_LIST}- [ ] #${num}"$'\n'
done

gh issue edit "$EPIC_NUM" --repo "$REPO" --body "$(cat <<EOF
## Contexte

Introduire les **équipes** comme chemin principal d'accès collectif aux projets. Les tags projet restent **métier** (modèles).

Spec : [docs/issues/access-teams-epic.md](https://github.com/jeb-maker/revues/blob/main/docs/issues/access-teams-epic.md)

## Issues filles

${TASK_LIST}

## Hors scope

- Tags projet = accès
- SCIM / LDAP
- Expiration automatique des accès
EOF
)"

echo ""
echo "Terminé. Épique : https://github.com/$REPO/issues/$EPIC_NUM"
echo "Déléguer Issue 1 avec le prompt dans docs/issues/access-teams-epic.md"
