# Checklist revue PR

## Issue & périmètre

- [ ] 1 PR = 1 issue GitHub
- [ ] `Closes #N` dans la description
- [ ] Critères d'acceptation de l'issue tous cochés
- [ ] Section « Hors scope » remplie
- [ ] Pas de scope creep (features non demandées par l'issue)

## CI & qualité

- [ ] `./scripts/check.sh` passe
- [ ] `go test ./...` vert
- [ ] `gofmt` appliqué
- [ ] Pas de `TODO` sans `TODO(#N)`

## Sécurité

- [ ] RBAC serveur sur chaque nouvelle route (voir [docs/RBAC.md](../docs/RBAC.md))
- [ ] Contrôle IDOR projet/revue/point
- [ ] CSRF sur tous les POST (y compris HTMX)
- [ ] Pas de secret en clair (code, logs, commit)
- [ ] Templates : échappement HTML

## Données

- [ ] Migration goose si schéma modifié
- [ ] Aligné avec [docs/schema/canonical.sql](../docs/schema/canonical.sql)
- [ ] Transaction atomique sur update + audit
- [ ] Pas d'UPDATE destructif sur versions publiées

## Métier

- [ ] Commentaire obligatoire si `nok` (validation serveur)
- [ ] Snapshot revue = copie SQL transactionnelle
- [ ] `due_date` respecté si issue concerne les revues

## Éco

- [ ] Pas de SPA / React / Vue / Vite
- [ ] HTMX ciblé (fragments)
- [ ] Pas de polling / WebSocket

## Revue humaine obligatoire ?

- [ ] Issue `area:auth` → revue humaine requise
- [ ] Issue `area:integrations` → revue humaine requise
- [ ] Sinon → auto-merge possible si CI verte
