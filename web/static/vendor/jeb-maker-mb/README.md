# Vendored `@jeb-maker/mb` (+ Lit peer, bundled)

- **Version used by Revues: `0.3.0`** (Git tag `v0.3.0`)
- **Source**: https://github.com/jeb-maker/miniature-broccoli
- **Lit**: peer `^3.2.0` (build used `lit@3.3.x`) is **bundled into `mb-boot.js`** so the browser needs no import map. Not counted in the 15 KiB app JS budget (`scripts/check.sh` excludes `web/static/vendor/`).

## Layout

| Path | Role |
|------|------|
| `mb-boot.js` | Host ESM loader — registers custom elements (Lit inlined) |
| `components/*.js` | Atomic ESM entries from the package `dist/` (reference / future selective loads) |
| `lib/*.js` | Shared helpers used by atomic entries |
| `tokens/tokens-core.css` | **Preferred host entry** — variables + anti-FOUC, no `html`/`body` reset |
| `tokens/reference.css`, `semantic.css` | Imported by `tokens-core.css` |
| `tokens/tokens.css`, `typography.css` | Full baseline + fonts — **not loaded** by Revues (budget / coexistence) |
| `mb-bridge.css` | Host coexistence (accent/font remap, spacing) |

## Host load order (see `base.html`)

1. `app.css` (Revues shell)
2. `tokens/tokens-core.css`
3. `mb-bridge.css`
4. `mb-boot.js` (`type="module"`)

Docs upstream: [`docs/go-htmx.md`](https://github.com/jeb-maker/miniature-broccoli/blob/v0.3.0/docs/go-htmx.md).

## `mb-boot.js` registers

`mb-button`, `mb-badge`, `mb-alert`, `mb-card`, `mb-input`, `mb-textarea`, `mb-checkbox`, `mb-select`, `mb-modal`, `mb-progress`, `mb-segmented-control`, `mb-empty-state`, `mb-pagination`, `mb-toast`, `mb-radio`, `mb-radio-group`, `mb-tag`, `mb-breadcrumbs`, `mb-nav`, `mb-nav-toggle`, `mb-avatar`, `mb-spinner`, `mb-toolbar`.

## Rebuild (maintainers)

```bash
export PATH="/tmp/node-v20.18.1/bin:$PATH"   # or any Node ≥ 20
git clone --depth 1 --branch v0.3.0 https://github.com/jeb-maker/miniature-broccoli.git /tmp/mb-0.3.0
cd /tmp/mb-0.3.0 && npm ci && npm run build
# copy dist components/*.js, lib/*.js, tokens/*.css → this directory
# then esbuild-bundle an entry that imports every component → mb-boot.js (Lit from npm, bundled)
```

## Consumed in Revues

Shell (nav, breadcrumbs, avatar, toast), forms (`mb-*` FACE), lists (toolbar, pagination, empty-state, tags), run grid (select/textarea + `mb-change`), progress, alerts, badges, cards.

**Still host-owned**: data-tables, `hx-confirm`, noscript bug-report form (native controls — CE need JS).
