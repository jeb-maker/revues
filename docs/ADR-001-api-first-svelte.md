# ADR-001 — API-first Go + SvelteKit + miniature-broccoli

Statut : **accepté**  
Date : 2026-10-02

## Contexte

Revues est un monolithe Go + `html/template` + HTMX. La plomberie HTTP/HTML et l’absence de contrat API freinent l’évolution. Objectif : **minimiser le code non fonctionnel** (framework ou génération) et remplacer le front templates par un framework JS léger.

## Décision

| Couche | Choix | Non-retenu |
|--------|--------|------------|
| API | Go + chi + **OpenAPI** (`ogen` ou `oapi-codegen`) | Quarkus |
| Données | **sqlc** + goose + SQLite (SQL explicite) | ORM / Panache |
| Front | **SvelteKit** (SPA, adapter static) | Vue, vanilla seul, HTMX |
| UI kit | **@jeb-maker/mb** (Web Components Lit) | réécrire les composants |
| Auth | Sessions cookie + CSRF (header), pas de JWT cookie | — |
| Migration | **Big bang** — pas de coexistence HTML/API | dual-stack prolongé |

## Conséquences

- Les handlers HTML/HTMX et `web/templates/` sont **supprimés** (vague fondation).
- Le contrat public HTTP métier est **`/api/v1`** décrit par OpenAPI.
- Le client TypeScript front est **généré** depuis OpenAPI.
- Les règles RBAC, le schéma et les intégrations restent la source métier ; seuls les adapteurs changent.
- Budgets éco HTML/HTMX historiques sont **remplacés** (voir PLAN.md) : mesurer bundle SvelteKit app vs vendor mb.
- Docs agents (`AGENTS.md`, `CONVENTIONS.md`, DoD) imposent la nouvelle stack.

## Références

- Orchestration agents : [rewrite/README.md](./rewrite/README.md)
- Lots : [rewrite/WORK_PACKAGES.md](./rewrite/WORK_PACKAGES.md)
- API : [API.md](./API.md) · Front : [FRONTEND.md](./FRONTEND.md)
