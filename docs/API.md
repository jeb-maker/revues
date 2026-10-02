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

Pourquoi oapi-codegen : intégration native chi (`HandlerFromMux`), compatible Go 1.22 du module, surface minimale pour démarrer.

### Client TypeScript (WP-005)

| Élément | Choix |
|---------|--------|
| Outil | **openapi-typescript** `7.x` + **openapi-fetch** |
| Script | `make frontend-api` / [`./scripts/generate-frontend-api.sh`](../scripts/generate-frontend-api.sh) |
| Types | `frontend/src/lib/api/schema.d.ts` (**commité**) |
| Wrappers | `frontend/src/lib/api/client.ts` (credentials + CSRF) |

Régénération serveur :

```bash
make api
# ou
./scripts/generate-api.sh
```

Régénération client front :

```bash
make frontend-api
```

Le code généré est **commité** (pas de génération obligatoire en CI pour l’instant) : les PR qui touchent la spec doivent inclure `oapi.gen.go` **et** `schema.d.ts` à jour. La CI exécute `./scripts/check.sh` (Go + npm front) sur le code présent dans le dépôt.

## Surface

| Préfixe | Rôle |
|---------|------|
| `GET /healthz` | Health probe infra (texte `ok`, hors OpenAPI) |
| `GET /api/v1/health` | Health API versionnée (JSON `{"status":"ok"}`, tag `system`) |
| `GET /api/v1/bootstrap` | Auth bootstrap : user optionnel + `csrf_token` (+ cookie guest) |
| `GET /api/v1/me` | Utilisateur courant (401 sans session) |
| `POST /api/v1/auth/login` | Login email/mot de passe (CSRF guest) |
| `POST /api/v1/auth/register` | Inscription locale (CSRF guest) |
| `POST /api/v1/auth/logout` | Logout (CSRF session) |
| `GET/POST /api/v1/templates` | Catalogue / création modèle (v1) |
| `GET/PUT/DELETE /api/v1/templates/{id}` | Détail / save (= nouvelle version) / archive |
| `GET/POST /api/v1/templates/{id}/versions` | Historique / publier version |
| `GET /api/v1/templates/{id}/versions/{n}` | Snapshot immuable |
| `GET /api/v1/runs` | Liste des revues visibles (filtre status/q) |
| `POST /api/v1/subjects/{id}/runs` | Lancer une revue (snapshot transactionnel) |
| `GET /api/v1/runs/{id}` | Détail + items + progression |
| `PATCH /api/v1/runs/{id}/items/{itemId}` | Status / commentaire / assignation |
| `POST /api/v1/runs/{id}/complete` | Clôturer (bloque si required pending) |
| `GET /api/v1/me/tasks` | Mes tâches (items assignés, filtre status/q) |
| `GET\|PUT\|DELETE /api/v1/admin/settings/smtp` | Config SMTP chiffrée (password masqué) |
| `POST /api/v1/admin/settings/smtp/test` | Email de test SMTP |
| `GET /api/v1/admin/integrations` | Hub intégrations (états + config_path SPA) |
| `GET\|PUT\|DELETE /api/v1/admin/integrations/jira` | Config Jira Cloud chiffrée (jeton masqué) |
| `POST /api/v1/admin/integrations/jira/test` | Test connexion Jira (`/myself`, safehttp) |
| `GET\|PUT\|POST /api/v1/runs/{runId}/items/{itemId}/jira` | État / lier / créer issue Jira |
| `GET\|POST /api/v1/runs/{runId}/items/{itemId}/attachments` | Métadonnées / upload pièce jointe |
| `GET .../attachments/{attachmentId}` | Download (`Content-Disposition: attachment`) |
| `/api/v1/**` | API métier versionnée |
| `GET /auth/github/start` · `/callback` | OAuth GitHub (redirects browser) |
| `/` + assets | SPA SvelteKit (static) servie par Go en prod |

## Auth & CSRF

- Cookie `revues_session` HttpOnly, Secure, SameSite=Lax (comportement actuel).
- Cookie `revues_guest` court pour CSRF des formulaires login/register anonymes.
- `GET /api/v1/bootstrap` et `GET /api/v1/me` renvoient **csrf_token**.
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
- Régénérer après chaque changement de spec : `make api` (serveur) **et** `make frontend-api` (client TS).

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
