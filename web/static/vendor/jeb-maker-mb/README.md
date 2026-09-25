# Vendored `@jeb-maker/mb` (+ Lit peer, bundled)

- **Version used by Revues: `0.4.1`** (Git tag `v0.4.1`)
- **Source**: https://github.com/jeb-maker/miniature-broccoli
- **Lit**: peer `^3.2.0` (build used `lit@3.3.x`) is **bundled into `mb-boot.js`** so the browser needs no import map. Not counted in the 15 KiB app JS budget (`scripts/check.sh` excludes `web/static/vendor/`).

## Layout

| Path | Role |
|------|------|
| `mb-boot.js` | Host ESM loader — registers custom elements (Lit inlined) |
| `tokens/tokens-core.css` | **Preferred host entry** — variables + anti-FOUC, no `html`/`body` reset |
| `tokens/reference.css`, `semantic.css` | Imported by `tokens-core.css` |
| `mb-bridge.css` | Host coexistence (accent/font remap, spacing) |

## Host load order (see `base.html`)

1. `app.css` (Revues shell)
2. `tokens/tokens-core.css`
3. `mb-bridge.css`
4. `mb-boot.js` (`type="module"`)

Docs upstream: [`docs/go-htmx.md`](https://github.com/jeb-maker/miniature-broccoli/blob/v0.4.1/docs/go-htmx.md).

## `mb-boot.js` registers

`mb-button`, `mb-badge`, `mb-alert`, `mb-card`, `mb-input`, `mb-textarea`, `mb-checkbox`, `mb-select`, `mb-modal`, `mb-progress`, `mb-segmented-control`, `mb-empty-state`, `mb-pagination`, `mb-toast`, `mb-radio`, `mb-radio-group`, `mb-tag`, `mb-breadcrumbs`, `mb-nav`, `mb-nav-toggle`, `mb-avatar`, `mb-spinner`, `mb-toolbar`, `mb-table`, `mb-table-row`, `mb-table-cell`.

## Rebuild (maintainers)

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

## 0.4.1 (vs 0.3.1)

- `mb-table` / `mb-table-row` / `mb-table-cell` — responsive lists, sections, sort, reorder
- Section `meta` / `count: false` / `hide-count`; `sticky-header`; `reorder-label` / `sort-label`; cell `hide-label` / `actions`
- Docs: HTMX `outerHTML` row-swap + sections contract

## Consumed in Revues

Shell, forms, lists (`mb-table`), run items (sections + HTMX row swap).

**Still host-owned**: template-editor table (DnD + indices), `hx-confirm`, noscript bug-report form.
