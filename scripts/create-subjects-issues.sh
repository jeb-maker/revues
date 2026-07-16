#!/usr/bin/env bash
# Crée l'épique et les 6 issues « Sujets greenfield » sur GitHub.
set -euo pipefail

REPO="${GITHUB_REPOSITORY:-jeb-maker/revues}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC="$ROOT/docs/issues/subjects-epic.md"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI requis" >&2
  exit 1
fi

ensure_label() {
  gh label create "$1" --repo "$REPO" --color "$2" --description "$3" 2>/dev/null || true
}

ensure_label "vague-subjects" "0E8A16" "Sujets greenfield"
ensure_label "area:integrations" "5319E7" "Integrations"

extract_issue() {
  local n="$1"
  sed -n "/^## Issue ${n}[ —]/,/^---$/p" "$SPEC" | head -n -1 | tail -n +2
}

EPIC_BODY='## Contexte

Remplacer projets par sujets. Spec: docs/issues/subjects-epic.md

## Issues filles

_(Compléter après création)_'

if gh issue list --repo "$REPO" --search "[Epic] Sujets in:title" --state all --json number -q '.[0].number' 2>/dev/null | grep -qE '^[0-9]+$'; then
  EPIC_NUM=$(gh issue list --repo "$REPO" --search "[Epic] Sujets in:title" --state all --json number -q '.[0].number')
else
  EPIC_URL=$(gh issue create --repo "$REPO" --title "[Epic] Sujets — remplacement projets (greenfield)" --label "epic" --label "vague-subjects" --label "area:data" --label "area:core" --body "$EPIC_BODY")
  EPIC_NUM=$(echo "$EPIC_URL" | grep -oE '[0-9]+$')
fi
echo "Épique #$EPIC_NUM"

TITLES=("[data] Schéma sujets greenfield" "[store] Store sujets et domaines" "[auth] RBAC sujets v1 + tests IDOR" "[features] Handlers sujets + wizard revue" "[ui][integrations] Templates, intégrations, docs" "[ui] Preset libellé sujet (admin org)")
LABELS=("area:data vague-subjects" "area:data area:core vague-subjects" "area:auth area:core vague-subjects" "area:core area:ui vague-subjects" "area:ui area:integrations area:core vague-subjects" "area:ui area:admin vague-subjects")
BLOCK=("" "Issue 1" "Issue 2" "Issue 3" "Issue 4" "Issue 5")
TASK_LIST=""
for i in "${!TITLES[@]}"; do
  n=$((i+1))
  body="$(extract_issue "$n")

---
Épique parente : #${EPIC_NUM}"
  [[ -n "${BLOCK[$i]}" ]] && body="${body}
Bloqué par : ${BLOCK[$i]}"
  args=()
  for l in ${LABELS[$i]}; do args+=(--label "$l"); done
  if gh issue list --repo "$REPO" --search "${TITLES[$i]} in:title" --state all --json number -q '.[0].number' 2>/dev/null | grep -qE '^[0-9]+$'; then
    num=$(gh issue list --repo "$REPO" --search "${TITLES[$i]} in:title" --state all --json number -q '.[0].number')
  else
    url=$(gh issue create --repo "$REPO" --title "${TITLES[$i]}" "${args[@]}" --body "$body")
    num=$(echo "$url" | grep -oE '[0-9]+$')
  fi
  echo "Issue #$num : ${TITLES[$i]}"
  TASK_LIST="${TASK_LIST}- [ ] #${num}
"
done

gh issue edit "$EPIC_NUM" --repo "$REPO" --body "## Contexte

Remplacer projets par sujets. Spec: docs/issues/subjects-epic.md

## Issues filles

${TASK_LIST}"
echo "https://github.com/$REPO/issues/$EPIC_NUM"
