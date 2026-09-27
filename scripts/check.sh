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
    # App JS budget (15 KiB): first-party scripts only.
    # Exclude web/static/vendor/ (e.g. @jeb-maker/reports, @jeb-maker/mb + Lit) from this sum.
    js_size=$(find web/static \( -path 'web/static/vendor' -o -path 'web/static/vendor/*' \) -prune -o -name '*.js' -type f -print0 2>/dev/null | xargs -0 cat 2>/dev/null | wc -c || echo 0)
    if [[ "$js_size" -gt 15360 ]]; then
      fail "JS total $js_size octets > 15 Ko (hors vendor/)"
    fi
    js_gzip_sum=0
    while IFS= read -r -d '' f; do
      g=$(gzip -c "$f" | wc -c)
      js_gzip_sum=$((js_gzip_sum + g))
    done < <(find web/static \( -path 'web/static/vendor' -o -path 'web/static/vendor/*' \) -prune -o -name '*.js' -type f -print0 2>/dev/null)

    vendor_js_raw=0
    vendor_js_gzip=0
    while IFS= read -r -d '' f; do
      vendor_js_raw=$((vendor_js_raw + $(wc -c < "$f")))
      vendor_js_gzip=$((vendor_js_gzip + $(gzip -c "$f" | wc -c)))
    done < <(find web/static/vendor -name '*.js' -type f -print0 2>/dev/null)

    # CSS découpé : core (app.css) léger pour le parcours commun ; run/editor à la demande.
    # Vérité 3G = gzip par fichier (somme) ; brut total = garde-fou éditorial.
    # Vendor CSS (mb tokens-core / bridge) excluded — same policy as vendor JS.
    css_core=$(wc -c < web/static/css/app.css 2>/dev/null || echo 0)
    if [[ "$css_core" -gt 24576 ]]; then
      fail "CSS core app.css $css_core octets > 24 Ko"
    fi
    css_core_gzip=$(gzip -c web/static/css/app.css 2>/dev/null | wc -c || echo 0)
    if [[ "$css_core_gzip" -gt 8192 ]]; then
      fail "CSS core gzip $css_core_gzip octets > 8 Ko"
    fi
    css_size=$(find web/static \( -path 'web/static/vendor' -o -path 'web/static/vendor/*' \) -prune -o -name '*.css' -type f -print0 2>/dev/null | xargs -0 cat 2>/dev/null | wc -c || echo 0)
    if [[ "$css_size" -gt 40960 ]]; then
      fail "CSS total $css_size octets > 40 Ko (hors vendor/)"
    fi
    css_gzip_sum=0
    while IFS= read -r -d '' f; do
      g=$(gzip -c "$f" | wc -c)
      css_gzip_sum=$((css_gzip_sum + g))
    done < <(find web/static \( -path 'web/static/vendor' -o -path 'web/static/vendor/*' \) -prune -o -name '*.css' -type f -print0 2>/dev/null)
    if [[ "$css_gzip_sum" -gt 12288 ]]; then
      fail "CSS gzip cumulé $css_gzip_sum octets > 12 Ko (hors vendor/)"
    fi

    vendor_css_raw=0
    vendor_css_gzip=0
    while IFS= read -r -d '' f; do
      vendor_css_raw=$((vendor_css_raw + $(wc -c < "$f")))
      vendor_css_gzip=$((vendor_css_gzip + $(gzip -c "$f" | wc -c)))
    done < <(find web/static/vendor -name '*.css' -type f -print0 2>/dev/null)

    echo "  app JS     : ${js_size} o brut / ${js_gzip_sum} o gzip (seuil 15 Ko brut)"
    echo "  vendor JS  : ${vendor_js_raw} o brut / ${vendor_js_gzip} o gzip (hors seuil app ; lazy-load reports)"
    echo "  app CSS    : ${css_size} o brut / ${css_gzip_sum} o gzip (seuil 40 Ko / 12 Ko)"
    echo "  vendor CSS : ${vendor_css_raw} o brut / ${vendor_css_gzip} o gzip"
    # Soft budget: typical authenticated shell = htmx+app + mb-boot + tokens (reports lazy).
    # Warn above 45 KiB gzip shell estimate (mb-boot ~22k + htmx/app ~3k + tokens ~3k).
    shell_gzip_est=$((js_gzip_sum + vendor_js_gzip + css_gzip_sum + vendor_css_gzip))
    # Subtract reports from shell estimate when present (lazy-loaded on demand).
    reports_gzip=0
    if [[ -f web/static/vendor/jeb-maker-reports/reports.min.js ]]; then
      reports_gzip=$(gzip -c web/static/vendor/jeb-maker-reports/reports.min.js | wc -c)
      reports_gzip=$((reports_gzip + $(gzip -c web/static/vendor/jeb-maker-reports/init.js | wc -c)))
    fi
    shell_gzip_no_reports=$((shell_gzip_est - reports_gzip))
    echo "  transfert shell estimé (sans reports lazy) : ${shell_gzip_no_reports} o gzip"
    if [[ "$shell_gzip_no_reports" -gt 46080 ]]; then
      echo -e "${YELLOW}WARN${NC} shell gzip estimé > 45 Ko — vérifier lazy-load mb/reports (voir PLAN.md § Budget sobriété)"
    fi
  fi
else
  step "Pas de go.mod — vérifications Go ignorées (harness documentaire seul)"
fi

# ---------------------------------------------------------------------------
# 4. Cohérence schéma
# ---------------------------------------------------------------------------
step "Vérification tables canoniques"
for table in users sessions allowed_emails subjects subject_tags subject_domains \
  checklist_templates template_domains template_versions template_items \
  checklist_runs run_items run_item_events email_deliveries; do
  grep -q "CREATE TABLE ${table}" docs/schema/canonical.sql || fail "Table manquante : $table"
done

echo -e "${GREEN}OK${NC} — check.sh passé"
