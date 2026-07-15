# Conventions — Revues

Normes de code et d'architecture. Tout agent et contributeur les suit.

## Arborescence

```
cmd/revues/main.go          # point d'entrée
internal/
  auth/                     # OAuth, sessions, CSRF
  subjects/                 # sujets, domaines
  templates/                # modèles versionnés
  runs/                     # exécutions
  items/                    # points, statuts, audit
  notifications/            # email SMTP
  integrations/             # jira, notion, webhooks
  admin/                    # users, settings
  store/                    # SQL (seul endroit avec requêtes)
  web/                      # handlers, middleware, templates
migrations/                 # goose SQL
web/templates/              # html/template
web/static/                 # CSS, HTMX
data/                       # SQLite + attachments (gitignored)
```

## Go

Voir le guide complet : **[GO.md](./GO.md)** — obligatoire pour agents.

Résumé :
- Go **1.22+**
- Handlers fins : logique dans `internal/<domain>/`
- Erreurs wrappées : `fmt.Errorf("context: %w", err)` — jamais ignorées
- `context.Context` propagé sur tout I/O
- Pas d'ORM : SQL paramétré dans `internal/store/` uniquement
- Interfaces petites, côté consommateur
- Tests table-driven ; `go test -race`
- `log/slog` structuré — pas de secrets loggés
- Pas de `panic` hors `main`

## SQL

- Migrations : `migrations/NNNNN_description.sql` via goose
- Schéma normatif : [schema/canonical.sql](./schema/canonical.sql)
- Dates : ISO 8601 UTC en `TEXT`
- Enums : `TEXT` + `CHECK` constraint
- `PRAGMA foreign_keys=ON` à chaque connexion
- Pool SQLite : `REVUES_DB_MAX_OPEN_CONNS` (défaut 10) — voir benchmarks `internal/store/concurrency_bench_test.go`

## Routes HTTP

```
GET  /healthz
GET  /login
GET  /auth/github/callback
POST /logout

GET|POST /subjects/...
GET /subjects/{id}/modeles
POST /subjects/{id}/revues
GET|PATCH /runs/{id}/items/{id}

GET|POST /admin/users
GET|POST /admin/settings/smtp
GET|POST /admin/integrations/...
```

- Admin sous `/admin/` — `RequireRole("admin")`
- IDs dans URL : valider existence **et** permission

## Templates HTML

- Layout : `web/templates/layouts/base.html`
- Fragments HTMX : suffixe `_fragment.html`
- Partials : `web/templates/partials/`
- Échappement auto `html/template` — pas de `template.HTML` sur input utilisateur

## HTMX

```html
<form hx-post="/runs/1/items/2" hx-target="#item-2" hx-swap="outerHTML">
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
</form>
```

- CSRF via champ hidden **et** `hx-headers` pour requêtes sans form
- Réponses : fragment HTML partiel, pas JSON API

## RBAC

Voir [RBAC.md](./RBAC.md). Règle : **deny by default**.

- `403` = pas le droit
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

## Webhooks (vague 2)

- Signature : `X-Revues-Signature: sha256=<hmac>`
- Payload : `event_id` UUID stable pour idempotence
- Anti-SSRF :
  - Refuser IP privées, loopback, link-local, metadata
  - Timeout 5s, max 1 redirect
  - Pas de schémas autres que `https://` (sauf `http://localhost` en dev)

## Uploads (vague 3)

- Vérifier magic bytes, pas seulement extension
- Nom stockage : UUID, pas le nom original
- `Content-Disposition: attachment` sur téléchargement
- Auth + contrôle projet sur chaque GET

## Commits

```
type(scope): description courte

Types : feat, fix, docs, test, chore, refactor
Scope : auth, projects, runs, admin, integrations, infra
```

Exemple : `feat(auth): GitHub OAuth callback with PKCE (#7)`

## Interdits

- SPA frameworks (React, Vue, Svelte, Angular)
- Bundlers frontend (webpack, vite)
- Redis, Elasticsearch, Kafka
- JWT en cookie (sessions ID en base)
- SQL dans les handlers
- `panic()` en production (sauf main)
