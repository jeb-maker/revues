# Work packages — rewrite Go API + SvelteKit

> **Historique** — tous les WP ci-dessous sont mergés (octobre 2026, clôture WP-030 #274). Les cases à cocher ne sont pas tenues à jour : le suivi fait foi sur GitHub (issues / PR listées dans [ISSUE_MAP.md](./ISSUE_MAP.md)). Ce document reste la référence des périmètres et critères d'origine.

Chaque `## WP-xxx` = une issue GitHub. Labels : première ligne `Labels:` · dépendances `Bloqué par:`.

---

## WP-001 — Scaffold big bang : layout API + SvelteKit, retrait UI templates

Labels: `epic`, `area:infra`  
Bloqué par: aucune

### Objectif

Poser l’arborescence cible et **retirer** l’UI `html/template` + HTMX pour que l’app ne dépende plus du rendu serveur métier.

### Critères d’acceptation

- [ ] Arborescence documentée dans `docs/CONVENTIONS.md` : `api/openapi/`, `frontend/` (SvelteKit), `internal/` conservé pour métier/store
- [ ] App SvelteKit minimale buildable (`frontend/`) page placeholder
- [ ] Go sert `/healthz` et, si possible, les assets SPA (ou stub documenté)
- [ ] Suppression (ou isolation gitignored morte interdite) de `web/templates/**` pages métier et routes HTML associées — l’ancienne UI ne doit plus être le chemin nominal
- [ ] README + PLAN alignés (plus de « HTMX imposé »)
- [ ] `./scripts/check.sh` adapté au minimum pour ne pas exiger l’ancien budget HTMX cassé sans remplacement
- [ ] Hors scope : OpenAPI complet, sqlc, auth UI, domaines métier

### Notes techniques

- Big bang assumé : casser les handlers HTML est OK.
- Conserver `internal/store`, auth, intégrations pour réemploi.
- mb : prévoir emplacement vendor consommé par SvelteKit (voir WP-005 si reporté).

---

## WP-002 — Toolchain OpenAPI : spec de base, codegen serveur, `/api/v1`

Labels: `area:infra`, `area:core`  
Bloqué par: WP-001

### Objectif

OpenAPI comme source de vérité ; génération des stubs serveur ; première route versionnée.

### Critères d’acceptation

- [ ] Spec `api/openapi/openapi.yaml` avec `openapi: 3.x`, tag `system`, path `GET /api/v1/health` (ou équivalent)
- [ ] Outil choisi (`ogen` **ou** `oapi-codegen`) documenté dans `docs/API.md` + script `make api` / `./scripts/generate-api.sh`
- [ ] Code généré commité ou généré en CI de façon reproductible (trancher et documenter)
- [ ] Router chi monte les handlers générés sous `/api/v1`
- [ ] Test HTTP minimal sur health API
- [ ] Hors scope : domaines métier, sqlc, front client généré (WP-005)

---

## WP-003 — sqlc : requêtes typées alignées schéma

Labels: `area:data`  
Bloqué par: WP-001

### Objectif

Introduire sqlc pour les accès données ciblés rewrite, sans ORM.

### Critères d’acceptation

- [ ] `sqlc.yaml` + dossier requêtes ; génération documentée (`make sqlc` / script)
- [ ] Au moins les requêtes nécessaires au health métier suivant (runs/items ou users/sessions) migrées en sqlc **ou** pont clair store existant → sqlc
- [ ] Pas de régression : `go test ./internal/store/...` vert
- [ ] `docs/API.md` / CONVENTIONS : SQL uniquement via sqlc ou `internal/store`
- [ ] Hors scope : réécrire tout le store d’un coup si trop large — alors lister les fichiers restants et issue suiveuse `TODO(#N)`

### Notes

Peut avancer en parallèle de WP-002 après WP-001.

---

## WP-004 — Auth API + shell Svelte login/logout/OAuth/CSRF

Labels: `area:auth`  
Bloqué par: WP-002  
Revue humaine: oui (OAuth)

### Objectif

Remplacer les pages login templates par API auth + écrans SvelteKit ; conserver sessions + CSRF.

### Critères d’acceptation

- [ ] Endpoints OpenAPI : bootstrap/`me`, login local, register, logout ; flux GitHub start/callback fonctionnel
- [ ] CSRF : token exposé au front, validé sur mutations API
- [ ] Pages SvelteKit login/register (+ erreurs) utilisant mb
- [ ] Dev auth / seed session documenté dans AGENTS.md Cloud section
- [ ] Tests sécurité : CSRF reject, session requise, whitelist email
- [ ] RBAC inchangé sémantiquement ([RBAC.md](../RBAC.md))
- [ ] Hors scope : admin users complet, orgs UI

---

## WP-005 — Front : client OpenAPI généré, mb, CI/check.sh dual stack

Labels: `area:infra`, `area:ui`  
Bloqué par: WP-002, WP-004 (auth bootstrap pour appels authentifiés)

### Objectif

Client TS généré, mb intégré, gatekeeper CI pour Go + frontend.

### Critères d’acceptation

- [ ] Génération client TS depuis OpenAPI ; utilisé sous `frontend/src/lib/api/` (ou équivalent)
- [ ] mb tokens + CE utilisables depuis SvelteKit (vendor ou package)
- [ ] Proxy dev Vite → Go documenté
- [ ] `./scripts/check.sh` : Go existant + `npm ci`/`npm run check`/`build` front ; nouveaux budgets documentés dans PLAN.md
- [ ] CI GitHub Actions exécute le même gatekeeper
- [ ] Hors scope : pages métier domaines

---

## WP-010 — Organisations : API + UI Svelte

Labels: `area:core`, `area:ui`  
Bloqué par: WP-005

### Critères d’acceptation

- [ ] OpenAPI : liste/création org, sélection org active, invitations (parité comportement actuel)
- [ ] Écrans SvelteKit + mb
- [ ] RBAC/IDOR + CSRF
- [ ] Tests API table-driven sur les cas nominaux + refus
- [ ] Hors scope : admin teams détaillé (WP-015)

---

## WP-011 — Sujets (subjects) : API + UI

Labels: `area:core`, `area:ui`  
Bloqué par: WP-005

### Critères d’acceptation

- [ ] CRUD/list sujets, membres, tags/domains selon parité actuelle
- [ ] UI liste + détail + actions principales en SvelteKit/mb
- [ ] IDOR org active
- [ ] Tests accès
- [ ] Hors scope : lancement revue (WP-013)

---

## WP-012 — Modèles de check-list : API + UI éditeur

Labels: `area:core`, `area:ui`  
Bloqué par: WP-005

### Critères d’acceptation

- [ ] API templates + versions + items ; **immutabilité versions publiées** respectée
- [ ] UI liste + édition (parité fonctionnelle éditeur ; DnD si déjà présent)
- [ ] Tests métier versionnement
- [ ] Hors scope : import Notion (WP-021)

---

## WP-013 — Revues (runs) + items : API + UI

Labels: `area:core`, `area:ui`  
Bloqué par: WP-005, WP-011, WP-012

### Critères d’acceptation

- [ ] Lancer revue = snapshot transactionnel
- [ ] `PATCH` item : status/comment/assign ; commentaire obligatoire si `nok`
- [ ] Events d’audit ; run non éditable si terminé
- [ ] UI run + item + progression en SvelteKit/mb
- [ ] Webhook emit `review.item.nok` / complete branché si infra webhooks encore présente
- [ ] Tests statut/validation/accès (reprendre cas `internal/features/runs/*_test.go`)
- [ ] Hors scope : Jira/Notion UI (vague 3)

---

## WP-014 — Mes tâches : API + UI

Labels: `area:core`, `area:ui`  
Bloqué par: WP-013

### Critères d’acceptation

- [ ] Liste items assignés (filtres parité)
- [ ] Page SvelteKit + mb-table
- [ ] Tests API
- [ ] Hors scope : notifications email

---

## WP-015 — Admin org : users, teams, policies

Labels: `area:admin`, `area:ui`  
Bloqué par: WP-005, WP-010

### Critères d’acceptation

- [ ] API + UI : allowed emails, rôles, teams, lead policies (parité)
- [ ] `RequireOrgAdmin` / équivalent API
- [ ] Tests RBAC admin
- [ ] Hors scope : SMTP/Jira/Notion/webhooks config (WP-016 / vague 3)

---

## WP-016 — Admin SMTP + hub intégrations + attachments API/UI

Labels: `area:admin`, `area:notifications`, `area:attachments`, `area:ui`  
Bloqué par: WP-005, WP-013

### Critères d’acceptation

- [ ] Config SMTP chiffrée (settings) via API + UI
- [ ] Hub intégrations (liste états) + navigation
- [ ] Upload/download attachments run item (magic bytes, disposition) via API + UI
- [ ] Tests anti-régression sécu uploads + absence secret en clair
- [ ] Hors scope : implémentation Jira/Notion/webhooks delivery (vague 3)

---

## WP-020 — Jira Cloud : API + UI

Labels: `area:integrations`  
Bloqué par: WP-016  
Revue humaine: oui

### Critères d’acceptation

- [ ] Config credentials chiffrés, test connexion, link/create issue depuis item
- [ ] UI SvelteKit (plus de forms HTMX)
- [ ] safehttp / anti-SSRF conservés
- [ ] Tests mock API
- [ ] Hors scope : Jira Server/DC (#65 icebox)

---

## WP-021 — Notion : API + UI

Labels: `area:integrations`  
Bloqué par: WP-016  
Revue humaine: oui

### Critères d’acceptation

- [ ] Import template / export run parité
- [ ] UI SvelteKit
- [ ] Tests mock
- [ ] Hors scope : nouvelles features Notion

---

## WP-022 — Webhooks sortants : API + UI + drain

Labels: `area:integrations`  
Bloqué par: WP-016  
Revue humaine: oui

### Critères d’acceptation

- [ ] Config, test, signature HMAC, queue `webhook_deliveries`, drain
- [ ] UI admin SvelteKit
- [ ] Tests HMAC + anti-SSRF
- [ ] Doc [WEBHOOKS.md](../WEBHOOKS.md) à jour si chemins changent

---

## WP-030 — Clôture rewrite : purge legacy, docs, budgets

Labels: `area:infra`, `epic`  
Bloqué par: WP-013, WP-015 (minimum) ; idéalement vague 3 mergeée

### Critères d’acceptation

- [ ] Plus aucune dépendance runtime à HTMX / templates Go métier
- [ ] `grep` CI ou check : pas de réintroduction `web/templates/pages` mort
- [ ] AGENTS.md, PLAN, CONVENTIONS, DoD, ONBOARDING, ROADMAP cohérents
- [ ] Budgets SPA mesurés et seuils appliqués dans `check.sh`
- [ ] README démarrage : API + `npm run dev` / binaire unique
- [ ] Hors scope : nouvelles features produit

---

## Mapping labels → skills agents

| Label | Focus lecture |
|-------|----------------|
| `area:infra` | CONVENTIONS, CI, scripts |
| `area:data` | canonical.sql, goose, sqlc |
| `area:auth` | RBAC, ADR, tests sécu |
| `area:core` | PLAN règles métier, store |
| `area:ui` | FRONTEND.md, mb |
| `area:admin` | RBAC org admin |
| `area:integrations` | WEBHOOKS/NOTION/jira + revue humaine |
| `area:attachments` | CONVENTIONS uploads |
| `area:notifications` | NOTIFICATIONS.md |
