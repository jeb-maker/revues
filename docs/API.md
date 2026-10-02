# API — conventions Revues

Contrat HTTP JSON pour le front SvelteKit et éventuels clients. Complète [ADR-001](./ADR-001-api-first-svelte.md).

## Principes

1. **OpenAPI 3** est la source de vérité des routes `/api/v1/**`.
2. Handlers / types serveur **générés** avec **oapi-codegen** — ne pas dupliquer les DTO à la main.
3. SQL via **sqlc** (ou store existant jusqu’à bascule complète) — pas de SQL dans les handlers.
4. Erreurs JSON homogènes ; **pas de HTML** dans `/api/v1`.
5. RBAC et IDOR **côté serveur** (matrice [RBAC.md](./RBAC.md)).

## Toolchain OpenAPI (WP-002)

| Élément | Choix |
|---------|--------|
| Spec | [`api/openapi/openapi.yaml`](../api/openapi/openapi.yaml) |
| Codegen serveur | **oapi-codegen** `v2.4.1` (`chi-server` + `models`) — pas ogen |
| Config | [`api/openapi/oapi-codegen.yaml`](../api/openapi/oapi-codegen.yaml) |
| Script | `make api` ou [`./scripts/generate-api.sh`](../scripts/generate-api.sh) |
| Sortie | `internal/api/v1/oapi.gen.go` (**commité** ; régénérer après chaque changement de spec) |
| Montage chi | `r.Route("/api/v1", …)` + `apiv1.HandlerFromMux` dans `internal/web/router.go` |

Pourquoi oapi-codegen : intégration native chi (`HandlerFromMux`), compatible Go 1.22 du module, surface minimale pour démarrer. Le client TypeScript (WP-005) consommera la même spec.

Régénération :

```bash
make api
# ou
./scripts/generate-api.sh
```

Le code généré est **commité** (pas de génération obligatoire en CI pour l’instant) : les PR qui touchent la spec doivent inclure `oapi.gen.go` à jour. La CI exécute `./scripts/check.sh` (compile + tests) sur le code présent dans le dépôt.

## Surface

| Préfixe | Rôle |
|---------|------|
| `GET /healthz` | Health probe infra (texte `ok`, hors OpenAPI) |
| `GET /api/v1/health` | Health API versionnée (JSON `{"status":"ok"}`, tag `system`) |
| `/api/v1/**` | API métier versionnée |
| `/auth/**` | Démarrage OAuth / callbacks (peuvent rediriger) |
| `/` + assets | SPA SvelteKit (static) servie par Go en prod |

## Auth & CSRF

- Cookie `revues_session` HttpOnly, Secure, SameSite=Lax (comportement actuel).
- `GET /api/v1/me` (ou bootstrap) renvoie user + **csrf_token**.
- Mutations (`POST`/`PATCH`/`PUT`/`DELETE`) : header `X-CSRF-Token`.
- Same-origin SPA ; pas de JWT en cookie.

## Erreurs

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Un commentaire est obligatoire pour le statut Non validé."
  }
}
```

| HTTP | Usage |
|------|--------|
| 400 | Validation / bad request |
| 401 | Non authentifié |
| 403 | Authentifié sans droit (rare ; préférer 404 si fuite d’existence) |
| 404 | Absent **ou** non visible (IDOR) |
| 409 | Conflit métier (ex. run non éditable) |
| 500 | Erreur interne (pas de détail sensible) |

## OpenAPI — organisation

- Spec : `api/openapi/openapi.yaml` (ou découpe par domaine + bundler).
- `operationId` stable → noms générés stables.
- Tags = domaines (`system`, `auth`, `orgs`, `subjects`, `templates`, `runs`, `admin`, `integrations`).
- Régénérer après chaque changement de spec (`make api` / `./scripts/generate-api.sh`).

## sqlc

- Requêtes : `internal/store/sqlc/` (ou chemin fixé par WP-003).
- Schéma aligné sur [schema/canonical.sql](./schema/canonical.sql) + migrations goose.
- Ne pas introduire d’ORM.

## Tests API

- Table-driven sur handlers / services.
- `area:auth` : session, CSRF, RBAC.
- Pas d’appel réseau réel aux intégrations (mocks).
