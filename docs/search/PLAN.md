# Recherche globale — plan

Remplacer les champs « Rechercher… » **par page** (sujets, modèles, revues, mes tâches) par une **recherche unique dans le shell**, qui interroge plusieurs entités sous le même RBAC que les listes.

## Problème

Aujourd’hui chaque liste a son propre `q` (`LIKE` via `searchTerms`) :

| Surface | Endpoint | Champs matchés |
|---------|----------|----------------|
| Sujets | `GET /subjects?q=` | nom, description |
| Modèles | `GET /templates?q=` | nom, domaines |
| Revues | `GET /runs?q=` | sujet, modèle, login |
| Mes tâches | `GET /me/tasks?q=` | sujet, modèle, label, section |

L’utilisateur doit savoir *où* chercher. Les filtres métier (statut, échéance) restent utiles **par page** ; le texte libre est transversal.

## Cible UX

1. **Barre de recherche dans le header** (auth only, hors login/register).
2. Soumission → page **`/search?q=`** avec résultats groupés par type (sujets, revues, modèles, tâches).
3. Clic → fiche / détail existant (`/subjects/{id}`, `/runs/{id}`, …).
4. Retrait des inputs texte des listes ; **conserver** les filtres non-texte (`status`, etc.).
   Les `?q=` legacy sur les listes SPA redirigent en soft vers `/search?q=` (`listQueryRedirect`).
5. Pas de typeahead / debounce chatty en v1 (budget API : 1 requête par soumission).

## Cible technique

- **Un endpoint** `GET /api/v1/search?q=&limit=` (OpenAPI, tag `search`).
- Réponse discriminée : `{ results: [{ kind, id, title, subtitle?, href_hint }], total_by_kind? }`.
- Visibilité **identique** aux listes : org active + `ResolveSubjectAccess` / règles `show_my_tasks` / `show_modeles`.
- v1 : réutiliser `searchTerms` + `LIKE` (pas d’Elasticsearch — [CONVENTIONS.md](../CONVENTIONS.md)).
- v1.1 optionnelle : SQLite **FTS5** (`area:data`) si volume / pertinence le justifient.

## Hors scope (icebox)

- Recherche dans le contenu des commentaires / pièces jointes.
- Fuzzy / typos / ranking ML.
- Raccourci clavier global type Spotlight (peut suivre une fois la page `/search` stable).
- Admin (users, whitelist) — hors org métier.

## Découpage

Voir [WORK_PACKAGES.md](./WORK_PACKAGES.md). Ordre : API → shell SPA → retrait `q` listes → (option) FTS5.

## Suivi

| WP | Issue |
|----|-------|
| Epic | [#287](https://github.com/jeb-maker/revues/issues/287) |
| API | [#288](https://github.com/jeb-maker/revues/issues/288) |
| Shell | [#289](https://github.com/jeb-maker/revues/issues/289) |
| Retrait q | [#290](https://github.com/jeb-maker/revues/issues/290) |
| FTS5 | [#291](https://github.com/jeb-maker/revues/issues/291) |

Mapping : [ISSUE_MAP.md](./ISSUE_MAP.md). Roadmap : [ROADMAP.md](../ROADMAP.md).
