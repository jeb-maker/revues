# Front — SvelteKit + miniature-broccoli

Complète [ADR-001](./ADR-001-api-first-svelte.md).

## Stack

| Élément | Choix |
|---------|--------|
| Framework | SvelteKit (mode SPA / `adapter-static`) |
| UI | `@jeb-maker/mb` Web Components (Lit), tokens CSS |
| Données | Client **TypeScript généré** (`openapi-typescript` + `openapi-fetch`) |
| Build | Vite (via SvelteKit) |

## Emplacement

```
frontend/                 # app SvelteKit
  src/routes/             # pages
  src/lib/api/            # schema.d.ts généré + wrappers (client.ts)
  src/lib/mb/             # ensureMb() — charge tokens + CE depuis /static/vendor
  src/lib/components/     # composition mb + logique écran
web/static/vendor/jeb-maker-mb/  # tokens + mb-boot.js (servi par Go)
```

Go sert le build `frontend/build` (ou `REVUES_SPA_DIR`) en production ; en dev : proxy Vite → API `:8080`.

### Démarrage front

```bash
# Terminal 1 — API
go run ./cmd/revues          # :8080 (sert aussi /static/vendor/mb)

# Terminal 2 — SPA
cd frontend && npm ci && npm run dev   # http://localhost:5173
```

### Proxy Vite → Go (dev)

Configurée dans [`frontend/vite.config.ts`](../frontend/vite.config.ts) :

| Préfixe | Cible |
|---------|--------|
| `/api` | `http://127.0.0.1:8080` |
| `/auth` | idem (OAuth) |
| `/healthz` | idem |
| `/static` | idem (vendor **mb**) |

En production, SPA et API sont same-origin : pas de proxy.

Build servi par Go :

```bash
cd frontend && npm ci && npm run build
go run ./cmd/revues                   # sert frontend/build ; sinon page stub
```

### Client OpenAPI

```bash
make frontend-api
# ou : ./scripts/generate-frontend-api.sh
# ou : cd frontend && npm run generate:api
```

| Élément | Chemin |
|---------|--------|
| Spec | `api/openapi/openapi.yaml` |
| Types générés | `frontend/src/lib/api/schema.d.ts` (**commité**) |
| Client | `frontend/src/lib/api/client.ts` (`openapi-fetch`, cookies + CSRF) |

### mb (tokens + CE)

```ts
import { ensureMb } from '$lib/mb';
// appelé depuis le layout racine — injecte tokens-core.css, mb-bridge.css, mb-boot.js
```

Assets : `/static/vendor/jeb-maker-mb/` (embed Go). Voir `web/static/vendor/jeb-maker-mb/README.md`.

## Règles

1. **Pas de `html/template` Go** ni HTMX pour l’UI métier.
2. Composants visuels : préférer **mb-*** (`mb-button`, `mb-table`, `mb-input`, …) avant HTML custom.
3. Aucune règle RBAC « de confiance » uniquement côté client — le serveur tranche ; le front masque seulement.
4. CSRF : lire le token bootstrap, l’envoyer en `X-CSRF-Token` sur chaque mutation (`createApiClient({ csrfToken })`).
5. Pas de polling ni WebSocket (inchangé).
6. i18n hors scope sauf issue dédiée.

## Budgets (SPA)

Appliqués par `./scripts/check.sh` — détail [PLAN.md](./PLAN.md) :

| Métrique | Seuil |
|----------|-------|
| JS app (`frontend/build/_app/**/*.js`) | ≤ 264 KiB brut / ≤ 112 KiB gzip-9 |
| CSS app (`frontend/build/_app/**/*.css`) | ≤ 40 KiB brut / ≤ 17 KiB gzip-9 |
| Vendor mb | mesuré, hors fail strict |

## Tests front

- Minimum : `npm run check` / `npm run build` vert en CI (via `check.sh`).
- Tests e2e hors scope v1 rewrite sauf issue dédiée.
- Parité fonctionnelle validée manuellement ou via tests API pour la logique.

## Référence mb

Voir `web/static/vendor/jeb-maker-mb/README.md` et upstream [miniature-broccoli](https://github.com/jeb-maker/miniature-broccoli).
