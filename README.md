# Revues

Application de gestion de check-lists pour revues de projets — simple d'utilisation, éco-conçue, riche fonctionnellement.

## Docs

- [AGENTS.md](AGENTS.md) — contrat agents · `./scripts/check.sh`
- [Onboarding](docs/ONBOARDING.md) · [Plan](docs/PLAN.md) · [Roadmap](docs/ROADMAP.md)
- [GO.md](docs/GO.md) · [RBAC.md](docs/RBAC.md) · [REVIEW_ADVERSE.md](docs/REVIEW_ADVERSE.md)
- [Déploiement](deploy/README.md) · [Issues](https://github.com/jeb-maker/revues/issues)

## Démarrage

```bash
go run ./cmd/revues   # :8080 — migrations goose au boot
curl -sf http://localhost:8080/healthz   # → ok
```

Variables : [.env.example](.env.example) (pas de chargement auto de `.env`).

## Stack

Go · SQLite · HTML + HTMX · GitHub OAuth · SMTP · Jira / webhooks / Notion
