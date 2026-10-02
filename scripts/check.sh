#!/usr/bin/env bash
# Gatekeeper local et CI — Revues
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
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
  docs/ADR-001-api-first-svelte.md
  docs/API.md
  docs/FRONTEND.md
)
for f in "${required_files[@]}"; do
  [[ -f "$f" ]] || fail "Fichier manquant : $f"
done

[[ -d api/openapi ]] || fail "Dossier manquant : api/openapi"
[[ -d frontend ]] || fail "Dossier manquant : frontend"
[[ -f frontend/package.json ]] || fail "Fichier manquant : frontend/package.json"

# ---------------------------------------------------------------------------
# 2. Interdits éco / stack (legacy)
# ---------------------------------------------------------------------------
step "Vérification interdits stack"
# SPA SvelteKit est la stack cible (ADR-001). Interdits : React/Vue SPA et bundlers hors Vite/SvelteKit.
if git grep -l -E '\b(react|vue|webpack)\b' -- '*.go' '*.html' '*.js' '*.css' 2>/dev/null | grep -vE '^(docs/|frontend/)' ; then
  fail "Framework frontend interdit détecté (React/Vue/webpack)"
fi
# Pas de retour de pages métier html/template
if [[ -d web/templates/pages ]]; then
  fail "web/templates/pages ne doit plus exister (rewrite WP-001)"
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
  export PATH="$(go env GOPATH)/bin:${PATH}"
  if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run ./...
  else
    echo "golangci-lint absent localement — CI l'exécutera"
  fi
else
  step "Pas de go.mod — vérifications Go ignorées (harness documentaire seul)"
fi

# ---------------------------------------------------------------------------
# 4. Frontend SvelteKit (build minimal)
# ---------------------------------------------------------------------------
if [[ -f frontend/package.json ]]; then
  if command -v npm >/dev/null 2>&1; then
    step "frontend npm ci + build"
    (
      cd frontend
      if [[ -f package-lock.json ]]; then
        npm ci --no-audit --no-fund
      else
        npm install --no-audit --no-fund
      fi
      npm run check
      npm run build
    )
    [[ -f frontend/build/index.html ]] || fail "frontend/build/index.html manquant après build"
  else
    echo -e "${YELLOW}WARN${NC} npm absent — build frontend ignoré localement"
  fi
fi

# ---------------------------------------------------------------------------
# 5. Cohérence schéma
# ---------------------------------------------------------------------------
step "Vérification tables canoniques"
for table in users sessions allowed_emails subjects subject_tags subject_domains \
  checklist_templates template_domains template_versions template_items \
  checklist_runs run_items run_item_events email_deliveries; do
  grep -q "CREATE TABLE ${table}" docs/schema/canonical.sql || fail "Table manquante : $table"
done

echo -e "${GREEN}OK${NC} — check.sh passé"
