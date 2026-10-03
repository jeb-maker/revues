# Vendored `@jeb-maker/mb` (+ Lit peer, bundled)

- **Version used by Revues: `0.4.1`** (Git tag `v0.4.1`)
- **Source**: https://github.com/jeb-maker/miniature-broccoli
- **Lit**: peer `^3.2.0` (build used `lit@3.3.x`) is **bundled into `mb-boot.js`** so the browser needs no import map. Vendor files are measured but excluded from the SPA app budgets (`scripts/check.sh` only fails on `frontend/build/_app/**`).

## Layout

| Path | Role |
|------|------|
| `mb-boot.js` | Host ESM loader — registers custom elements (Lit inlined) |
| `tokens/tokens-core.css` | **Preferred host entry** — variables + anti-FOUC, no `html`/`body` reset |
| `tokens/reference.css`, `semantic.css` | Imported by `tokens-core.css` |
| `mb-bridge.css` | Host coexistence (accent/font remap, spacing) |

## Host load order (SvelteKit SPA)

Assets are embedded by Go (`web/fs.go`) and served under `/static/vendor/jeb-maker-mb/`. In Vite dev, `frontend/vite.config.ts` proxies `/static` to the Go API on `:8080`.

`frontend/src/lib/mb/index.ts` exposes `ensureMb()`, called once from the root layout (`frontend/src/routes/+layout.svelte`, `onMount`). It injects, in order:

1. `tokens/tokens-core.css` (`<link rel="stylesheet">`)
2. `mb-bridge.css` (`<link rel="stylesheet">`)
3. `mb-boot.js` (`<script type="module" data-mb-boot>`)

Page styles live in Svelte `<style>` blocks (bundled by Vite) and cascade after the vendor stylesheets. `ensureMb()` is idempotent (guards on existing `<link>`/`<script>`).

TypeScript typings for the custom elements used in templates: `frontend/src/lib/mb/elements.d.ts`.

## `mb-boot.js` registers

`mb-button`, `mb-badge`, `mb-alert`, `mb-card`, `mb-input`, `mb-textarea`, `mb-checkbox`, `mb-select`, `mb-modal`, `mb-progress`, `mb-segmented-control`, `mb-empty-state`, `mb-pagination`, `mb-toast`, `mb-radio`, `mb-radio-group`, `mb-tag`, `mb-breadcrumbs`, `mb-nav`, `mb-nav-toggle`, `mb-avatar`, `mb-spinner`, `mb-toolbar`, `mb-table`, `mb-table-row`, `mb-table-cell`.

## Rebuild (maintainers)

### Bundle courant (`mb-boot.js` — toutes les CE)

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"   # or any Node ≥ 20
git clone --depth 1 --branch v0.4.1 https://github.com/jeb-maker/miniature-broccoli.git /tmp/mb-0.4.1
cd /tmp/mb-0.4.1 && npm ci && npm run build
COMPS="button badge alert card input textarea checkbox select modal progress segmented-control empty-state pagination toast radio radio-group tag breadcrumbs nav nav-toggle avatar spinner toolbar table"
{ for c in $COMPS; do echo "import './dist/components/\$c.js';"; done; } > boot-entry.js
npx esbuild boot-entry.js --bundle --format=esm \
  --outfile=/path/to/revues/web/static/vendor/jeb-maker-mb/mb-boot.js \
  --minify --legal-comments=none
cp dist/tokens/tokens-core.css dist/tokens/reference.css dist/tokens/semantic.css \
  /path/to/revues/web/static/vendor/jeb-maker-mb/tokens/
```

`internal/web/staticassets_test.go` (`TestVendoredMBBundlePresent`) vérifie la présence des fichiers et l'enregistrement des CE dans `mb-boot.js`.

### Imports atomiques (cible future)

Upstream recommande d’importer **seulement** les CE utilisées par page (pas de barrel). Pour y arriver côté Revues :

1. Rebuild chaque `dist/components/<name>.js` (+ chunk Lit partagé) vers `web/static/vendor/jeb-maker-mb/`
2. Remplacer le `mb-boot.js` monolithe de `ensureMb()` par des imports par route (ou passer `@jeb-maker/mb` en dépendance npm du `frontend/` et laisser Vite tree-shaker)
3. Garder `mb-table*` hors pages login

Aujourd’hui Revues ship **`mb-boot.js` monolithe** (simple, une requête).

## 0.4.1 (vs 0.3.1)

- `mb-table` / `mb-table-row` / `mb-table-cell` — responsive lists, sections, sort, reorder
- Section `meta` / `count: false` / `hide-count`; `sticky-header`; `reorder-label` / `sort-label`; cell `hide-label` / `actions`

## Consumed in Revues

Pages SvelteKit (`frontend/src/routes/**`) : `mb-button`, `mb-alert`, `mb-input`, `mb-textarea`, `mb-badge`, `mb-tag`, `mb-spinner`, `mb-empty-state`, `mb-table` / `mb-table-row` / `mb-table-cell`.

**Still host-owned**: template-editor table (`frontend/src/lib/components/TemplateEditor.svelte`, DnD + indices), native `confirm()` for destructive actions.
