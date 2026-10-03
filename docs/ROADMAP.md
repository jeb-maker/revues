# Roadmap — Revues

État courant et reste éventuel. Harness : [AGENTS.md](../AGENTS.md) · `./scripts/check.sh`.

## Rewrite API-first + SvelteKit

Décision : [ADR-001](./ADR-001-api-first-svelte.md).  
Orchestration : [rewrite/README.md](./rewrite/README.md) · [WORK_PACKAGES.md](./rewrite/WORK_PACKAGES.md) · [ISSUE_MAP.md](./rewrite/ISSUE_MAP.md).

Big bang : UI `html/template` + HTMX remplacée par **Go OpenAPI + SvelteKit + mb**.

| WP | Issue | Statut |
|----|-------|--------|
| WP-001–005, 010–016, 020–022 | #244–#258 | Mergé sur `main` |
| WP-030 Clôture | #259 | Cette PR — garde HTMX + docs |

## Livré (pré-rewrite — référence métier)

Cœur métier, auth GitHub, RBAC, orgs/équipes, SMTP, Jira Cloud, webhooks, Notion, pièces jointes — comportement de référence pour les WP sur l’ancienne stack HTML/HTMX.

## Ouvert hors rewrite

| Issue | Notes |
|-------|-------|
| [#65](https://github.com/jeb-maker/revues/issues/65) | Jira Server/DC — **icebox** (pas sans demande produit) |

## Icebox (pas d’issue tant que signal d’usage)

Séries/campagnes, fusion sujets, rapport org, Slack/Teams, Google OAuth, gouvernance avancée, audit admin, progressive disclosure SPA (SimpleUI / ShowSubject*), écran admin libellés UI, antivirus uploads, PostgreSQL, CSP stricte, rotation clés.

Livré hors icebox : verrou optimiste `run_items.updated_at` (#281).

Voir aussi [REVIEW_ADVERSE.md](./REVIEW_ADVERSE.md) § CAN DEFER.
