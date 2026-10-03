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
# Pas de retour de pages métier html/template / HTMX
if [[ -d web/templates ]]; then
  fail "web/templates/ ne doit plus exister (rewrite WP-001 / WP-030)"
fi
# Code applicatif seulement (docs/ADR/AGENTS peuvent mentionner HTMX comme interdit).
if git grep -l -i -E '\bhtmx\b|hx-get|hx-post|hx-boost' -- \
  ':*.go' ':*.html' ':frontend/src/**/*.js' ':frontend/src/**/*.svelte' ':frontend/src/**/*.ts' \
  2>/dev/null | grep -vE '^docs/' ; then
  fail "Réintroduction HTMX / attributs hx-* détectée dans le code"
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
  # Même résolution qu'avant (GOPATH/bin d'abord), mais .golangci.yml est au format v2 :
  # une v1 ignore les exclusions et échoue à tort (misspell FR dans cmd/seed, etc.).
  export PATH="$(go env GOPATH)/bin:${PATH}"
  if command -v golangci-lint >/dev/null 2>&1; then
    lint_version="$(golangci-lint --version 2>/dev/null | grep -oE 'version v?[0-9]+' | grep -oE '[0-9]+' || echo 0)"
    if [[ "${lint_version}" -lt 2 ]]; then
      fail "golangci-lint v${lint_version}.x détecté — v2.x requis (config .golangci.yml v2, CI v2.12). Installer : https://golangci-lint.run/docs/welcome/install/"
    fi
    golangci-lint run ./...
  else
    echo "golangci-lint absent localement — CI l'exécutera"
  fi
else
  step "Pas de go.mod — vérifications Go ignorées (harness documentaire seul)"
fi

# ---------------------------------------------------------------------------
# 4. Frontend SvelteKit (npm ci / check / build + budgets SPA)
# ---------------------------------------------------------------------------
# Budgets documentés dans docs/PLAN.md (WP-005). Vendor mb mesuré, hors fail.
# Mesuré (PR-B shell + thème app.css) : JS 240 KiB / 104 KiB gz ; CSS 9,2 KiB / 3,1 KiB gz — marge ~10 %.
SPA_JS_RAW_MAX=270336      # 264 KiB — stack + Jira/Notion/Webhooks
SPA_JS_GZIP_MAX=114688     # 112 KiB gzip-9 — stack + Jira/Notion/Webhooks
SPA_CSS_RAW_MAX=10240      # 10 KiB — build/_app/**/*.css (app.css global + styles locaux résiduels)
SPA_CSS_GZIP_MAX=3584      # 3,5 KiB gzip-9

if [[ -f frontend/package.json ]]; then
  if command -v npm >/dev/null 2>&1; then
    step "frontend npm ci + check + build"
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

    step "budgets SPA (app JS/CSS hors vendor mb)"
    python3 - "$SPA_JS_RAW_MAX" "$SPA_JS_GZIP_MAX" "$SPA_CSS_RAW_MAX" "$SPA_CSS_GZIP_MAX" <<'PY' || fail "budgets SPA dépassés"
import gzip, pathlib, sys

js_raw_max, js_gz_max, css_raw_max, css_gz_max = map(int, sys.argv[1:5])
root = pathlib.Path("frontend/build/_app")
if not root.is_dir():
    print("missing frontend/build/_app", file=sys.stderr)
    sys.exit(1)

def measure(pattern: str):
    files = list(root.rglob(pattern))
    raw = sum(f.stat().st_size for f in files)
    gz = sum(len(gzip.compress(f.read_bytes(), compresslevel=9)) for f in files)
    return len(files), raw, gz

n_js, js_raw, js_gz = measure("*.js")
n_css, css_raw, css_gz = measure("*.css")
print(f"SPA JS  : {n_js} files, {js_raw} B raw / {js_gz} B gzip-9 (max {js_raw_max}/{js_gz_max})")
print(f"SPA CSS : {n_css} files, {css_raw} B raw / {css_gz} B gzip-9 (max {css_raw_max}/{css_gz_max})")

mb = pathlib.Path("web/static/vendor/jeb-maker-mb")
if mb.is_dir():
    mfiles = [p for p in mb.rglob("*") if p.is_file()]
    mb_raw = sum(f.stat().st_size for f in mfiles)
    print(f"Vendor mb: {len(mfiles)} files, {mb_raw} B raw (mesuré, hors fail — PLAN.md)")

ok = True
if js_raw > js_raw_max or js_gz > js_gz_max:
    print("FAIL JS app budget", file=sys.stderr)
    ok = False
if css_raw > css_raw_max or css_gz > css_gz_max:
    print("FAIL CSS app budget", file=sys.stderr)
    ok = False
sys.exit(0 if ok else 1)
PY
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
