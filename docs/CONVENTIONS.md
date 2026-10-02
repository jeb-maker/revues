# Conventions — Revues

Normes de code et d'architecture. Tout agent et contributeur les suit.

Stack cible : [ADR-001-api-first-svelte.md](./ADR-001-api-first-svelte.md).

## Arborescence

```
cmd/revues/                 # point d'entrée
api/openapi/                # OpenAPI source de vérité
frontend/                   # SvelteKit + mb
internal/
  features/                 # services métier (vertical)
  auth/                     # OAuth, sessions, CSRF, RBAC
  store/                    # accès données (seul package SQL)
    queries/                # SQL source sqlc
    sqlc/                   # code généré (make sqlc)
  web/                      # router API, middleware, static SPA
  integrations/             # jira, notion, webhooks
  notifications/ attachments/ crypto/ config/
migrations/                 # goose SQL
sqlc.yaml                   # config sqlc
scripts/generate-sqlc.sh    # régénération
data/                       # SQLite + PJ (gitignored)
docs/rewrite/               # orchestration rewrite agents
```

## Go

Voir le guide complet : **[GO.md](./GO.md)** — obligatoire pour agents.

Résumé :
- Go **1.22+**
- Handlers fins (générés / fins) : logique dans `internal/features/` ou services
- Erreurs wrappées : `fmt.Errorf("context: %w", err)` — jamais ignorées
- `context.Context` propagé sur tout I/O
- Pas d'ORM : **sqlc** + SQL paramétré dans `internal/store/`
- Interfaces petites, côté consommateur
- Tests table-driven ; `go test -race`
- `log/slog` structuré — pas de secrets loggés
- Pas de `panic` hors `main`

## SQL

- Migrations : `migrations/NNNNN_description.sql` via goose
- Schéma normatif : [schema/canonical.sql](./schema/canonical.sql)
- **Accès données** : sqlc (`internal/store/queries` → `internal/store/sqlc`) ou pont `internal/store` — **jamais** de SQL dans les handlers
- Régénérer après changement de requête : `make sqlc`
- Dates : ISO 8601 UTC en `TEXT`
- Enums : `TEXT` + `CHECK` constraint
- `PRAGMA foreign_keys=ON` à chaque connexion
- Pool SQLite : `REVUES_DB_MAX_OPEN_CONNS` (défaut 10)

## HTTP

Voir **[API.md](./API.md)**.

- Métier : JSON `/api/v1/**` (OpenAPI)
- Auth browser : `/auth/**` (redirects OAuth OK)
- SPA : assets statiques SvelteKit servis par Go en prod
- Admin org : routes sous tag/paths admin — `RequireOrgAdmin`
- IDs : valider existence **et** permission (IDOR → 404)

## Front

Voir **[FRONTEND.md](./FRONTEND.md)**.

- SvelteKit SPA + Web Components **mb**
- Client TS généré depuis OpenAPI
- CSRF header sur mutations
- **Interdit** : nouvelles pages `html/template` / HTMX métier

## RBAC

Voir [RBAC.md](./RBAC.md). Règle : **deny by default**.

- `403` = pas le droit (usage rare)
- `404` = ressource absente **ou** non visible (pas de fuite d'existence)

## Configuration

Préfixe `REVUES_` :

| Variable | Description |
|----------|-------------|
| `REVUES_ADDR` | `:8080` |
| `REVUES_DATABASE_PATH` | `data/revues.db` |
| `REVUES_DB_MAX_OPEN_CONNS` | Taille du pool SQLite (défaut `10`) |
| `REVUES_SESSION_SECRET` | 32+ octets aléatoires |
| `REVUES_ENCRYPTION_KEY` | 32 octets base64 (AES-256-GCM) |
| `REVUES_GITHUB_CLIENT_ID` | OAuth |
| `REVUES_GITHUB_CLIENT_SECRET` | OAuth |
| `REVUES_BASE_URL` | `https://revues.example.com` |

## Chiffrement settings

- Algorithme : **AES-256-GCM**
- Clé : `REVUES_ENCRYPTION_KEY` en env uniquement
- Jamais logger credentials déchiffrés

## Webhooks

- Signature : `X-Revues-Signature: sha256=<hmac>`
- Payload : `event_id` UUID stable pour idempotence
- Anti-SSRF : IP privées refusées, timeout 5s, https only (localhost http en dev), re-check à chaque retry
- Retry durable : `webhook_deliveries` — [WEBHOOKS.md](./WEBHOOKS.md)

## Uploads

- Magic bytes, pas seulement extension
- Nom stockage : UUID
- `Content-Disposition: attachment`
- Auth + contrôle sujet/run sur chaque GET

## Commits

```
type(scope): description courte

Types : feat, fix, docs, test, chore, refactor
Scope : auth, api, frontend, subjects, runs, admin, integrations, infra
```

## Interdits

- ORM (GORM, ent runtime, etc.) — sqlc uniquement
- Quarkus / stack JVM (hors décision ADR-001)
- Redis, Elasticsearch, Kafka
- JWT en cookie (sessions ID en base)
- SQL dans les handlers
- `panic()` en production (sauf main)
- Réintroduire HTMX / templates Go pour l'UI métier
