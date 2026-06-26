# Instructions pour agents Cloud — Revues

Contrat d'exécution pour tout agent qui implémente une issue GitHub.

## Mission

Implémenter **une seule** issue. Lire le contexte, respecter le périmètre, livrer une PR mergeable.

## Ordre de lecture (obligatoire)

1. Ce fichier (`AGENTS.md`)
2. Issue GitHub assignée (critères d'acceptation)
3. [docs/PLAN.md](docs/PLAN.md) — vision
4. [docs/CONVENTIONS.md](docs/CONVENTIONS.md) — code et structure
5. [docs/DEFINITION_OF_DONE.md](docs/DEFINITION_OF_DONE.md) — critères merge
6. [docs/RBAC.md](docs/RBAC.md) — si route ou permission touchée
7. [docs/schema/canonical.sql](docs/schema/canonical.sql) — si données touchées
8. [docs/REVIEW_ADVERSE.md](docs/REVIEW_ADVERSE.md) — pièges connus

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

- Go + chi + `html/template` + HTMX
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
./scripts/check.sh   # doit passer
go test ./...        # vert
```

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
Lis AGENTS.md, docs/CONVENTIONS.md, docs/DEFINITION_OF_DONE.md.
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
