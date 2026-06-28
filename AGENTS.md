# Instructions pour agents Cloud — Revues

Contrat d'exécution pour tout agent qui implémente une issue GitHub.

## Mission

Implémenter **une seule** issue. Lire le contexte, respecter le périmètre, livrer une PR mergeable.

## Ordre de lecture (obligatoire)

1. Ce fichier (`AGENTS.md`)
2. Issue GitHub assignée (critères d'acceptation)
3. [docs/PLAN.md](docs/PLAN.md) — vision
4. [docs/CONVENTIONS.md](docs/CONVENTIONS.md) — code et structure
5. [docs/GO.md](docs/GO.md) — **bonnes pratiques Go obligatoires**
6. [docs/DEFINITION_OF_DONE.md](docs/DEFINITION_OF_DONE.md) — critères merge
7. [docs/RBAC.md](docs/RBAC.md) — si route ou permission touchée
8. [docs/schema/canonical.sql](docs/schema/canonical.sql) — si données touchées
9. [docs/REVIEW_ADVERSE.md](docs/REVIEW_ADVERSE.md) — pièges connus

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

- Go + chi + `html/template` + HTMX — voir [docs/GO.md](docs/GO.md)
- SQLite WAL + goose
- **Pas de SPA** (React, Vue, Vite, webpack)
- Pas de polling ni WebSocket
- Appels API externes à la demande uniquement

### Sécurité (non négociable)

- RBAC **côté serveur** sur chaque route sensible
- Contrôle **IDOR** : vérifier appartenance projet/revue
- CSRF sur **tous** les POST (y compris HTMX via `hx-headers`)
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

| Métrique | Max |
|----------|-----|
| HTML page | 50 Ko |
| CSS | 20 Ko |
| JS / HTMX | 15 Ko |
| Requêtes / page | 8 |

### Tests minimum

```bash
./scripts/check.sh   # doit passer (gofmt, vet, test -race, golangci-lint)
go test ./...        # vert
```

- Suivre [docs/GO.md](docs/GO.md) : context, erreurs wrappées, SQL dans `store`, tests table-driven
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
Lis AGENTS.md, docs/CONVENTIONS.md, docs/GO.md, docs/DEFINITION_OF_DONE.md.
Si RBAC : docs/RBAC.md. Si données : docs/schema/canonical.sql.
Branche cursor/issue-N-<slug>-f21b. PR avec Closes #N.
./scripts/check.sh doit passer avant push.
```

## Issues à revue humaine obligatoire

- #7 Auth GitHub OAuth
- Toute issue `area:integrations`
- Toute issue touchant chiffrement ou webhooks

## Fichiers sacrés (ne pas modifier sans issue dédiée)

- `docs/schema/canonical.sql`
- `docs/RBAC.md`
- `AGENTS.md`
- `.github/workflows/ci.yml`

## Cursor Cloud specific instructions

Contexte durable pour les agents Cloud (l'update script a déjà installé les dépendances).

- **Stack/run** : app Go 1.22 (pas de CGO — driver `modernc.org/sqlite` pur Go). Lancer en dev : `go run ./cmd/revues` (écoute `:8080`, migrations goose appliquées au démarrage). Variables : `.env.example` (le binaire lit `os.Getenv`, **il n'y a pas de chargement automatique de `.env`** — exporter les variables soi-même).
- **Gatekeeper** : `./scripts/check.sh` lance gofmt, `go vet`, `go test -race`, build, **`go mod tidy` puis échoue si `go.mod`/`go.sum` changent**, et `golangci-lint`. `golangci-lint` (v1.62, comme la CI) doit être sur le `PATH` ; il est installé dans `$(go env GOPATH)/bin`, déjà présent dans le `PATH` du login shell.
- **Auth / démo locale** : toutes les pages métier sont derrière l'OAuth GitHub. Sans `REVUES_GITHUB_CLIENT_ID`/`REVUES_GITHUB_CLIENT_SECRET`, `/auth/github/start` redirige vers `/login?error=...` et on ne peut pas se connecter via l'UI. Pour exercer les fonctionnalités authentifiées en local sans OAuth : créer un user + session directement en base (réutiliser `store.UpsertGitHubUser` puis `store.CreateSession` avec le hash de `auth.RandomToken`), puis poser le cookie `revues_session`. Le jeton CSRF est dérivé de `session token + REVUES_SESSION_SECRET` et rendu dans les pages (`<meta name="csrf-token">` et champ caché `csrf_token`). `REVUES_BOOTSTRAP_ADMIN_EMAIL` donne le rôle admin au premier login de cet email.
- **SQLite** : `MaxOpenConns(1)` + WAL, base `data/revues.db` (gitignored). Accès concurrent (serveur + script de seed) OK grâce au WAL.
