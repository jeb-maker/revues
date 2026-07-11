# Épique — Organisations multi-tenant (self-service)

Issues à créer sur GitHub (repo `jeb-maker/revues`). Ordre d'implémentation strict.

```bash
# Depuis la racine du repo, avec gh authentifié :
./scripts/create-organization-issues.sh
```

---

## Epic — `[Epic] Organisations multi-tenant self-service`

**Labels** : `epic`, `vague-4` (ou backlog), `area:data`, `area:auth`

### Contexte

Revues v1 est mono-instance : une whitelist globale (`allowed_emails`), pas de notion d'organisation. Pour le self-service B2B, introduire une couche **Organisation** au-dessus des **Projets**.

Hiérarchie cible :

```
Organisation
  └── Projet (« espace de travail » en UI — inchangé)
        └── Revue → Points
```

Décisions produit (validées) :

- Terme UI : **Organisation** (pas « domaine », pas « espace »)
- Self-service : le premier utilisateur **crée** une org (nom libre + slug)
- Un email peut appartenir à **plusieurs orgs** → choix / switcher au login
- Invitation **projet** → adhésion **org** induite (rôle org minimal `member`)
- Le domaine email sert de suggestion, pas de clé technique

### Issues filles

- [ ] #TBD — Schéma DB + store organizations
- [ ] #TBD — Session + organisation active + middleware
- [ ] #TBD — Parcours auth : création org + sélecteur multi-org
- [ ] #TBD — Scoper projets et données métier par org
- [ ] #TBD — Invitation projet → membership org induite
- [ ] #TBD — Remplacer whitelist globale par gestion org
- [ ] #TBD — UI switcher + invitations en attente

### Hors scope épique

- PostgreSQL / schéma par tenant
- Domaine email vérifié automatique (`@acme.com`)
- Google OAuth

---

## Issue 1 — `[data] Schéma organizations + organization_members + store`

**Labels** : `area:data`, `vague-4`  
**Dépendances** : aucune (base technique)

### Objectif

Introduire les tables `organizations` et `organization_members` avec couche store testée.

### Critères d'acceptation

- [ ] Migration goose `00008_organizations.sql` :
  - `organizations` : `id`, `name` (TEXT NOT NULL), `slug` (TEXT NOT NULL UNIQUE), `created_at`, `created_by` (FK users, nullable)
  - `organization_members` : `organization_id`, `user_id`, `role` CHECK (`owner`, `admin`, `member`), `created_at`, PK `(organization_id, user_id)`
  - Index sur `organization_members(user_id)`
- [ ] `docs/schema/canonical.sql` aligné (issue `area:data`)
- [ ] Store (`internal/store/organizations.go`) :
  - `CreateOrganization(ctx, name, slug, createdBy) (*Organization, error)` — slug normalisé (lowercase, `[a-z0-9-]`, unique)
  - `OrganizationBySlug`, `OrganizationByID`
  - `AddOrganizationMember`, `RemoveOrganizationMember`, `MemberRole`
  - `ListUserOrganizations(ctx, userID) ([]OrganizationMembership, error)`
  - `CountUserOrganizations(ctx, userID) (int, error)`
- [ ] Migration données existantes : org par défaut `default` (slug) + rattacher tous les users existants en `owner` pour le bootstrap admin, `member` pour les autres
- [ ] Tests table-driven store (création, slug dupliqué, list memberships)
- [ ] `./scripts/check.sh` vert

### Notes techniques

- SQLite, placeholders `?`, pas de `LastInsertId` fragile si évitable (SELECT after insert OK comme ailleurs)
- Ne pas toucher auth/handlers/UI dans cette issue

---

## Issue 2 — `[auth] Organisation active en session + middleware`

**Labels** : `area:auth`, `vague-4`  
**Bloqué par** : Issue 1

### Objectif

Persister l'organisation active dans la session et l'injecter dans le contexte requête.

### Critères d'acceptation

- [ ] Colonne `organization_id` (FK, NOT NULL après migration) sur `sessions`
- [ ] `CreateSession` accepte `organizationID`
- [ ] Middleware : après auth user, charge org active ; si session sans org valide → redirect `/org/select`
- [ ] Vérifier membership : user doit être membre de l'org active (sinon logout ou redirect)
- [ ] Helper `OrganizationFromContext(ctx) (*Organization, bool)`
- [ ] Tests middleware + session
- [ ] `./scripts/check.sh` vert

---

## Issue 3 — `[auth][ui] Création org self-service + sélecteur multi-org`

**Labels** : `area:auth`, `area:ui`, `vague-4`  
**Bloqué par** : Issue 2

### Objectif

Parcours post-login GitHub selon le nombre d'organisations de l'utilisateur.

### Critères d'acceptation

- [ ] **0 org** : redirect `/org/new` — formulaire nom + slug (slug pré-rempli depuis nom), créateur = `owner`, session org active, redirect dashboard
- [ ] **1 org** : sélection automatique, session org active
- [ ] **N orgs** : page `/org/select` liste des orgs, POST choisit org active
- [ ] Mémoriser dernière org (cookie ou colonne session) pour défaut sur `/org/select`
- [ ] CSRF sur tous les POST
- [ ] Pages sobres HTMX, budget éco respecté
- [ ] `./scripts/check.sh` vert

---

## Issue 4 — `[data][core] Scoper projets et entités métier par organization_id`

**Labels** : `area:data`, `area:core`, `vague-4`  
**Bloqué par** : Issue 2

### Objectif

Toutes les données métier appartiennent à une organisation.

### Critères d'acceptation

- [ ] `organization_id NOT NULL` sur : `projects`, `checklist_templates` (via project ou direct), `settings`, `integrations`, `allowed_emails` (renommer conceptuellement en org-scoped)
- [ ] Migration : rattacher au org `default`
- [ ] Toutes les requêtes store filtrent par `organization_id` du contexte
- [ ] IDOR : impossible d'accéder à un projet d'une autre org (tests)
- [ ] `./scripts/check.sh` vert

---

## Issue 5 — `[core] Invitation projet → adhésion org induite`

**Labels** : `area:core`, `vague-4`  
**Bloqué par** : Issue 4

### Objectif

Ajouter un membre au projet l'inscrit automatiquement à l'organisation si nécessaire.

### Critères d'acceptation

- [ ] Lors de l'ajout membre projet : si user pas membre org → `AddOrganizationMember(..., role=member)` puis ajout projet
- [ ] Seuls `owner`, `admin` org ou `lead` projet peuvent inviter une adresse **absente de l'org** (RBAC serveur)
- [ ] Message UI explicite : « Cette personne sera ajoutée à l'organisation »
- [ ] User invité pas encore inscrit (`UserByEmail` absent) : conserver comportement « doit s'être connecté une fois » en v1 ; prévoir champ `organization_invitations` optionnel (email, org_id, project_id, role) si scope permet — sinon issue follow-up
- [ ] Tests RBAC invitation cross-org refusée
- [ ] `./scripts/check.sh` vert

---

## Issue 6 — `[admin][auth] Whitelist globale → membres / invitations org`

**Labels** : `area:admin`, `area:auth`, `vague-4`  
**Bloqué par** : Issue 4

### Objectif

Remplacer la whitelist instance-wide par la gestion au niveau organisation.

### Critères d'acceptation

- [ ] Admin org (`owner`/`admin`) gère qui peut rejoindre (liste emails autorisés scoped org, ou invitations)
- [ ] `ResolveLoginRole` : login autorisé si membre d'au moins une org OU invitation pending OU création nouvelle org (0 org)
- [ ] Supprimer / déprécier écran admin global `/admin/users` ou le scoper org admin
- [ ] `REVUES_BOOTSTRAP_ADMIN_EMAIL` : crée org default + owner (migration path documentée dans ONBOARDING.md)
- [ ] Matrice RBAC mise à jour dans PR (docs/RBAC.md = fichier sacré → issue dédiée ou mention « hors scope, follow-up #RBAC-org »)
- [ ] `./scripts/check.sh` vert

---

## Issue 7 — `[ui] Switcher organisation + invitations en attente`

**Labels** : `area:ui`, `vague-4`  
**Bloqué par** : Issue 3, Issue 5

### Objectif

Navigation multi-org dans le chrome applicatif.

### Critères d'acceptation

- [ ] Switcher org dans le header (liste orgs user, POST change session org)
- [ ] Au login : bandeau invitations pending si table `organization_invitations` existe
- [ ] Accepter invitation → member org + redirect projet si applicable
- [ ] `./scripts/check.sh` vert

---

## Prompt agent (Issue 1)

```
Repo jeb-maker/revues. Implémente UNIQUEMENT l'issue « [data] Schéma organizations + organization_members + store »
(docs/issues/organizations-epic.md — Issue 1).

Lis AGENTS.md, docs/CONVENTIONS.md, docs/GO.md, docs/DEFINITION_OF_DONE.md.
Branche : cursor/issue-organizations-schema-f21b
PR titre : [data] Schéma organizations + organization_members + store
Corps PR : Closes #N (remplacer N après création issue GitHub)
./scripts/check.sh avant push.
```
