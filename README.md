# Revues

Application de gestion de check-lists pour revues de projets — simple d'utilisation, éco-conçue, riche fonctionnellement.

## Documentation

- [Plan produit & technique](docs/PLAN.md)
- [Roadmap & tâches déléguables](docs/ROADMAP.md)
- [Guide de délégation GitHub](docs/DELEGATION.md)
- [Revue adverse (juin 2026)](docs/REVIEW_ADVERSE.md)
- [Instructions agents Cloud](AGENTS.md)
- [Issues GitHub](https://github.com/jeb-maker/revues/issues)

## Harness agents

Avant toute délégation, lire [AGENTS.md](AGENTS.md) et exécuter :

```bash
./scripts/check.sh
```

## Démarrage (squelette)

```bash
go mod tidy
go run ./cmd/revues
curl http://localhost:8080/healthz
```

Variables : voir [.env.example](.env.example).

## Principes

- **Revues** exécute et trace les revues
- **Jira** traite les points `nok`
- **Webhooks** notifient la stack
- **Notion** archive et documente

## Stack (cible)

Go · SQLite · HTML + HTMX · GitHub OAuth · SMTP admin · intégrations Jira / webhooks / Notion
