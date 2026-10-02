# Instructions pour agents Cloud — Revues

Contrat d'exécution pour tout agent qui implémente une issue GitHub.

## Mission

Implémenter **une seule** issue. Lire le contexte, respecter le périmètre, livrer une PR mergeable.

## Ordre de lecture (obligatoire)

1. Ce fichier (`AGENTS.md`)
2. Issue GitHub assignée (critères d'acceptation)
3. [docs/ADR-001-api-first-svelte.md](docs/ADR-001-api-first-svelte.md) — stack cible
4. [docs/PLAN.md](docs/PLAN.md) — vision
5. [docs/CONVENTIONS.md](docs/CONVENTIONS.md) — code et structure
6. [docs/API.md](docs/API.md) — si surface HTTP `/api/v1`
7. [docs/FRONTEND.md](docs/FRONTEND.md) — si SvelteKit / mb
8. [docs/GO.md](docs/GO.md) — **bonnes pratiques Go obligatoires**
9. [docs/DEFINITION_OF_DONE.md](docs/DEFINITION_OF_DONE.md) — critères merge
10. [docs/RBAC.md](docs/RBAC.md) — si route ou permission touchée
11. [docs/schema/canonical.sql](docs/schema/canonical.sql) — si données touchées
12. [docs/REVIEW_ADVERSE.md](docs/REVIEW_ADVERSE.md) — pièges connus
13. Rewrite : [docs/rewrite/README.md](docs/rewrite/README.md) + WP dans [docs/rewrite/WORK_PACKAGES.md](docs/rewrite/WORK_PACKAGES.md)

## Règles strictes

### Périmètre

- **1 issue = 1 PR** — jamais de PR fourre-tout
- **Interdit** : refactor hors scope, feature adjacente, « tant qu'on y est »
- **Hors scope** : lister explicitement dans la PR
- Si un critère est ambigu → commenter sur l'issue, **ne pas deviner**

### Branche et PR

```
Branche : cursor/issue-<N>-<slug>-f21b
Titre PR : identique au titre de l'issue
Corps PR : Closes #<N>
```

### Stack imposée

- **API** : Go + chi + OpenAPI (codegen) — [docs/API.md](docs/API.md)
- **Données** : SQLite WAL + goose + **sqlc** (pas d'ORM) — [docs/GO.md](docs/GO.md)
- **Front** : SvelteKit (SPA) + **@jeb-maker/mb** — [docs/FRONTEND.md](docs/FRONTEND.md)
- Client TS **généré** depuis OpenAPI
- Pas de `html/template` / HTMX pour l'UI métier (rewrite big bang)
- Pas de polling ni WebSocket
- Appels API externes à la demande uniquement

### Sécurité (non négociable)

- RBAC **côté serveur** sur chaque route sensible
- Contrôle **IDOR** : vérifier appartenance sujet/revue org active
- CSRF sur **toutes** les mutations API (`X-CSRF-Token`)
- Secrets en variables d'environnement, credentials chiffrés en base
- Email GitHub **vérifié** avant whitelist
- Voir [docs/RBAC.md](docs/RBAC.md) pour la matrice

### Données

- Schéma normatif : [docs/schema/canonical.sql](docs/schema/canonical.sql)
- Migrations goose = source d'exécution
- Ne modifier `canonical.sql` que via issue `area:data` dédiée
- Snapshot revue = copie SQL transactionnelle
- Jamais d'UPDATE destructif sur une version de modèle publiée

### Éco-contraintes

Budgets SPA — détail et seuils dans [docs/PLAN.md](docs/PLAN.md) (mesurés par `check.sh`).

| Zone | Règle |
|------|--------|
| JS/CSS **app** | seuils fail dans `check.sh` |
| Vendor mb / reports | mesurés ; hors ou seuil dédié documenté |
| Requêtes API par navigation | viser sobriété (pas de chatty loops) |

### Tests minimum

```bash
./scripts/check.sh   # Go + front (gofmt, vet, test -race, golangci-lint, build front)
go test ./...        # vert
```

- Suivre [docs/GO.md](docs/GO.md) : context, erreurs wrappées, SQL dans store/sqlc, tests table-driven
- Ajouter un test si logique métier ou middleware RBAC touché
- Issues `area:auth` ou `area:integrations` : tests sécurité requis (voir DoD)

## Avant de pousser

```bash
./scripts/check.sh
git add -A && git commit -m "feat(scope): description (Closes #N)"
git push -u origin cursor/issue-<N>-<slug>-f21b
```

## Prompt type (pour lancer un agent)

```
Repo jeb-maker/revues. Implémente UNIQUEMENT l'issue #N.
Lis AGENTS.md, docs/ADR-001-api-first-svelte.md, docs/CONVENTIONS.md,
docs/API.md, docs/FRONTEND.md, docs/GO.md, docs/DEFINITION_OF_DONE.md.
Si rewrite : docs/rewrite/WORK_PACKAGES.md. Si RBAC : docs/RBAC.md.
Si données : docs/schema/canonical.sql.
Branche cursor/issue-N-<slug>-f21b. PR avec Closes #N.
./scripts/check.sh doit passer avant push.
```

## Issues à revue humaine obligatoire

- Auth GitHub OAuth
- Toute issue `area:integrations`
- Toute issue touchant chiffrement ou webhooks

## Fichiers sensibles (issue dédiée préférable)

- `docs/schema/canonical.sql`
- `docs/RBAC.md`
- `AGENTS.md`
- `.github/workflows/ci.yml`
- `api/openapi/openapi.yaml` (conflits multi-agents : coordonner par tags/domaines)

## Cursor Cloud specific instructions

Contexte durable pour les agents Cloud (l'update script a déjà installé les dépendances).

- **Stack/run** : API Go 1.22 (pas de CGO — driver `modernc.org/sqlite` pur Go) + front SvelteKit. Lancer l'API : `go run ./cmd/revues` (écoute `:8080`, migrations goose au démarrage). Front dev : voir [docs/FRONTEND.md](docs/FRONTEND.md) (proxy vers l'API). Variables : `.env.example` (le binaire lit `os.Getenv`, **pas de chargement automatique de `.env`**).
- **Gatekeeper** : `./scripts/check.sh` — Go (gofmt, vet, test -race, build, `go mod tidy` strict, golangci-lint) + front (`npm` check/build). `golangci-lint` v1.62 sur le `PATH` (`$(go env GOPATH)/bin`).
- **Auth / démo locale** :
  - Endpoints JSON : `GET /api/v1/bootstrap` (user + `csrf_token`, pose cookie guest si anonymes), `GET /api/v1/me`, `POST /api/v1/auth/login|register|logout`. OAuth browser : `GET /auth/github/start` + `/auth/github/callback` → redirect SPA `/login` ou `/`.
  - Pages Svelte : `/login`, `/register` (proxy Vite `/api` `/auth` `/static` → `:8080`).
  - Sans `REVUES_GITHUB_CLIENT_ID`/`SECRET`, le bouton GitHub est masqué (`github_oauth_enabled: false`) ; utiliser login/register local.
  - **Seed session** (sans UI) : `store.UpsertGitHubUser` ou `CreateLocalUser`, puis `store.CreateSession` avec le hash de `auth.RandomToken`, cookie HttpOnly `revues_session=<raw>`. CSRF = HMAC(`session token + REVUES_SESSION_SECRET`) — aussi renvoyé par bootstrap/`me` ; envoyer `X-CSRF-Token` sur chaque mutation API.
  - `REVUES_BOOTSTRAP_ADMIN_EMAIL` : rôle admin au premier login de cet email. `REVUES_DEV_AUTH=1` (hors production, loopback) auto-session + `POST /auth/dev/login` switch user.
  - Whitelist : `REVUES_LOGIN_REQUIRE_WHITELIST=1` refuse register/OAuth hors `allowed_emails` (tests sécurité).
- **SQLite** : pool `REVUES_DB_MAX_OPEN_CONNS` (défaut 10) + WAL + `busy_timeout`, base `data/revues.db` (gitignored).
- **Reset base dev** : `./scripts/reset-db.sh` ; `--seed` disponible. Migrations goose dans `migrations/`.
- **Rewrite** : orchestration [docs/rewrite/README.md](docs/rewrite/README.md). Créer les issues : `./scripts/create-rewrite-issues.sh`.
