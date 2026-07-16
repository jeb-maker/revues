# Épique — Accès par équipes et gouvernance org

Issues à créer sur GitHub (repo `jeb-maker/revues`). Ordre d'implémentation strict.

```bash
# Depuis la racine du repo, avec gh authentifié :
./scripts/create-access-teams-issues.sh
```

---

## Epic — `[Epic] Accès par équipes et gouvernance org`

**Labels** : `epic`, `vague-5`, `area:auth`, `area:data`

> **Note (2026)** : épique **rebasée sur `subjects`** (plus de `projects` / `project_members` / `project_tags`).  
> Voir [subjects-epic.md](./subjects-epic.md) pour le modèle v1 en vigueur (accès org-scoped) et [RBAC.md](../RBAC.md) § « Modèle cible équipes ».

### Contexte

Revues dispose déjà de :

- organisations multi-tenant (`organizations`, `organization_members`) ;
- **sujets** (`subjects`) avec accès v1 = membre de l'org ;
- domaines / étiquettes sujet pour **matching / classification** — **pas** pour l'accès.

Les besoins identifiés :

1. **Délimitation** — qui voit quels sujets et revues ;
2. **Autorisation** — ce que chacun peut faire (lecture vs contribution) ;
3. **Allocation scalable** — affecter des accès via des **équipes**, pas uniquement N×M invitations ;
4. **Mobilisation revue** — un lead ajoute une équipe existante sur son sujet ;
5. **Supervision org** — l'admin org voit toute l'activité de son organisation ;
6. **Isolation** — sujets privés si besoin intra-org.

Décisions produit (validées) :

- **Équipes nommées** = chemin principal d'accès collectif (pattern GitHub teams / Jira groups) ;
- **`subject_members` direct** = exception (prestataire, renfort ponctuel) ;
- **`subject_tags` / `subject_domains`** = classification / matching **uniquement** — jamais d'accès ;
- **Org owner/admin** voit **tous** les sujets et revues de l'org (supervision) ;
- **Org admin** gère membres et équipes ; actions métier (cocher, lancer) soumises au rôle **global** `editor` minimum ;
- **Lead** peut ajouter une équipe à son sujet si politique org `leads_may_assign_teams` ;
- **Création org** : inchangé (1ʳᵉ org self-service, suivantes par invitation).

Hiérarchie cible :

```
Organisation
  ├── Équipes (organization_teams)
  │     └── Membres (team_members)
  ├── Sujets
  │     ├── Membres directs (subject_members) — exception
  │     ├── Équipes + rôle (team_subject_roles)
  │     ├── Domaines (subject_domains) — matching modèles
  │     └── Étiquettes (subject_tags) — descriptif seulement
  └── Revues → Points
```

Règle d'accès (normative — voir `docs/RBAC.md`) :

```
ResolveSubjectAccess(user, subject) :
  visible si admin global
          OU org owner/admin (org du sujet)
          OU subject_members direct
          OU ∃ équipe T : user ∈ T ∧ T a un rôle sur subject

  rôle_effectif = max(rôle direct, max rôles via équipes)
  action permise = rôle_global suffisant ET rôle_effectif suffisant
```

### Issues filles

- [x] Spec RBAC.md (équipes, org admin, sujets privés) — vocabulaire `subjects`
- [x] Migration teams + store (`organization_teams`, `team_members`, `subject_members`, `team_subject_roles`) — greenfield `00001` + `canonical.sql`
- [x] `ResolveSubjectAccess` + tests (`internal/store/subject_access.go`)
- [x] Refactor handlers sur `ResolveSubjectAccess` (transition legacy ungated)
- [x] Org admin voit tout + TestIDOR
- [ ] UI admin équipes CRUD
- [ ] UI sujet — équipes + preview + sources
- [ ] Sujets privés (`visibility`)
- [ ] Politiques org

### Hors scope épique

- Étiquettes / domaines sujet = accès (ABAC) ;
- sync SCIM / LDAP / Google Groups ;
- rôle « lead d'équipe » (gestion membres équipe réservée org admin en v1) ;
- expiration automatique des accès ;
- audit log admin complet (issue follow-up `area:auth`) ;
- modification de `docs/schema/canonical.sql` hors issues `area:data` dédiées.

---

## Issue 1 — `[auth] Spec RBAC — équipes, org admin, sujets privés`

**Labels** : `area:auth`, `vague-5`  
**Dépendances** : aucune  
**Statut** : vocabulaire `subjects` appliqué dans `docs/RBAC.md` (modèle cible) + intro de cette épique.

### Objectif

Mettre à jour `docs/RBAC.md` comme document normatif pour le modèle équipes + gouvernance org sur **sujets**, **avant** toute migration ou code.

### Critères d'acceptation

- [x] `docs/RBAC.md` décrit les trois chemins d'accès : équipe→sujet, membre direct, org admin ;
- [x] Matrice actions étendue : org owner/admin (visibilité vs action métier) ;
- [x] Règle `ResolveSubjectAccess` documentée (visible, rôle effectif, sources) ;
- [x] Étiquettes / domaines sujet explicitement **hors** périmètre accès ;
- [x] Sujets `private` : comportement documenté ;
- [x] Politiques org listées (`leads_may_assign_teams`, etc.) ;
- [x] Tests exigés listés : `TestRBAC_Matrix`, `TestIDOR_CrossSubject`, `TestIDOR_TeamAccess`, `TestIDOR_OrgAdmin`, `TestIDOR_PrivateSubject` ;
- [ ] Issues filles 2–9 encore à réécrire entièrement (détail migration / UI) ;
- [ ] `./scripts/check.sh` vert.

### Notes

- Fichier sacré : validation produit requise avant merge.
- Ne pas implémenter de code dans cette issue.

---

## Issue 2 — `[data] Migration organization_teams + store`

**Labels** : `area:data`, `vague-5`  
**Bloqué par** : Issue 1  
**Statut** : livré (greenfield — tables dans `00001_initial_schema.sql` + `canonical.sql`, store `teams.go`).

### Objectif

Introduire les tables équipes / accès sujet et la couche store.

### Critères d'acceptation

- [x] Schéma (greenfield, pas `00014`) :
  - `organization_teams`, `team_members`
  - `subject_members` (chemin direct)
  - `team_subject_roles` + index `subject_id` / `user_id`
- [x] Store `internal/store/teams.go` :
  - `CreateTeam`, `TeamByID`, `ListOrganizationTeams`
  - `AddTeamMember`, `RemoveTeamMember`, `ListTeamMembers`, `ListUserTeams`
  - `UpsertDirectSubjectMember`, `RemoveDirectSubjectMember`, `ListDirectSubjectMembers`
  - `GrantTeamSubjectRole`, `RevokeTeamSubjectRole`, `ListTeamSubjects`, `ListSubjectTeams`
- [x] Slug équipe normalisé ; requêtes scopées org active
- [x] Tests store ; `canonical.sql` mis à jour

### Notes techniques

- Ne pas toucher `ResolveSubjectAccess` ni handlers dans cette issue.
- `ListSubjectMembers` (v1 = membres org pour assignation) **conservé** — distinct de `ListDirectSubjectMembers`.

---

## Issue 3 — `[store][auth] ResolveSubjectAccess + tests`

**Labels** : `area:auth`, `area:data`, `vague-5`  
**Bloqué par** : Issue 2  
**Statut** : livré (`subject_access.go` + listing aligné handlers).

### Objectif

Implémenter la fonction unique de résolution d'accès sujet conforme à `docs/RBAC.md`.

### Critères d'acceptation

- [x] `internal/store/subject_access.go` : `SubjectAccess` + `ResolveSubjectAccess`
- [x] Rôle effectif = `max(direct, équipes)` lead > contributor > viewer
- [x] Org owner/admin → `Visible=true`, `Role` vide (actions en service)
- [x] Admin global → visible + source `global_admin`
- [x] Étiquettes / domaines **ignorés** pour l'accès
- [x] `ListSubjects` filtré via accès (+ legacy ungated)
- [x] Tests table-driven : direct, équipe, max rôle, hors périmètre, org admin, cross-org, legacy

---

## Issue 4 — `[auth] Refactor handlers — ResolveSubjectAccess`

**Labels** : `area:auth`, `area:core`, `vague-5`  
**Bloqué par** : Issue 3  
**Statut** : livré (transition legacy : sujets sans grants restent visibles aux membres org)

### Objectif

Remplacer les appels dispersés à `MemberRole` + branche `admin` par `ResolveSubjectAccess` sur toutes les routes sensibles.

### Critères d'acceptation

- [x] Handlers `subjects`, `runs`, `checklisttemplates`, attachments, exports, Jira, Notion utilisent `ResolveSubjectAccess`
- [x] Suppression des checks ad hoc `MemberRole` redondants (garder `MemberRole` store si utile en interne)
- [x] 404 uniforme si `Visible=false` (pas 403)
- [x] Listings revues (`ListFiltered` / active / completed) alignés sur la visibilité sujet
- [x] `internal/web/rbac_test.go` : matrice étendue équipes (sujet gated)
- [x] `./scripts/check.sh` vert

### Fichiers cibles indicatifs

- `internal/features/subjects/handlers.go`
- `internal/features/runs/handlers*.go`
- `internal/features/checklisttemplates/handlers.go`
- `internal/store/dashboard.go` (filtres listing)
- `internal/features/subjects/service.go` (helpers CanViewAccess, CanContributeAccess, …)

---

## Issue 5 — `[auth] Org admin — visibilité globale org + TestIDOR`

**Labels** : `area:auth`, `vague-5`  
**Bloqué par** : Issue 3  
**Statut** : livré (visibilité org admin + `TestIDOR_OrgAdmin` ; pas de bypass lead implicite).

### Objectif

Org owner/admin voit tous les sujets et revues de l'organisation active sans membership direct ni équipe.

### Critères d'acceptation

- [x] `ListSubjects` : org owner/admin reçoit tous les sujets non archivés de l'org
- [x] `ListActiveRunSummaries` / page `/revues` : idem
- [x] Accès GET sujet/revue/export : org admin `Visible=true`
- [x] Actions PATCH/POST métier : org admin soumis au rôle **global** (`editor` minimum pour cocher/lancer) — pas de bypass lead implicite
- [x] Tests : org admin voit sujet sans membership ; org `member` non ; cross-org 404
- [x] `./scripts/check.sh` vert

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
