# Revues

Application de gestion de check-lists pour revues de projets — simple d'utilisation, éco-conçue, riche fonctionnellement.

## Documentation

- [Plan produit & technique](docs/PLAN.md)
- [Roadmap & tâches déléguables](docs/ROADMAP.md)
- [Guide de délégation GitHub](docs/DELEGATION.md)
- [Revue adverse (juin 2026)](docs/REVIEW_ADVERSE.md)
- [Instructions agents Cloud](AGENTS.md)
- [Bonnes pratiques Go](docs/GO.md)
- [Issues GitHub](https://github.com/jeb-maker/revues/issues)

## Harness agents

Avant toute délégation, lire [AGENTS.md](AGENTS.md) et exécuter :

```bash
./scripts/check.sh
```

## Démarrage

```bash
go mod tidy
go run ./cmd/revues
```

Le serveur écoute sur `:8080` par défaut (`REVUES_ADDR`).

Vérifications :

```bash
curl http://localhost:8080/healthz   # → ok
curl -I http://localhost:8080/       # → page HTML d'accueil
open http://localhost:8080/login     # → connexion GitHub OAuth
```

Variables d'environnement : voir [.env.example](.env.example) (`REVUES_DATABASE_PATH`, `REVUES_GITHUB_CLIENT_*`, `REVUES_SESSION_SECRET`, `REVUES_BOOTSTRAP_ADMIN_EMAIL`).

Au démarrage, les migrations goose s'appliquent automatiquement.

## Structure (bootstrap)

```
cmd/revues/           # point d'entrée, wiring serveur
internal/store/       # connexion SQLite, migrations goose
internal/web/         # router chi, handlers HTTP
internal/config/      # configuration REVUES_*
migrations/           # SQL goose (source d'exécution)
web/static/           # CSS, JS (servi sur /static/)
web/templates/        # html/template (layout + pages)
data/                 # SQLite (gitignored)
```

## Principes

- **Revues** exécute et trace les revues
- **Jira** traite les points `nok`
- **Webhooks** notifient la stack
- **Notion** archive et documente

## Stack (cible)

Go · SQLite · HTML + HTMX · GitHub OAuth · SMTP admin · intégrations Jira / webhooks / Notion
