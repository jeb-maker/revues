# API — conventions Revues

Contrat HTTP JSON pour le front SvelteKit et éventuels clients. Complète [ADR-001](./ADR-001-api-first-svelte.md).

## Principes

1. **OpenAPI 3** est la source de vérité des routes `/api/v1/**`.
2. Handlers / types serveur **générés** (`ogen` ou `oapi-codegen`) — ne pas dupliquer les DTO à la main.
3. SQL via **sqlc** (ou store existant jusqu’à bascule complète) — pas de SQL dans les handlers.
4. Erreurs JSON homogènes ; **pas de HTML** dans `/api/v1`.
5. RBAC et IDOR **côté serveur** (matrice [RBAC.md](./RBAC.md)).

## Surface

| Préfixe | Rôle |
|---------|------|
| `GET /healthz` | Health (texte ou JSON simple, hors versioning) |
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
- Tags = domaines (`auth`, `orgs`, `subjects`, `templates`, `runs`, `admin`, `integrations`).
- Régénérer après chaque changement de spec (`make api` / script documenté dans le WP fondation).

## sqlc

| Élément | Chemin |
|---------|--------|
| Config | `sqlc.yaml` (engine SQLite, schéma `docs/schema/canonical.sql`) |
| Requêtes SQL | `internal/store/queries/*.sql` |
| Code généré | `internal/store/sqlc/` (commité ; ne pas éditer à la main) |
| Régénération | `make sqlc` ou `./scripts/generate-sqlc.sh` (sqlc ≥ 1.27) |

- Schéma source sqlc = [schema/canonical.sql](./schema/canonical.sql) (normatif) ; exécution runtime = migrations goose.
- **Règle** : SQL uniquement via sqlc **ou** `internal/store` (pont legacy). Pas de SQL dans handlers / features.
- Pont actuel : `users` / `sessions` passent par sqlc ; `Store.Queries()` expose le client généré. Autres domaines encore en SQL manuel dans `internal/store/*.go` (migration progressive).
- Ne pas introduire d’ORM.

## Tests API

- Table-driven sur handlers / services.
- `area:auth` : session, CSRF, RBAC.
- Pas d’appel réseau réel aux intégrations (mocks).
