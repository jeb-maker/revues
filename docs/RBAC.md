# Matrice RBAC — Revues

Document normatif. Toute PR touchant une route doit mettre à jour la matrice dans la description PR.

## Rôles

### Globaux (`users.role`)

| Rôle | Description |
|------|-------------|
| `admin` | Tout + admin système (users, SMTP, intégrations) ; voit tous les projets de l'org active |
| `editor` | Créer modèles, lancer revues, cocher (tous projets où accès projet) |
| `reader` | Lecture seule (plafond — ne coche pas même si rôle projet contributor) |

### Organisation (`organization_members.role`)

| Rôle | Description |
|------|-------------|
| `owner` | Gouvernance org : équipes, whitelist, politiques ; **voit tous** les projets/revues de l'org |
| `admin` | Idem `owner` sauf actions réservées owner si ajoutées ultérieurement |
| `member` | Membre org ; accès projet via équipe, membership direct, ou invitation |

### Projet — direct (`project_members.role`)

| Rôle | Description |
|------|-------------|
| `lead` | Gérer membres et équipes du projet (si politique org), tout faire sur le projet |
| `contributor` | Cocher, commenter, lancer revues |
| `viewer` | Lecture seule sur ce projet |

### Projet — via équipe (`team_project_roles.role`)

Même sémantique que `project_members.role`. Une équipe se voit attribuer **un rôle** sur un projet ; chaque membre de l'équipe hérite de ce rôle pour ce projet.

---

## Chemins d'accès à un projet

Un utilisateur accède à un projet par **exactement l'un** des mécanismes suivants (évalués par `ResolveProjectAccess`) :

| Chemin | Condition | Usage |
|--------|-----------|-------|
| **Global admin** | `users.role = admin` | Bypass org et projet |
| **Org admin** | `organization_members.role ∈ {owner, admin}` dans l'org du projet | Supervision : voit tout dans l'org |
| **Membre direct** | Ligne `project_members` | Exception : invité, prestataire, renfort |
| **Équipe** | ∃ équipe T : user ∈ `team_members` ∧ (T, projet) ∈ `team_project_roles` | Cas nominal collectif |

### Hors périmètre accès

- **`project_tags`** : classification métier et matching de modèles **uniquement**. Un tag projet **ne donne jamais** d'accès.
- **`template_tags`** : idem, matching modèles seulement.

---

## Règle de composition

```
ResolveProjectAccess(user, project) :

  1. Si admin global → Visible=true, bypass actions

  2. Si org owner/admin (org du projet) → Visible=true
     (actions métier soumises au rôle global, pas de bypass lead implicite)

  3. Sinon calculer :
     role_direct  = project_members.role ou ∅
     role_teams   = max(team_project_roles.role) pour équipes du user
     role_effectif = max(role_direct, role_teams)   // lead > contributor > viewer

  4. Visible = (role_effectif ≠ ∅) OU org admin OU admin global

  5. Action :
     - Vérifier rôle global (reader ne coche pas)
     - Vérifier role_effectif suffisant pour l'action
     - Org admin : lecture/export OK ; écriture seulement si editor+ global

  6. Si ¬Visible → 404 (pas 403)
```

### Ordre des rôles (max)

```
lead > contributor > viewer
```

### Sources d'accès (affichage UI / debug)

| Source | Valeur `Sources` |
|--------|------------------|
| Membership direct | `direct` |
| Équipe | `team:{team_id}` |
| Org admin | `org_admin` |
| Admin global | `global_admin` |

---

## Projets privés (`projects.visibility`)

| Visibilité | Membre org sans accès direct/équipe | Org owner/admin | Admin global |
|------------|-------------------------------------|-----------------|--------------|
| `normal` | 404 | visible | visible |
| `private` | 404 | visible | visible |

Un projet privé n'apparaît pas aux org `member` sans chemin d'accès explicite. L'org admin le voit toujours (supervision).

---

## Politiques organisation (settings)

| Clé | Défaut | Effet |
|-----|--------|-------|
| `leads_may_assign_teams` | `true` | Lead peut ajouter une équipe existante à son projet |
| `leads_may_invite_members` | `true` | Lead peut inviter un membre direct |
| `leads_may_invite_externals` | `false` | Lead peut inviter une adresse hors org |

Seuls org `owner` / `admin` modifient ces politiques.

---

## Matrice des actions

| Action | admin global | org owner/admin | editor + lead | editor + contributor | editor + viewer | reader + * |
|--------|--------------|-----------------|---------------|----------------------|-----------------|------------|
| Admin users/SMTP/intégrations | ✓ | — | — | — | — | — |
| Gérer équipes org | ✓ | ✓ | — | — | — | — |
| Créer projet | ✓ | ✓ | ✓ | ✓ | — | — |
| Voir tous projets org | ✓ | ✓ | — | — | — | — |
| Voir projet (accès direct/équipe) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Gérer membres projet | ✓ | ✓ | ✓ (lead) | — | — | — |
| Ajouter équipe au projet | ✓ | ✓ | ✓ (lead)* | — | — | — |
| CRUD modèles | ✓ | ✓† | ✓ (membre) | — | — | — |
| Lancer revue | ✓ | ✓† | ✓ (lead/contrib) | ✓ | — | — |
| Cocher / commenter point | ✓ | ✓† | ✓ (lead/contrib) | ✓ | — | — |
| Assigner point | ✓ | ✓† | ✓ (lead) | — | — | — |
| Clôturer revue | ✓ | ✓† | ✓ (lead) | — | — | — |
| Lire revue / export CSV | ✓ | ✓ | ✓ (membre) | ✓ | ✓ | ✓ (membre) |
| Lier / créer ticket Jira | ✓ | ✓† | ✓ (lead/contrib) | ✓ | — | — |
| Config intégrations | ✓ | — | — | — | — | — |

\* Si politique `leads_may_assign_teams`.  
† Org admin : seulement si rôle global `editor` minimum pour les actions d'écriture ; lecture/export sans restriction.

---

## Routes — contrôles requis

| Route | Contrôle |
|-------|----------|
| `GET /projects` | Auth ; liste selon `ResolveProjectAccess` ; org admin → tous projets org |
| `GET /projects/{id}` | Auth + `Visible` |
| `POST /projects/{id}/teams` | Auth + lead ou org admin ; politique `leads_may_assign_teams` |
| `POST /projects/{id}/members` | Auth + lead ou org admin ; politique `leads_may_invite_members` |
| `GET /admin/teams` | Auth + org owner/admin |
| `POST /projects/{id}/runs` | Auth + contributor+ effectif ou admin |
| `GET /runs/{id}` | Auth + `Visible` sur projet de la revue |
| `PATCH /runs/{id}/items/{itemId}` | Auth + contributor+ effectif ou admin |
| `GET /attachments/{id}` | Auth + membre projet de la revue liée |
| `POST /admin/*` | Auth + admin global ou org admin selon route |

Toutes les routes sensibles appellent `ResolveProjectAccess` (ou helper dérivé) — pas de `MemberRole` seul.

---

## Tests obligatoires

Chaque PR `area:auth` ou `area:core` ajoute ou maintient :

```go
TestRBAC_Matrix          // table-driven : rôle global × org × équipe × direct × route → status
TestIDOR_CrossProject    // user A n'accède pas ressources projet B
TestIDOR_CrossOrg        // user org A n'accède pas projet org B
TestIDOR_TeamAccess      // accès via équipe ; retrait équipe → 404
TestIDOR_OrgAdmin        // org admin voit sans membership ; member non
TestIDOR_PrivateProject  // private invisible sauf accès explicite ou org admin
TestCSRF_MissingToken    // POST sans CSRF → 403
```

Fichiers cibles : `internal/web/rbac_test.go`, `internal/store/project_access_test.go`.

---

## Évolution

Modifier ce fichier uniquement via PR dédiée `area:auth` avec validation produit.

Spec épique : [docs/issues/access-teams-epic.md](./issues/access-teams-epic.md).
