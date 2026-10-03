# Work packages — recherche globale

Chaque `## SEARCH-xxx` = une issue GitHub. Labels : première ligne `Labels:` · dépendances `Bloqué par:`.

Plan produit : [PLAN.md](./PLAN.md).

---

## SEARCH-000 — Epic : recherche fulltext globale (shell)

Labels: `epic`, `area:ui`, `enhancement`  
Bloqué par: aucune

### Objectif

Une recherche unique dans le header à la place des champs texte par liste.

### Critères d’acceptation

- [ ] Issues filles livrées : SEARCH-001 → SEARCH-003 (SEARCH-004 optionnelle)
- [ ] Doc [PLAN.md](./PLAN.md) + mapping issues à jour
- [ ] ROADMAP pointe vers ce chantier
- [ ] Hors scope : FTS5 (SEARCH-004), typeahead, commentaires/PJ, Elasticsearch

---

## SEARCH-001 — API `GET /api/v1/search`

Labels: `area:core`, `enhancement`  
Bloqué par: aucune

### Objectif

Endpoint unique de recherche multi-entités, auth + org, RBAC aligné sur les listes.

### Critères d’acceptation

- [ ] OpenAPI : `GET /api/v1/search` avec `q` (requis, minLength ≥ 1), `limit` optionnel (plafond serveur, ex. 20–50 total ou par kind)
- [ ] Réponse typée : résultats discriminés `kind` ∈ `subject` | `run` | `template` | `task` + `id`, `title`, `subtitle` optionnel, chemin relatif pour navigation SPA
- [ ] Store : agrège les requêtes existantes (ou SQL dédié) avec `searchTerms` / `likeContainsPattern` ; **pas** d’accès hors `ResolveSubjectAccess` / règles listes
- [ ] `template` omis si `show_modeles` faux ; `task` omis si `show_my_tasks` faux
- [ ] `q` vide / whitespace → `400` clair
- [ ] Tests HTTP + store : match multi-kind, IDOR (sujet autre org → absent), org sans session → 401
- [ ] Matrice RBAC mise à jour dans la PR
- [ ] Codegen OpenAPI serveur + client TS régénérés
- [ ] Hors scope : UI shell, retrait `q` listes, FTS5, ranking avancé

### Notes techniques

- Réutiliser `internal/store/search.go` ; factoriser si besoin sans casser les `?q=` listes (encore utilisés jusqu’à SEARCH-003).
- Une seule round-trip SQL « fan-out » ou 4 requêtes bornées — documenter le choix ; viser sobriété (≤ 4 queries).

---

## SEARCH-002 — Shell SPA : barre header + page `/search`

Labels: `area:ui`, `enhancement`  
Bloqué par: SEARCH-001

### Objectif

Barre de recherche dans le chrome auth ; page résultats groupés.

### Critères d’acceptation

- [ ] Champ recherche dans `+layout.svelte` (auth only ; absent login/register)
- [ ] Soumission → `/search?q=` (replaceState OK) ; `GET /api/v1/search` une fois
- [ ] Page `/search` : groupes par kind avec libellés i18n (`ui_subject_label` / `ui_run_label`), liens vers routes existantes
- [ ] États : vide (invite), loading, erreur API, aucun résultat
- [ ] Accessible : label, focus, pas de piège clavier ; mobile : champ utilisable (pas masqué par nav)
- [ ] Pas de polling / debounce multi-requêtes en v1
- [ ] Hors scope : retirer les `q` des listes (SEARCH-003), typeahead, raccourci ⌘K

### Notes UX

- Placement : zone header (ex. entre nav et compte, ou sous la barre sur mobile) — rester cohérent avec la grille actuelle brand / nav centrée / compte.
- Charte mb : `mb-input` + résultats en listes simples (pas de cards décoratives).

---

## SEARCH-003 — Retrait des recherches texte par page

Labels: `area:ui`, `enhancement`  
Bloqué par: SEARCH-002

### Objectif

Les listes ne dupliquent plus le texte libre ; filtres métier conservés.

### Critères d’acceptation

- [ ] Retirer le champ `q` / « Rechercher… » de : `/subjects`, `/modeles`, `/runs`, `/mes-taches`
- [ ] Conserver filtres non-texte (`status` sur runs et mes-taches, etc.)
- [ ] Param `q` URL listes : ignorer ou rediriger soft vers `/search?q=` (trancher dans la PR, documenter)
- [ ] Endpoints listes `?q=` **conservés** côté API (pas de breaking) — usage déprécié UI only ; pas d’obligation de les supprimer
- [ ] Lede / empty states : pas d’incitation à un filtre texte local disparu
- [ ] Hors scope : supprimer `q` OpenAPI listes, FTS5

---

## SEARCH-004 — SQLite FTS5 (qualité / perf)

Labels: `area:data`, `enhancement`  
Bloqué par: SEARCH-001

### Objectif

Remplacer ou compléter les `LIKE %term%` du search global par FTS5 si le volume / pertinence le justifie.

### Critères d’acceptation

- [ ] Migration goose : tables/virtual FTS + triggers ou rebuild documenté
- [ ] `canonical.sql` mis à jour **via cette issue** (area:data)
- [ ] `GET /api/v1/search` consomme FTS (même contrat OpenAPI)
- [ ] Rebuild / backfill idempotent ; doc ops courte dans PLAN.md
- [ ] Tests pertinence basique (tokenisation) + non-régression RBAC
- [ ] Hors scope : Elasticsearch, recherche pièces jointes / commentaires

### Notes

Icebox jusqu’à signal (lenteur ou plaintes pertinence). Peut rester ouverte sans bloquer SEARCH-002/003.
