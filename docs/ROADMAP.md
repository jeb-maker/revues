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

Cœur métier, auth GitHub, RBAC, orgs/équipes, SMTP, Jira Cloud, webhooks, pièces jointes — comportement de référence pour les WP sur l’ancienne stack HTML/HTMX.

## Recherche fulltext globale (shell)

Plan : [search/PLAN.md](./search/PLAN.md) · WP : [search/WORK_PACKAGES.md](./search/WORK_PACKAGES.md) · mapping : [search/ISSUE_MAP.md](./search/ISSUE_MAP.md).

| WP | Issue | Notes |
|----|-------|-------|
| SEARCH-001 | [#288](https://github.com/jeb-maker/revues/issues/288) | API `GET /api/v1/search` |
| SEARCH-002 | [#289](https://github.com/jeb-maker/revues/issues/289) | Shell + page `/search` |
| SEARCH-003 | [#290](https://github.com/jeb-maker/revues/issues/290) | Retrait `q` listes |
| SEARCH-004 | [#291](https://github.com/jeb-maker/revues/issues/291) | FTS5 — **icebox** |

## Ouvert hors rewrite

| Issue | Notes |
|-------|-------|
| [#340](https://github.com/jeb-maker/revues/issues/340) | Jira Cloud OAuth 3LO (utilisateur = soi-même) — **cible auth** ; token org devient transitoire |
| [#65](https://github.com/jeb-maker/revues/issues/65) | Jira Server/DC — **icebox** (pas sans demande produit) |

Livré hors rewrite : **retrait Notion** (SPA + OpenAPI + package `internal/integrations/notion`) — hors périmètre produit ; colonnes legacy en base seulement, voir [NOTION.md](./NOTION.md). **Confluence Cloud publish** (#337) — config admin + publication d’une revue `done` (auth token org ; OAuth Atlassian = #340).

## Icebox (pas d’issue tant que signal d’usage)

Séries/campagnes, fusion sujets, rapport org, Slack/Teams, Google OAuth, gouvernance avancée, audit admin, progressive disclosure SPA (SimpleUI / ShowSubject*), écran admin libellés UI, antivirus uploads, PostgreSQL, CSP stricte, rotation clés, **Notion** (réintro seulement sur signal). Confluence Server/DC et sync bidirectionnelle restent hors scope.

Livré hors icebox : verrou optimiste `run_items.updated_at` (#281).

Voir aussi [REVIEW_ADVERSE.md](./REVIEW_ADVERSE.md) § CAN DEFER.
