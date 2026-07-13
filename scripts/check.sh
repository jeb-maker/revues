#!/usr/bin/env bash
# Gatekeeper local et CI — Revues
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

step() { echo -e "${GREEN}==>${NC} $1"; }
fail() { echo -e "${RED}FAIL:${NC} $1" >&2; exit 1; }

# ---------------------------------------------------------------------------
# 1. Fichiers harness obligatoires
# ---------------------------------------------------------------------------
step "Vérification harness documentaire"
required_files=(
  AGENTS.md
  docs/PLAN.md
  docs/CONVENTIONS.md
  docs/GO.md
  docs/DEFINITION_OF_DONE.md
  docs/RBAC.md
  docs/schema/canonical.sql
  docs/REVIEW_ADVERSE.md
)
for f in "${required_files[@]}"; do
  [[ -f "$f" ]] || fail "Fichier manquant : $f"
done

# ---------------------------------------------------------------------------
# 2. Interdits éco / stack
# ---------------------------------------------------------------------------
step "Vérification interdits stack"
if git grep -l -E '\b(react|vue|webpack|vite|svelte)\b' -- '*.go' '*.html' '*.js' '*.css' 2>/dev/null | grep -v docs/; then
  fail "Framework frontend interdit détecté"
fi

# ---------------------------------------------------------------------------
# 3. Vérifications Go (si module présent)
# ---------------------------------------------------------------------------
if [[ -f go.mod ]]; then
  step "gofmt"
  unformatted=$(gofmt -l . 2>/dev/null | grep -v '^$' || true)
  [[ -z "$unformatted" ]] || fail "Fichiers non formatés : $unformatted"

  step "go vet"
  go vet ./...

  step "go test"
  go test -race -count=1 ./...

  step "go build"
  go build -o /tmp/revues ./cmd/revues

  step "go mod tidy check"
  go mod tidy
  if ! git diff --exit-code go.mod go.sum 2>/dev/null; then
    fail "go.mod/go.sum non à jour — exécuter go mod tidy"
  fi

  step "golangci-lint"
  if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run ./...
  else
    echo "golangci-lint absent localement — CI l'exécutera"
  fi

  step "Taille assets static"
  if [[ -d web/static ]]; then
    js_size=$(find web/static -name '*.js' -exec cat {} + 2>/dev/null | wc -c || echo 0)
    if [[ "$js_size" -gt 15360 ]]; then
      fail "JS total $js_size octets > 15 Ko"
    fi
    css_size=$(wc -c < web/static/css/app.css)
    if [[ "$css_size" -gt 20480 ]]; then
      fail "CSS app.css $css_size octets > 20 Ko"
    fi
  fi
else
  step "Pas de go.mod — vérifications Go ignorées (harness documentaire seul)"
fi

# ---------------------------------------------------------------------------
# 4. Cohérence schéma
# ---------------------------------------------------------------------------
step "Vérification tables canoniques"
for table in users sessions allowed_emails projects project_members \
  checklist_templates template_versions template_items \
  checklist_runs run_items run_item_events; do
  grep -q "CREATE TABLE ${table}" docs/schema/canonical.sql || fail "Table manquante : $table"
done

echo -e "${GREEN}OK${NC} — check.sh passé"
