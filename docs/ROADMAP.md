# Roadmap — Revues

État courant et reste éventuel. Harness : [AGENTS.md](../AGENTS.md) · `./scripts/check.sh`.

## Rewrite API-first + SvelteKit

Décision : [ADR-001](./ADR-001-api-first-svelte.md).  
Orchestration : [rewrite/README.md](./rewrite/README.md) · [WORK_PACKAGES.md](./rewrite/WORK_PACKAGES.md) · [ISSUE_MAP.md](./rewrite/ISSUE_MAP.md).

Big bang : UI `html/template` + HTMX remplacée par **Go OpenAPI + SvelteKit + mb**.

| WP | Issue | Statut |
|----|-------|--------|
| WP-001–005, 010–013 | #244–#252 | Mergé sur `main` |
| WP-014 Mes tâches | #253 | PR [#271](https://github.com/jeb-maker/revues/pull/271) |
| WP-015 Admin org | #254 | PR [#272](https://github.com/jeb-maker/revues/pull/272) |
| WP-016 SMTP / hub / attachments | #255 | PR [#273](https://github.com/jeb-maker/revues/pull/273) (stack #272) |
| WP-020 Jira | #256 | En cours (vague 3) |
| WP-021 Notion | #257 | En cours (vague 3) |
| WP-022 Webhooks | #258 | En cours (vague 3) |
| WP-030 Clôture | #259 | PR de clôture docs/check |

## Livré (pré-rewrite — référence métier)

Cœur métier, auth GitHub, RBAC, orgs/équipes, SMTP, Jira Cloud, webhooks, Notion, pièces jointes — comportement de référence pour les WP sur l’ancienne stack HTML/HTMX.

## Ouvert hors rewrite

| Issue | Notes |
|-------|-------|
| [#65](https://github.com/jeb-maker/revues/issues/65) | Jira Server/DC — **icebox** (pas sans demande produit) |

## Icebox (pas d’issue tant que signal d’usage)

Séries/campagnes, fusion sujets, rapport org, Slack/Teams, Google OAuth, gouvernance avancée, audit admin, concurrency items, antivirus uploads, PostgreSQL, CSP stricte, rotation clés.

Voir aussi [REVIEW_ADVERSE.md](./REVIEW_ADVERSE.md) § CAN DEFER.
