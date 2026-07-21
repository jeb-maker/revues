# Vendored @jeb-maker/reports IIFE bundle (not counted in app JS budget; see scripts/check.sh).

- **Version**: 0.3.0
- **Source**: https://github.com/jeb-maker/reports
- **Bundle**: `reports.min.js` (+ `reports.min.js.map`) from the package `dist/` IIFE build

## Init (`init.js`)

Host wiring for Revues:

- Adapter `webhook` → `POST /signaler/api` (same-origin, CSRF via `meta[name="csrf-token"]`)
- Locale `fr`
- Metadata from `#revues-reports-meta` JSON (server-rendered)
- **autoReport** (0.3.0): `{ errors: true, maxPerSession: 5, cooldownMs: 30000 }` — opt-in uncaught JS / unhandledrejection reports (no screenshot; signature dedup; fail-safe against report loops)

## 0.3.0 highlights

- `autoReport: { errors: true }` — automatic full reports on uncaught errors / unhandled rejections (`trigger: 'auto:error'`)
- Modal close/cancel fix (`.rp-backdrop` no longer overrides `[hidden]`)
