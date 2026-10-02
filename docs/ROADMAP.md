# Roadmap — Revues

État courant et reste éventuel. Harness : [AGENTS.md](../AGENTS.md) · `./scripts/check.sh`.

## En cours — Rewrite API-first + SvelteKit

Décision : [ADR-001](./ADR-001-api-first-svelte.md).  
Orchestration agents : [rewrite/README.md](./rewrite/README.md) · [WORK_PACKAGES.md](./rewrite/WORK_PACKAGES.md).

Big bang : remplacement de l’UI `html/template` + HTMX par **Go OpenAPI + SvelteKit + mb**.

Créer les issues : `./scripts/create-rewrite-issues.sh`.

## Livré (pré-rewrite)

Cœur métier, auth GitHub, RBAC, orgs/équipes, SimpleUI, preuve ZIP, SMTP, Jira Cloud, webhooks, Notion, pièces jointes, hardening sécu — sur l’ancienne stack HTML/HTMX (référence comportementale pour les WP).

## Ouvert hors rewrite

| Issue | Notes |
|-------|-------|
| [#65](https://github.com/jeb-maker/revues/issues/65) | Jira Server/DC — **icebox** (pas sans demande produit) |

## Icebox (pas d’issue tant que signal d’usage)

Séries/campagnes, fusion sujets, rapport org, Slack/Teams, Google OAuth, gouvernance avancée, audit admin, concurrency items, antivirus uploads, PostgreSQL, CSP stricte, rotation clés.

Voir aussi [REVIEW_ADVERSE.md](./REVIEW_ADVERSE.md) § CAN DEFER.
