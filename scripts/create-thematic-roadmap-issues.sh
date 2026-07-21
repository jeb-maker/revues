#!/usr/bin/env bash
# Crée les épiques A–G et les issues de la roadmap thématique (lots 0–5).
# Spec : docs/issues/thematic-roadmap-epic.md
# Prérequis : gh auth login
set -euo pipefail

REPO="${GITHUB_REPOSITORY:-jeb-maker/revues}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC="$ROOT/docs/issues/thematic-roadmap-epic.md"
LOT=""
DRY_RUN=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --lot) LOT="${2:-}"; shift 2 ;;
    --lot=*) LOT="${1#--lot=}"; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help)
      echo "Usage: $0 [--lot N|all] [--dry-run]"
      echo "  --lot 0  épiques A–G"
      echo "  --lot 1  gates sécu (A2,A4,A5,G2)"
      echo "  --lot 2  SimpleUI + empty states"
      echo "  --lot 3  preuve C0–C3"
      echo "  --lot 4  checklist #66"
      echo "  --lot 5  post-#66 (D1,D6,E*,B4*)"
      echo "  (sans --lot) tous les lots"
      exit 0
      ;;
    *) echo "Option inconnue: $1" >&2; exit 1 ;;
  esac
done

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI requis : https://cli.github.com/" >&2
  exit 1
fi

if [[ "$DRY_RUN" -eq 0 ]] && ! gh auth status --hostname github.com >/dev/null 2>&1; then
  echo "gh non authentifié. Options :" >&2
  echo "  gh auth login" >&2
  echo "  export GH_TOKEN=ghp_...   # puis relancer ce script" >&2
  exit 1
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "==> DRY-RUN (aucune écriture GitHub)"
fi

if [[ ! -f "$SPEC" ]]; then
  echo "Fichier spec introuvable : $SPEC" >&2
  exit 1
fi

ensure_label() {
  local name="$1" color="$2" desc="$3"
  gh label create "$name" --repo "$REPO" --color "$color" --description "$desc" 2>/dev/null || true
}

echo "==> Labels..."
ensure_label "epic" "5319E7" "Issue épique regroupant plusieurs tâches"
ensure_label "vague-thematic" "0E8A16" "Roadmap thématique post-cœur"
ensure_label "area:data" "0075CA" "Schéma, migrations, store"
ensure_label "area:auth" "D93F0B" "OAuth, sessions, RBAC"
ensure_label "area:ui" "FBCA04" "Templates, HTMX, écrans"
ensure_label "area:core" "1D76DB" "Logique métier cœur"
ensure_label "area:admin" "B60205" "Administration"
ensure_label "area:integrations" "5319E7" "Integrations"
ensure_label "area:notifications" "C5DEF5" "Email SMTP notifications"
ensure_label "area:attachments" "D4C5F9" "Pièces jointes"
ensure_label "area:infra" "BFDADC" "Infra, rate limit, headers"

extract_section() {
  # Extract from "## Issue KEY —" or "## Epic KEY —" until next ## at same level or ---
  local heading="$1"
  awk -v h="$heading" '
    $0 ~ "^## " h " —" {p=1; next}
    p && /^## / {exit}
    p && /^---$/ {exit}
    p {print}
  ' "$SPEC"
}

find_issue_by_title() {
  local title="$1"
  gh issue list --repo "$REPO" --search "${title} in:title" --state all --json number,title \
    -q ".[] | select(.title == \"${title}\") | .number" 2>/dev/null | head -1
}

create_or_get_issue() {
  local title="$1"
  local body_file="$2"
  shift 2
  local labels=("$@")
  local existing
  if [[ "$DRY_RUN" -eq 0 ]]; then
    existing="$(find_issue_by_title "$title" || true)"
  fi
  if [[ -n "${existing:-}" ]]; then
    echo "$existing"
    return
  fi
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "DRY:$title" >&2
    echo "dry"
    return
  fi
  local args=(--repo "$REPO" --title "$title" --body-file "$body_file")
  local l
  for l in "${labels[@]}"; do
    [[ -n "$l" ]] && args+=(--label "$l")
  done
  local url
  url="$(gh issue create "${args[@]}")"
  echo "$url" | grep -oE '[0-9]+$'
}

link_existing() {
  # Prefer existing numbered issue if open with matching keyword search
  local search="$1"
  local want_title="$2"
  local num
  num="$(gh issue list --repo "$REPO" --search "$search" --state all --json number,title \
    -q ".[0].number" 2>/dev/null || true)"
  if [[ "$num" =~ ^[0-9]+$ ]]; then
    echo "$num"
  else
    echo ""
  fi
}

should_run_lot() {
  local n="$1"
  [[ -z "$LOT" || "$LOT" == "all" || "$LOT" == "$n" ]]
}

declare -A EPIC_NUMS=()
declare -A ISSUE_NUMS=()

create_epics() {
  echo "==> Lot 0 — Épiques A–G"
  local letter title labels body num
  declare -a LETTERS=(A B C D E F G)
  declare -a TITLES=(
    "[Epic] Adoption — parcours pilote"
    "[Epic] Progressive disclosure (SimpleUI)"
    "[Epic] Conformité & preuve"
    "[Epic] Opérationnel qualité"
    "[Epic] Intégrations v2+ (légères)"
    "[Epic] Gouvernance suite (icebox)"
    "[Epic] Hardening & dette"
  )
  declare -a LABEL_SETS=(
    "epic vague-thematic area:ui area:auth"
    "epic vague-thematic area:ui"
    "epic vague-thematic area:core"
    "epic vague-thematic area:core"
    "epic vague-thematic area:integrations"
    "epic vague-thematic area:auth area:admin"
    "epic vague-thematic area:infra area:auth"
  )
  local i
  for i in "${!LETTERS[@]}"; do
    letter="${LETTERS[$i]}"
    title="${TITLES[$i]}"
    body="$(extract_section "Epic ${letter}")"
    if [[ -z "$body" ]]; then
      body="Spec : docs/issues/thematic-roadmap-epic.md — Epic ${letter}"
    fi
    body="${body}

---
Spec complète : docs/issues/thematic-roadmap-epic.md"
    num=""
    if [[ "$DRY_RUN" -eq 0 ]]; then
      num="$(find_issue_by_title "$title" || true)"
    fi
    if [[ -n "$num" ]]; then
      echo "Épique ${letter} déjà existante : #$num"
    else
      # shellcheck disable=SC2206
      local labs=(${LABEL_SETS[$i]})
      local tmp
      tmp="$(mktemp)"
      printf '%s\n' "$body" >"$tmp"
      num="$(create_or_get_issue "$title" "$tmp" "${labs[@]}")"
      rm -f "$tmp"
      echo "Épique ${letter} → ${num} ($title)"
    fi
    EPIC_NUMS[$letter]="$num"
  done
}

create_issue_from_key() {
  local key="$1"
  local title="$2"
  local epic_letter="$3"
  shift 3
  local labels=("$@")
  local body num epic
  body="$(extract_section "Issue ${key}")"
  if [[ -z "$body" ]]; then
    echo "WARN: section Issue ${key} introuvable dans spec" >&2
    body="Voir docs/issues/thematic-roadmap-epic.md — Issue ${key}"
  fi
  epic="${EPIC_NUMS[$epic_letter]:-}"
  body="${body}

---
Épique parente : #${epic}
Spec : docs/issues/thematic-roadmap-epic.md"
  num=""
  if [[ "$DRY_RUN" -eq 0 ]]; then
    num="$(find_issue_by_title "$title" || true)"
  fi
  if [[ -n "$num" ]]; then
    echo "  skip $key → #$num (existe)"
  else
    local tmp
    tmp="$(mktemp)"
    printf '%s\n' "$body" >"$tmp"
    num="$(create_or_get_issue "$title" "$tmp" "${labels[@]}")"
    rm -f "$tmp"
    echo "  créé $key → ${num} — $title"
  fi
  ISSUE_NUMS[$key]="$num"
}

update_issue_66() {
  echo "==> Lot 4 — Checklist pilote sur #66"
  local body
  body="$(extract_section "Issue A3")"
  if [[ -z "$body" ]]; then
    echo "WARN: Issue A3 introuvable" >&2
    return
  fi
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "  dry-run: commenter #66 avec checklist A3"
    ISSUE_NUMS[A3]="66"
    return
  fi
  local existing
  existing="$(gh issue view 66 --repo "$REPO" --json number -q .number 2>/dev/null || true)"
  if [[ "$existing" != "66" ]]; then
    echo "Issue #66 introuvable — création A3 standalone"
    create_issue_from_key "A3" "[meta] Pilote vague 1a — checklist PASS/FAIL" "A" "vague-thematic"
    return
  fi
  gh issue comment 66 --repo "$REPO" --body "$(cat <<EOF
## Checklist terrain (roadmap thématique)

Mise à jour via roadmap thématique — spec \`docs/issues/thematic-roadmap-epic.md\` (Issue A3).

${body}

Épique Adoption : #${EPIC_NUMS[A]:-?}
EOF
)"
  echo "  commentaire checklist ajouté sur #66"
  ISSUE_NUMS[A3]="66"
}

# --- Lots ---

if should_run_lot 0 || [[ -z "$LOT" || "$LOT" == "all" ]]; then
  create_epics
fi

# Always resolve epic numbers if later lots run alone
if [[ ${#EPIC_NUMS[@]} -eq 0 ]]; then
  EPIC_NUMS[A]="$(find_issue_by_title "[Epic] Adoption — parcours pilote" || true)"
  EPIC_NUMS[B]="$(find_issue_by_title "[Epic] Progressive disclosure (SimpleUI)" || true)"
  EPIC_NUMS[C]="$(find_issue_by_title "[Epic] Conformité & preuve" || true)"
  EPIC_NUMS[D]="$(find_issue_by_title "[Epic] Opérationnel qualité" || true)"
  EPIC_NUMS[E]="$(find_issue_by_title "[Epic] Intégrations v2+ (légères)" || true)"
  EPIC_NUMS[F]="$(find_issue_by_title "[Epic] Gouvernance suite (icebox)" || true)"
  EPIC_NUMS[G]="$(find_issue_by_title "[Epic] Hardening & dette" || true)"
fi

if should_run_lot 1 || [[ -z "$LOT" || "$LOT" == "all" ]]; then
  echo "==> Lot 1 — Gates sécu"
  local62=""
  if [[ "$DRY_RUN" -eq 0 ]]; then
    local62="$(gh issue view 62 --repo "$REPO" --json number -q .number 2>/dev/null || true)"
  fi
  if [[ "$local62" == "62" ]]; then
    echo "  A4 → #62 (existant)"
    ISSUE_NUMS[A4]="62"
    gh issue comment 62 --repo "$REPO" --body "Lié roadmap thématique (A4). Spec : docs/issues/thematic-roadmap-epic.md — Épique #${EPIC_NUMS[A]:-?}" 2>/dev/null || true
  else
    create_issue_from_key "A4" "[auth] Tests OAuth GitHub mockés" "A" "area:auth" "vague-thematic"
  fi

  local64=""
  if [[ "$DRY_RUN" -eq 0 ]]; then
    local64="$(gh issue view 64 --repo "$REPO" --json number -q .number 2>/dev/null || true)"
  fi
  if [[ "$local64" == "64" ]]; then
    echo "  A5 → #64 (existant — couverture CSRF prioritaire)"
    ISSUE_NUMS[A5]="64"
    gh issue comment 64 --repo "$REPO" --body "Roadmap thématique A5 : **tests CSRF** sur toutes routes mutantes d’abord ; refactor deps ensuite. Spec : docs/issues/thematic-roadmap-epic.md — Épique #${EPIC_NUMS[A]:-?}" 2>/dev/null || true
  else
    create_issue_from_key "A5" "[core] Couverture CSRF HTMX toutes routes mutantes" "A" "area:core" "vague-thematic"
  fi

  create_issue_from_key "A2" "[auth] Message post-OAuth si email non autorisé" "A" "area:auth" "vague-thematic"
  create_issue_from_key "G2" "[infra] Rate limit auth, invite, export, webhook" "G" "area:infra" "area:auth" "vague-thematic"
fi

if should_run_lot 2 || [[ -z "$LOT" || "$LOT" == "all" ]]; then
  echo "==> Lot 2 — SimpleUI + empty states"
  create_issue_from_key "B1" "[ui] Flags SimpleUI runtime + tests middleware" "B" "area:ui" "vague-thematic"
  create_issue_from_key "B2" "[ui] Vocabulaire Listes vs Modèles (ShowSubjectColumn)" "B" "area:ui" "vague-thematic"
  create_issue_from_key "B3" "[ui] Hub org solo minimal vs onglet Organisation" "B" "area:ui" "area:admin" "vague-thematic"
  create_issue_from_key "B0" "[ui] Moments d'unlock P0→P1 et P1→P2" "B" "area:ui" "vague-thematic"
  create_issue_from_key "B5" "[docs] Documenter paliers UI P0–P3 dans PLAN.md" "B" "area:ui" "vague-thematic"
  create_issue_from_key "B6" "[ui] Wizard : préremplir sujet si SimpleSubjectID" "B" "area:ui" "vague-thematic"
  create_issue_from_key "A1a" "[ui] Empty states — sujets" "A" "area:ui" "vague-thematic"
  create_issue_from_key "A1b" "[ui] Empty states — listes / modèles" "A" "area:ui" "vague-thematic"
  create_issue_from_key "A1c" "[ui] Empty states — revues" "A" "area:ui" "vague-thematic"
  if [[ "$DRY_RUN" -eq 0 ]] && gh issue view 63 --repo "$REPO" --json number -q .number >/dev/null 2>&1; then
    gh issue comment 63 --repo "$REPO" --body "Scindé roadmap thématique : A1a sujets, A1b listes/modèles, A1c revues (après B1–B2). Voir #${ISSUE_NUMS[A1a]:-?} #${ISSUE_NUMS[A1b]:-?} #${ISSUE_NUMS[A1c]:-?}" 2>/dev/null || true
  fi
fi

if should_run_lot 3 || [[ -z "$LOT" || "$LOT" == "all" ]]; then
  echo "==> Lot 3 — Preuve"
  create_issue_from_key "C0" "[attachments] Harden PJ — IDOR, magic bytes, disposition" "C" "area:attachments" "area:auth" "vague-thematic"
  create_issue_from_key "C1" "[core] Export preuve ZIP (CSV + manifest + sha256)" "C" "area:core" "vague-thematic"
  create_issue_from_key "C2" "[ui] Afficher hash preuve + téléchargement sur revue done" "C" "area:ui" "area:core" "vague-thematic"
  create_issue_from_key "C3" "[core] Preuve ZIP — hashes PJ (binaires optionnels plafonnés)" "C" "area:core" "area:attachments" "vague-thematic"
fi

if should_run_lot 4 || [[ -z "$LOT" || "$LOT" == "all" ]]; then
  update_issue_66
fi

if should_run_lot 5 || [[ -z "$LOT" || "$LOT" == "all" ]]; then
  echo "==> Lot 5 — Post-#66 (créées maintenant ; bloquées par #66 PASS)"
  create_issue_from_key "D1" "[notifications] Rappels échéance J-1 + badge en retard" "D" "area:notifications" "area:ui" "vague-thematic"
  create_issue_from_key "D6" "[ui] Filtres /revues gated par palier" "D" "area:ui" "vague-thematic"
  create_issue_from_key "E3prime" "[integrations] Webhooks — retry durable léger" "E" "area:integrations" "vague-thematic"
  create_issue_from_key "E6" "[integrations] Notion — erreurs UX import/export" "E" "area:integrations" "area:ui" "vague-thematic"
  create_issue_from_key "B4a" "[ui] Matrice capability P3 (doc + PageData)" "B" "area:ui" "vague-thematic"
  create_issue_from_key "B4b" "[ui] Gates UI intégrations (Notion/Jira/webhooks)" "B" "area:ui" "area:integrations" "vague-thematic"
  create_issue_from_key "B4c" "[ui] Gate UI preuve scellée" "B" "area:ui" "vague-thematic"
  create_issue_from_key "D7" "[admin] Preset ui_subject_label org" "D" "area:admin" "area:ui" "vague-thematic"
fi

# Update epic bodies with child task lists
update_epic_children() {
  local letter="$1"
  shift
  local keys=("$@")
  local epic="${EPIC_NUMS[$letter]:-}"
  [[ -n "$epic" && "$epic" != "dry" ]] || return
  local list="" k
  for k in "${keys[@]}"; do
    local n="${ISSUE_NUMS[$k]:-}"
    if [[ -n "$n" && "$n" != "dry" ]]; then
      list="${list}- [ ] #${n}
"
    fi
  done
  local intro
  intro="$(extract_section "Epic ${letter}" | head -40)"
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "  dry-run: update epic ${letter} children"
    return
  fi
  gh issue edit "$epic" --repo "$REPO" --body "$(cat <<EOF
${intro}

## Issues filles

${list}

---
Spec : docs/issues/thematic-roadmap-epic.md
EOF
)" >/dev/null
  echo "Épique ${letter} #$epic mise à jour"
}

if [[ -z "$LOT" || "$LOT" == "all" ]]; then
  echo "==> Mise à jour listes filles des épiques"
  update_epic_children A A2 A4 A5 A1a A1b A1c A3
  update_epic_children B B1 B2 B3 B0 B5 B6 B4a B4b B4c
  update_epic_children C C0 C1 C2 C3
  update_epic_children D D1 D6 D7
  update_epic_children E E3prime E6
  update_epic_children G G2
fi

echo ""
echo "OK — roadmap thématique créée sur https://github.com/${REPO}/issues"
echo "Épiques :"
for letter in A B C D E F G; do
  echo "  ${letter}: #${EPIC_NUMS[$letter]:-?}"
done
