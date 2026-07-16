# Matrice RBAC — Revues

Document normatif. Toute PR touchant une route doit mettre à jour la matrice dans la description PR.

## Sujets v1 (greenfield — en vigueur)

Modèle actuel après migration `subjects` (épique [subjects-epic.md](./issues/subjects-epic.md)).

| Entité | Accès v1 |
|--------|----------|
| **Sujet** | Membre de l'organisation du sujet (`organization_members`) |
| **Étiquettes** (`subject_tags`) | Classification descriptive — **jamais** d'accès |
| **Domaines** (`subject_domains`, `template_domains`) | Matching modèles ↔ sujet — **jamais** d'accès |

### Permissions sujet (`internal/features/subjects/service.go`)

| Action | admin global | org owner/admin | editor + membre org | reader + membre org |
|--------|--------------|-----------------|---------------------|---------------------|
| Voir sujet | ✓ | ✓ | ✓ | ✓ |
| Créer sujet | ✓ | ✓ | ✓ | — |
| Modifier / archiver sujet | ✓ | ✓ | ✓ | — |
| Lancer revue | ✓ | ✓ | ✓ | — |
| Cocher / commenter | ✓ | ✓ | ✓ | — |
| Clôturer revue | ✓ | ✓ | ✓ | — |

- Pas de `project_members` / `subject_members` en v1.
- IDOR : sujet hors org active → **404**.
- Libellé UI injecté via `{{.Labels.Subject.*}}` (preset défaut : `sujet`).

### Routes sujets v1

| Route | Contrôle |
|-------|----------|
| `GET /subjects` | Auth ; liste sujets org active |
| `GET /subjects/{id}` | Auth + `CanViewSubject` |
| `POST /subjects` | Auth + `CanCreateSubject` |
| `POST /subjects/{id}` | Auth + `CanManageSubject` |
| `GET /subjects/{id}/modeles?for_run=1` | Auth + `CanLaunchRun` |
| `POST /subjects/{id}/revues` | Auth + `CanLaunchRun` |

---

## Modèle cible équipes (épique access-teams — **non implémenté**)

Les sections ci-dessous décrivent le modèle **futur** sur `subjects` (plus de `projects`).  
Spec : [access-teams-epic.md](./issues/access-teams-epic.md).  
**En vigueur aujourd'hui** : section « Sujets v1 » ci-dessus (accès = membre org).

## Rôles

### Globaux (`users.role`)

| Rôle | Description |
|------|-------------|
| `admin` | Tout + bypass org ; voit tous les sujets de l'org active |
| `editor` | Créer modèles, lancer revues, cocher (sujets où accès sujet) |
| `reader` | Lecture seule (plafond — ne coche pas même si rôle sujet contributor) |

### Organisation (`organization_members.role`)

| Rôle | Description |
|------|-------------|
| `owner` | Gouvernance org : équipes, whitelist, politiques, **intégrations de l'org** (SMTP, Jira, Notion, webhooks) ; **voit tous** les sujets/revues de l'org |
| `admin` | Idem `owner` sauf actions réservées owner si ajoutées ultérieurement |
| `member` | Membre org ; accès sujet via équipe, membership direct, ou invitation |

### Sujet — direct (`subject_members.role`)

| Rôle | Description |
|------|-------------|
| `lead` | Gérer membres et équipes du sujet (si politique org), tout faire sur le sujet |
| `contributor` | Cocher, commenter, lancer revues |
| `viewer` | Lecture seule sur ce sujet |

### Sujet — via équipe (`team_subject_roles.role`)

Même sémantique que `subject_members.role`. Une équipe se voit attribuer **un rôle** sur un sujet ; chaque membre de l'équipe hérite de ce rôle pour ce sujet.

---

## Chemins d'accès à un sujet

Un utilisateur accède à un sujet par **exactement l'un** des mécanismes suivants (évalués par `ResolveSubjectAccess`) :

| Chemin | Condition | Usage |
|--------|-----------|-------|
| **Global admin** | `users.role = admin` | Bypass org et sujet |
| **Org admin** | `organization_members.role ∈ {owner, admin}` dans l'org du sujet | Supervision : voit tout dans l'org |
| **Membre direct** | Ligne `subject_members` | Exception : invité, prestataire, renfort |
| **Équipe** | ∃ équipe T : user ∈ `team_members` ∧ (T, sujet) ∈ `team_subject_roles` | Cas nominal collectif |

### Hors périmètre accès

- **`subject_tags`** : classification descriptive **uniquement**. Une étiquette **ne donne jamais** d'accès.
- **`subject_domains` / `template_domains`** : matching modèles ↔ sujet **uniquement** — jamais d'accès.
- **`template_tags`** (si présents) : matching modèles seulement.

---

## Règle de composition

```
ResolveSubjectAccess(user, subject) :

  1. Si admin global → Visible=true, bypass actions

  2. Si org owner/admin (org du sujet) → Visible=true
     (actions métier soumises au rôle global, pas de bypass lead implicite)

  3. Sinon calculer :
     role_direct  = subject_members.role ou ∅
     role_teams   = max(team_subject_roles.role) pour équipes du user
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

## Sujets privés (`subjects.visibility`)

| Visibilité | Membre org sans accès direct/équipe | Org owner/admin | Admin global |
|------------|-------------------------------------|-----------------|--------------|
| `normal` | 404 | visible | visible |
| `private` | 404 | visible | visible |

Un sujet privé n'apparaît pas aux org `member` sans chemin d'accès explicite. L'org admin le voit toujours (supervision).

---

## Politiques organisation (settings)

| Clé | Défaut | Effet |
|-----|--------|-------|
| `leads_may_assign_teams` | `true` | Lead peut ajouter une équipe existante à son sujet |
| `leads_may_invite_members` | `true` | Lead peut inviter un membre direct |
| `leads_may_invite_externals` | `false` | Lead peut inviter une adresse hors org |

Seuls org `owner` / `admin` modifient ces politiques.

---

## Matrice des actions

| Action | admin global | org owner/admin | editor + lead | editor + contributor | editor + viewer | reader + * |
|--------|--------------|-----------------|---------------|----------------------|-----------------|------------|
| Admin users / SMTP / intégrations | ✓ | ✓ | — | — | — | — |
| Gérer équipes org | ✓ | ✓ | — | — | — | — |
| Créer sujet | ✓ | ✓ | ✓ | ✓ | — | — |
| Voir tous sujets org | ✓ | ✓ | — | — | — | — |
| Voir sujet (accès direct/équipe) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Gérer membres sujet | ✓ | ✓ | ✓ (lead) | — | — | — |
| Ajouter équipe au sujet | ✓ | ✓ | ✓ (lead)* | — | — | — |
| CRUD modèles | ✓ | ✓† | ✓ (membre) | — | — | — |
| Lancer revue | ✓ | ✓† | ✓ (lead/contrib) | ✓ | — | — |
| Cocher / commenter point | ✓ | ✓† | ✓ (lead/contrib) | ✓ | — | — |
| Assigner point | ✓ | ✓† | ✓ (lead) | — | — | — |
| Clôturer revue | ✓ | ✓† | ✓ (lead) | — | — | — |
| Lire revue / export CSV | ✓ | ✓ | ✓ (membre) | ✓ | ✓ | ✓ (membre) |
| Lier / créer ticket Jira | ✓ | ✓† | ✓ (lead/contrib) | ✓ | — | — |
| Config intégrations (org active) | ✓ | ✓ | — | — | — | — |

\* Si politique `leads_may_assign_teams`.  
† Org admin : seulement si rôle global `editor` minimum pour les actions d'écriture ; lecture/export sans restriction.

---

## Routes — contrôles requis

| Route | Contrôle |
|-------|----------|
| `GET /subjects` | Auth ; liste selon `ResolveSubjectAccess` ; org admin → tous sujets org |
| `GET /subjects/{id}` | Auth + `Visible` |
| `POST /subjects/{id}/teams` | Auth + lead ou org admin ; politique `leads_may_assign_teams` |
| `POST /subjects/{id}/members` | Auth + lead ou org admin ; politique `leads_may_invite_members` |
| `GET /admin/teams` | Auth + org owner/admin |
| `POST /subjects/{id}/revues` | Auth + contributor+ effectif ou admin |
| `GET /runs/{id}` | Auth + `Visible` sur sujet de la revue |
| `POST /runs/{id}/items/{itemId}` | Auth + contributor+ effectif ou admin |
| `GET /attachments/{id}` | Auth + membre sujet de la revue liée |
| `GET\|POST /admin/integrations*` · `/admin/settings/smtp` · `/admin/settings/webhooks` | Auth + org owner/admin (ou admin global) |
| `POST /admin/*` (autres) | Auth + admin global ou org admin selon route |

Toutes les routes sensibles appellent `ResolveSubjectAccess` (ou helper dérivé) — pas de rôle sujet seul.

---

## Tests obligatoires

Chaque PR `area:auth` ou `area:core` ajoute ou maintient :

```go
TestRBAC_Matrix          // table-driven : rôle global × org × équipe × direct × route → status
TestIDOR_CrossSubject    // user A n'accède pas ressources sujet B
TestIDOR_CrossOrg        // user org A n'accède pas sujet org B
TestIDOR_TeamAccess      // accès via équipe ; retrait équipe → 404
TestIDOR_OrgAdmin        // org admin voit sans membership ; member non
TestIDOR_PrivateSubject  // private invisible sauf accès explicite ou org admin
TestCSRF_MissingToken    // POST sans CSRF → 403
```

Fichiers cibles : `internal/web/rbac_test.go`, `internal/store/subject_access_test.go`.

---

## Évolution

Modifier ce fichier uniquement via PR dédiée `area:auth` avec validation produit.

Spec épique : [docs/issues/access-teams-epic.md](./issues/access-teams-epic.md).
