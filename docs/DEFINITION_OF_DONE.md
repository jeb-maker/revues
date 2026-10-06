# Définition of Done — Revues

Critères pour qu'une issue soit considérée **terminée** et mergeable.

## Par issue

- [ ] Tous les critères d'acceptation de l'issue sont remplis
- [ ] `./scripts/check.sh` passe en local
- [ ] CI GitHub Actions verte
- [ ] PR avec `Closes #N` et checklist complétée
- [ ] Section « Hors scope » remplie dans la PR
- [ ] Issue fermée au merge

## Code

- [ ] Compile sans warning (`go vet` clean)
- [ ] `golangci-lint run` vert (CI)
- [ ] `go test ./...` vert (`-race` via check.sh)
- [ ] Front : `npm run check` / build vert si fichiers `frontend/` touchés
- [ ] OpenAPI régénéré si spec modifiée (serveur + client TS)
- [ ] Conforme [GO.md](./GO.md) (context, erreurs, SQL store/sqlc)
- [ ] Pas de `TODO` sans référence issue (`TODO(#N)`)
- [ ] Pas de code commenté mort
- [ ] `gofmt` appliqué

## Sécurité

- [ ] Chaque nouvelle route documentée dans la matrice RBAC de la PR
- [ ] Mutations API protégées CSRF (`X-CSRF-Token`)
- [ ] Pas de secret en clair (code, logs, commits)
- [ ] Contrôle IDOR sur ressources projet/revue/point
- [ ] Pas de HTML métier servi par templates Go (SPA + JSON API)

## Données

- [ ] Migration goose si schéma modifié (+ `goose down` testé)
- [ ] `PRAGMA foreign_keys=ON` et WAL activés
- [ ] Index sur colonnes de jointure / filtre
- [ ] Transactions atomiques sur update + audit event

## UI (si applicable)

- [ ] SvelteKit + mb ([FRONTEND.md](./FRONTEND.md))
- [ ] Client API généré — pas de fetch ad hoc dupliquant le contrat
- [ ] Respect budgets éco SPA (`frontend/spa-budget.json` / check.sh) — pas de bump de plafond dans cette PR sauf `chore(budget)` dédiée

## Tests

| Area | Minimum |
|------|---------|
| `area:auth` | Test RBAC + CSRF + session |
| `area:core` | Test logique métier (snapshot, statuts, nok) |
| `area:integrations` | Test HMAC webhook, mock API, anti-SSRF |
| `area:data` | Test migration up/down |
| `area:ui` | Build front + tests API des endpoints exposés |

## Documentation

- [ ] `.env.example` mis à jour si nouvelle variable
- [ ] README mis à jour si nouvelle commande
- [ ] Pas de doc hors scope de l'issue

## Hors DoD (ne pas bloquer le merge)

- Perf tuning poussé
- Refactoring global
- i18n / traductions
- PostgreSQL
- Couverture 100 %
- Accessibilité WCAG complète (souhaitable, pas bloquant v1)
