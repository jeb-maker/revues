# Revues

Application de gestion de check-lists pour revues de projets — simple d'utilisation, éco-conçue, riche fonctionnellement.

## Docs

- [AGENTS.md](AGENTS.md) — contrat agents · `./scripts/check.sh`
- [ADR-001](docs/ADR-001-api-first-svelte.md) — stack API Go + SvelteKit + mb
- [Rewrite agents](docs/rewrite/README.md) · [Work packages](docs/rewrite/WORK_PACKAGES.md)
- [Onboarding](docs/ONBOARDING.md) · [Plan](docs/PLAN.md) · [Roadmap](docs/ROADMAP.md)
- [API.md](docs/API.md) · [FRONTEND.md](docs/FRONTEND.md) · [GO.md](docs/GO.md) · [RBAC.md](docs/RBAC.md)

## Démarrage

```bash
# API Go (:8080) — migrations goose au boot
go run ./cmd/revues
curl -sf http://localhost:8080/healthz   # → ok

# Front SvelteKit (dev) — proxy vers l'API
cd frontend && npm ci && npm run dev

# Ou build SPA servi par Go
cd frontend && npm ci && npm run build
go run ./cmd/revues   # sert frontend/build (sinon stub HTML documenté)
```

Variables : [.env.example](.env.example) (pas de chargement auto de `.env`).  
Optionnel : `REVUES_SPA_DIR` pour pointer vers un autre dossier de build SPA.

## Stack

Go · OpenAPI · sqlc · SQLite · SvelteKit · miniature-broccoli · GitHub OAuth · SMTP · Jira / webhooks / Notion

## Arborescence (cible)

```
api/openapi/     # contrat OpenAPI (WP-002+)
frontend/        # SvelteKit SPA
internal/        # métier, store, auth, integrations
web/static/      # vendor mb (intégration front = WP-005)
```

## Délégation rewrite

```bash
./scripts/create-rewrite-issues.sh
```
