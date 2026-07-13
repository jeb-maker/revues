# Épique — Accès par équipes et gouvernance org

Issues à créer sur GitHub (repo `jeb-maker/revues`). Ordre d'implémentation strict.

```bash
# Depuis la racine du repo, avec gh authentifié :
./scripts/create-access-teams-issues.sh
```

---

## Epic — `[Epic] Accès par équipes et gouvernance org`

**Labels** : `epic`, `vague-5`, `area:auth`, `area:data`

### Contexte

Revues dispose déjà de :

- organisations multi-tenant (`organizations`, `organization_members`) ;
- rôles projet (`project_members` : lead / contributor / viewer) ;
- tags projet pour le **matching de modèles** (`project_tags`) — **pas** pour l'accès.

Les besoins identifiés :

1. **Délimitation** — qui voit quels projets et revues ;
2. **Autorisation** — ce que chacun peut faire (lecture vs contribution) ;
3. **Allocation scalable** — affecter des accès via des **équipes**, pas uniquement N×M invitations ;
4. **Mobilisation revue** — un lead ajoute une équipe existante sur son projet ;
5. **Supervision org** — l'admin org voit toute l'activité de son organisation ;
6. **Isolation** — projets privés si besoin intra-org.

Décisions produit (validées) :

- **Équipes nommées** = chemin principal d'accès collectif (pattern GitHub teams / Jira groups) ;
- **`project_members` direct** = exception (prestataire, renfort ponctuel) ;
- **`project_tags`** = classification / modèles **uniquement** — jamais d'accès ;
- **Org owner/admin** voit **tous** les projets et revues de l'org (supervision) ;
- **Org admin** gère membres et équipes ; actions métier (cocher, lancer) soumises au rôle **global** `editor` minimum ;
- **Lead** peut ajouter une équipe à son projet si politique org `leads_may_assign_teams` ;
- **Création org** : inchangé (1ʳᵉ org self-service, suivantes par invitation).

Hiérarchie cible :

```
Organisation
  ├── Équipes (organization_teams)
  │     └── Membres (team_members)
  ├── Projets
  │     ├── Membres directs (project_members) — exception
  │     ├── Équipes + rôle (team_project_roles)
  │     └── Tags (project_tags) — modèles seulement
  └── Revues → Points
```

Règle d'accès (normative — voir `docs/RBAC.md`) :

```
ResolveProjectAccess(user, project) :
  visible si admin global
          OU org owner/admin (org du projet)
          OU project_members direct
          OU ∃ équipe T : user ∈ T ∧ T a un rôle sur project

  rôle_effectif = max(rôle direct, max rôles via équipes)
  action permise = rôle_global suffisant ET rôle_effectif suffisant
```

### Issues filles

- [ ] #174 — Spec RBAC.md (équipes, org admin, projets privés)
- [ ] #175 — Migration teams + store
- [ ] #176 — ResolveProjectAccess + tests
- [ ] #177 — Refactor handlers sur ResolveProjectAccess
- [ ] #178 — Org admin voit tout + TestIDOR
- [ ] #179 — UI admin équipes CRUD
- [ ] #180 — UI projet — équipes + preview + sources
- [ ] #181 — Projets privés (visibility)
- [ ] #182 — Politiques org

### Hors scope épique

- Tags projet = accès (ABAC) ;
- sync SCIM / LDAP / Google Groups ;
- rôle « lead d'équipe » (gestion membres équipe réservée org admin en v1) ;
- expiration automatique des accès ;
- audit log admin complet (issue follow-up `area:auth`) ;
- modification de `docs/schema/canonical.sql` hors issues `area:data` dédiées.

---

## Issue 1 — `[auth] Spec RBAC — équipes, org admin, projets privés`

**Labels** : `area:auth`, `vague-5`  
**Dépendances** : aucune

### Objectif

Mettre à jour `docs/RBAC.md` comme document normatif pour le modèle équipes + gouvernance org, **avant** toute migration ou code.

### Critères d'acceptation

- [ ] `docs/RBAC.md` décrit les trois chemins d'accès : équipe→projet, membre direct, org admin ;
- [ ] Matrice actions étendue : org owner/admin (visibilité vs action métier) ;
- [ ] Règle `ResolveProjectAccess` documentée (visible, rôle effectif, sources) ;
- [ ] Tags projet explicitement **hors** périmètre accès ;
- [ ] Projets `private` : comportement documenté ;
- [ ] Politiques org listées (`leads_may_assign_teams`, etc.) ;
- [ ] Tests exigés listés : `TestRBAC_Matrix`, `TestIDOR_CrossProject`, `TestIDOR_TeamAccess`, `TestIDOR_OrgAdmin`, `TestIDOR_PrivateProject` ;
- [ ] PR dédiée `area:auth` — seul fichier métier modifié : `docs/RBAC.md` ;
- [ ] `./scripts/check.sh` vert.

### Notes

- Fichier sacré : validation produit requise avant merge.
- Ne pas implémenter de code dans cette issue.

---

## Issue 2 — `[data] Migration organization_teams + store`

**Labels** : `area:data`, `vague-5`  
**Bloqué par** : Issue 1

### Objectif

Introduire les tables équipes et la couche store.

### Critères d'acceptation

- [ ] Migration goose `00014_organization_teams.sql` :
  - `organization_teams` : `id`, `organization_id` FK, `name`, `slug`, `description`, `created_at`, UNIQUE `(organization_id, slug)`
  - `team_members` : `team_id`, `user_id`, `created_at`, PK `(team_id, user_id)`
  - `team_project_roles` : `team_id`, `project_id`, `role` CHECK (`lead`,`contributor`,`viewer`), `granted_by` FK nullable, `created_at`, PK `(team_id, project_id)`
  - Index : `team_members(user_id)`, `team_project_roles(project_id)`
- [ ] Store `internal/store/teams.go` :
  - `CreateTeam`, `TeamByID`, `ListOrganizationTeams`
  - `AddTeamMember`, `RemoveTeamMember`, `ListTeamMembers`
  - `GrantTeamProjectRole`, `RevokeTeamProjectRole`, `ListTeamProjects`, `ListProjectTeams`
  - `ListUserTeams(ctx, orgID, userID)`
- [ ] Slug équipe normalisé (lowercase, `[a-z0-9-]`, unique par org)
- [ ] Toutes les requêtes filtrent par `organization_id` du contexte
- [ ] Tests table-driven store
- [ ] `./scripts/check.sh` vert
- [ ] `docs/schema/canonical.sql` : issue `area:data` follow-up ou incluse si politique repo le permet

### Notes techniques

- Ne pas toucher `ResolveProjectAccess` ni handlers dans cette issue.
- Ne pas utiliser `project_tags` pour l'accès.

---

## Issue 3 — `[store][auth] ResolveProjectAccess + tests`

**Labels** : `area:auth`, `area:data`, `vague-5`  
**Bloqué par** : Issue 2

### Objectif

Implémenter la fonction unique de résolution d'accès projet conforme à `docs/RBAC.md`.

### Critères d'acceptation

- [ ] `internal/store/project_access.go` :
  ```go
  type ProjectAccess struct {
      Visible bool
      Role    string // lead | contributor | viewer | "" 
      Sources []string // "direct", "team:42", "org_admin", "global_admin"
  }
  func (s *Store) ResolveProjectAccess(ctx, userID, projectID int64, globalRole string) (ProjectAccess, error)
  ```
- [ ] Rôle effectif = `max(direct, équipes)` avec ordre lead > contributor > viewer
- [ ] Org owner/admin → `Visible=true`, `Role` vide ou `viewer` pour action (action gérée couche service)
- [ ] Admin global → visible + bypass action (inchangé)
- [ ] Tags projet **ignorés** pour l'accès
- [ ] `ListProjects` refactoré pour utiliser `ResolveProjectAccess` ou requête équivalente performante
- [ ] Tests table-driven : direct seul, équipe seule, direct+équipe (max rôle), hors périmètre, org admin, cross-org IDOR
- [ ] `./scripts/check.sh` vert

---

## Issue 4 — `[auth] Refactor handlers — ResolveProjectAccess`

**Labels** : `area:auth`, `area:core`, `vague-5`  
**Bloqué par** : Issue 3

### Objectif

Remplacer les appels dispersés à `MemberRole` + branche `admin` par `ResolveProjectAccess` sur toutes les routes sensibles.

### Critères d'acceptation

- [ ] Handlers `projects`, `runs`, `mytasks`, `attachments`, exports, Jira, Notion utilisent `ResolveProjectAccess`
- [ ] Suppression des checks ad hoc `MemberRole` redondants (garder `MemberRole` store si utile en interne)
- [ ] 404 uniforme si `Visible=false` (pas 403)
- [ ] `internal/web/rbac_test.go` : matrice étendue équipes
- [ ] `./scripts/check.sh` vert

### Fichiers cibles indicatifs

- `internal/features/projects/handlers.go`
- `internal/features/runs/handlers.go`
- `internal/features/mytasks/handlers.go`
- `internal/web/handlers/*` (runs, attachments, export, jira, notion)
- `internal/features/projects/service.go` (helpers CanView, CanLaunch, …)

---

## Issue 5 — `[auth] Org admin — visibilité globale org + TestIDOR`

**Labels** : `area:auth`, `vague-5`  
**Bloqué par** : Issue 3

### Objectif

Org owner/admin voit tous les projets et revues de l'organisation active sans membership direct ni équipe.

### Critères d'acceptation

- [ ] `ListProjects` : org owner/admin reçoit tous les projets non archivés de l'org
- [ ] `ListActiveRunSummaries` / page `/revues` : idem
- [ ] Accès GET projet/revue/export : org admin `Visible=true`
- [ ] Actions PATCH/POST métier : org admin soumis au rôle **global** (`editor` minimum pour cocher/lancer) — pas de bypass lead implicite
- [ ] Tests : org admin voit projet sans membership ; org `member` non ; cross-org 404
- [ ] `./scripts/check.sh` vert

---

## Issue 6 — `[ui] Admin équipes — CRUD`

**Labels** : `area:ui`, `area:admin`, `vague-5`  
**Bloqué par** : Issue 2

### Objectif

Interface org admin pour gérer les équipes et leurs membres.

### Critères d'acceptation

- [ ] Routes : `GET/POST /admin/teams`, `GET /admin/teams/{id}`, `POST /admin/teams/{id}/members`, `POST /admin/teams/{id}/members/remove`
- [ ] RBAC : org owner/admin uniquement (+ admin global)
- [ ] CSRF sur tous les POST
- [ ] Pages sobres HTMX, budget éco respecté
- [ ] Liste membres équipe avec login/email
- [ ] Tests handlers RBAC
- [ ] `./scripts/check.sh` vert

---

## Issue 7 — `[ui] Projet — équipes, preview, sources d'accès`

**Labels** : `area:ui`, `area:core`, `vague-5`  
**Bloqué par** : Issue 4, Issue 6

### Objectif

Sur la fiche projet, gérer les équipes affectées et afficher les sources d'accès.

### Critères d'acceptation

- [ ] Section « Équipes » : liste équipes + rôle sur ce projet ; formulaire ajout équipe + rôle
- [ ] Preview avant ajout : « Équipe X : N membres auront le rôle Y »
- [ ] Section « Membres directs » : inchangée fonctionnellement, badge source (`direct` vs `via équipe …` en lecture seule sur fiche user)
- [ ] RBAC ajout équipe : lead projet OU org owner/admin ; politique `leads_may_assign_teams` (stub true si Issue 9 pas merge)
- [ ] Retrait équipe du projet : org admin ou lead
- [ ] CSRF, tests handlers
- [ ] `./scripts/check.sh` vert

---

## Issue 8 — `[data][auth] Projets privés (visibility)`

**Labels** : `area:data`, `area:auth`, `vague-5`  
**Bloqué par** : Issue 3

### Objectif

Permettre d'isoler un projet au sein d'une org.

### Critères d'acceptation

- [ ] Migration : `projects.visibility` TEXT `normal`|`private`, défaut `normal`
- [ ] Projet `private` : inaccessible aux org `member` sans direct/équipe ; org owner/admin et admin global voient toujours
- [ ] Formulaire création/édition projet : champ visibilité (org admin ou lead)
- [ ] Badge « Privé » sur fiche projet et listes admin
- [ ] Tests `TestIDOR_PrivateProject`
- [ ] `./scripts/check.sh` vert

---

## Issue 9 — `[admin] Politiques org — délégation lead`

**Labels** : `area:admin`, `area:auth`, `vague-5`  
**Bloqué par** : Issue 7

### Objectif

Réglages org contrôlant ce que les leads peuvent faire.

### Critères d'acceptation

- [ ] Settings org (table existante `settings` ou colonnes dédiées) :
  - `leads_may_assign_teams` (bool, défaut `true`)
  - `leads_may_invite_members` (bool, défaut `true`)
  - `leads_may_invite_externals` (bool, défaut `false`)
- [ ] UI `/admin/settings` ou section org : édition par org owner/admin
- [ ] Handlers projet respectent les flags (refus serveur + message UI)
- [ ] Tests RBAC par politique
- [ ] `./scripts/check.sh` vert

---

## Prompt agent (Issue 1)

```
Repo jeb-maker/revues. Implémente UNIQUEMENT l'issue « [auth] Spec RBAC — équipes, org admin, projets privés »
(docs/issues/access-teams-epic.md — Issue 1).

Lis AGENTS.md, docs/CONVENTIONS.md, docs/REVIEW_ADVERSE.md.
Branche : cursor/issue-access-rbac-spec-f21b
PR titre : [auth] Spec RBAC — équipes, org admin, projets privés
Corps PR : Closes #N
Seul docs/RBAC.md modifié. ./scripts/check.sh avant push.
```

## Prompt agent (Issue 2)

```
Repo jeb-maker/revues. Implémente UNIQUEMENT l'issue « [data] Migration organization_teams + store »
(docs/issues/access-teams-epic.md — Issue 2).

Lis AGENTS.md, docs/CONVENTIONS.md, docs/GO.md, docs/RBAC.md, docs/schema/canonical.sql.
Branche : cursor/issue-access-teams-schema-f21b
PR titre : [data] Migration organization_teams + store
Corps PR : Closes #N
Pas de handlers/UI. ./scripts/check.sh avant push.
```
