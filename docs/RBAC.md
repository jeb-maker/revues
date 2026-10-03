# Matrice RBAC — Revues

Document normatif. Toute PR touchant une route doit mettre à jour la matrice dans la description PR.

## Sujets v1 (greenfield — en vigueur)

Modèle actuel après migration `subjects`.

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
- Libellé UI : preset org `ui_subject_label` (champ `Organization`, défaut `sujet`).

### Routes sujets v1 (`/api/v1`, voir `api/openapi/openapi.yaml`)

| Route | Contrôle |
|-------|----------|
| `GET /subjects` | Auth + org active ; liste sujets selon `ResolveSubjectAccess` |
| `GET /subjects/{id}` | Auth + org active + `CanViewAccess` (sinon 404) |
| `POST /subjects` | Auth + org active + `CanCreateSubject` |
| `PATCH /subjects/{id}` · `POST /subjects/{id}/archive` | Auth + `CanViewAccess` + `CanManageAccess` (sinon 404) |
| `GET /subjects/{id}/run-templates` | Auth + `CanViewAccess` |
| `POST /subjects/{id}/runs` | Auth + `CanViewAccess` + `CanLaunchAccess` |

---

## Modèle équipes

Les sections ci-dessous décrivent le modèle équipes / grants sur `subjects`.  
**Accès org-only (SimpleUI / mono-membre)** : section « Sujets v1 » ci-dessus.

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
| `owner` | Gouvernance org : équipes, whitelist, politiques, **intégrations de l'org** (SMTP, Jira, webhooks) ; **voit tous** les sujets/revues de l'org |
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
| `normal` (sans grants) | visible via legacy `org_member_legacy` | visible | visible |
| `normal` (avec grants) | 404 sauf grant | visible | visible |
| `private` | 404 sauf grant | visible | visible |

Un sujet privé n'apparaît **jamais** via le chemin legacy ungated : il exige un accès direct/équipe, ou la supervision org/global admin.

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

Paths réels de `api/openapi/openapi.yaml` (préfixe `/api/v1`). Helpers : `ensureSubjectAccess` / `ensureRunAccess` (`internal/api/v1`) chargent la ressource dans l'org active, appellent `ResolveSubjectAccess` et renvoient **404** si non visible ; `requireOrgAdmin` = org owner/admin ou admin global + org active (403 sinon). Toute mutation passe par le middleware CSRF (`X-CSRF-Token`) ; `requireUser` → 401, `requireOrg` → 403 `org_required`.

| Route | Contrôle |
|-------|----------|
| `GET /bootstrap` · `GET /me` · `POST /auth/login\|register\|logout` | Session cookie ; CSRF sur les POST ; whitelist (`REVUES_LOGIN_REQUIRE_WHITELIST`) sur register |
| `GET\|POST /orgs` · `POST /orgs/active` · `POST /orgs/invitations/{id}/accept` | Auth ; appartenance org vérifiée côté service (`features/organizations`) |
| `GET /subjects` · `POST /subjects` | Auth + org active ; `CanCreateSubject` pour POST |
| `GET /subjects/{id}` · `GET /subjects/{id}/run-templates` · `GET /subjects/{id}/runs` · `GET /subjects/{id}/members` | Auth + `CanViewAccess` (404 sinon) |
| `PATCH /subjects/{id}` · `POST /subjects/{id}/archive` | `ensureSubjectAccess` + `CanManageAccess` ; visibilité via `CanSetSubjectVisibility` |
| `POST /subjects/{id}/members` · `DELETE /subjects/{id}/members/{userId}` | `ensureSubjectAccess` + `CanManageSubjectMembers` (org admin, ou lead selon politiques `leads_may_invite_*`) ; invitation d'un externe : `CanInviteSubjectMember` |
| `POST /subjects/{id}/runs` | `ensureSubjectAccess` + `CanLaunchAccess` |
| `GET /runs` | Auth + org active ; filtrage par accès sujet côté store (`ListFilteredRunSummaries`) |
| `GET /me/tasks` | Auth + org active ; points assignés à l'utilisateur courant uniquement |
| `GET /search` | Auth + org active ; agrège listes (sujets / revues / modèles / tâches) avec le même filtrage RBAC ; `template` si editor+ ; `task` si org ≥ 2 membres |
| `GET /runs/{id}` · `GET /runs/{id}/items/{itemId}` | `ensureRunAccess` (`CanViewAccess` sur le sujet de la revue) |
| `PATCH /runs/{id}/items/{itemId}` | `ensureRunAccess` + `CanUpdateAccess` (statut / commentaire) ; `CanAssignAccess` si `assigned_to` |
| `POST /runs/{id}/complete` | `ensureRunAccess` + `CanCompleteAccess` |
| `GET\|PUT\|POST /runs/{id}/items/{itemId}/jira` | `ensureRunAccess` + `CanLinkJiraAccess` |
| `GET\|POST /runs/{id}/items/{itemId}/attachments` · `GET /runs/{id}/items/{itemId}/attachments/{attachmentId}` | `ensureRunAccess` ; upload : `CanUpdateAccess` + validation magic bytes / taille |
| `GET /templates*` | Auth + org active |
| `POST\|PUT\|DELETE /templates*` | Auth + `CanManageGlobal` (editor+ / org admin) ; versions publiées immuables |
| `/admin/allowed-emails*` · `/admin/members*` · `/admin/invitations*` · `/admin/teams*` · `/admin/settings/policies` | `requireOrgAdmin` ; invitations : rôle `owner` réservé aux owners (ou admin global) |
| `/admin/settings/smtp*` · `/admin/integrations*` · `/admin/webhooks*` (config, test, deliveries, drain, retry) | `requireOrgAdmin` ; URLs sortantes validées anti-SSRF |

Toutes les routes sensibles appellent `ResolveSubjectAccess` (ou helper dérivé) — pas de rôle sujet seul. Les routes `/subjects/{id}/teams` (grant équipe ↔ sujet) n'existent pas dans l'API v1 : `GrantTeamSubjectRole` n'est accessible que par le store.

---

## Tests obligatoires

Chaque PR `area:auth` ou `area:core` maintient (ou étend) les tests existants :

| Exigence | Tests existants |
|----------|-----------------|
| Composition d'accès (global / org admin / direct / équipe, max des rôles, retrait équipe → invisible) | `TestResolveSubjectAccess`, `TestResolveSubjectAccess_PrivateNoLegacy`, `TestListSubjects_PrivateFilter` — `internal/store/subject_access_test.go` |
| Règles pures (org admin reader, lead sans bypass, politiques invitations, visibilité) | `TestCanManageOrgUsers`, `TestCanContributeAccess_OrgAdmin`, `TestCanLeadAccess_NoOrgAdminBypass`, `TestCanManageAccess_OrgAdminReaderDenied`, `TestCanSetSubjectVisibility`, `TestCanInviteSubjectMember` — `internal/features/subjects/service_test.go` ; `TestCanManageOrgUsers` — `internal/web/middleware/org_admin_test.go` |
| IDOR sujet / org / privé | `TestIDOR_CrossSubject` (`internal/features/subjects/idor_test.go`), `TestSubjectsAPI_IDOR_CrossOrg`, `TestSubjectsAPI_IDOR_PrivateSubject`, `TestSubjectsAPI_ReaderCannotCreate` — `internal/api/v1/subjects_test.go` |
| IDOR revue / point + CSRF | `TestRunsAPI_IDORAndCSRF`, `TestRunsAPI_RequiresAuth`, `TestRunsAPI_LaunchSnapshotUpdateCompleteAndGuards` — `internal/api/v1/runs_test.go` |
| Pièces jointes (IDOR, type, taille) | `TestAttachments_IDOR_CrossUser`, `TestAttachments_UploadDownloadSecurity` — `internal/api/v1/attachments_test.go` ; `TestProcessUpload_Rejects*` — `internal/attachments/process_test.go` |
| CSRF / session / whitelist auth | `TestAuthAPI_BootstrapGuestCSRF`, `TestAuthAPI_LoginRequiresCSRF`, `TestAuthAPI_LogoutRequiresCSRF`, `TestAuthAPI_MeRequiresSession`, `TestAuthAPI_RegisterWhitelistReject` — `internal/api/v1/auth_test.go` ; `TestTemplatesAPI_RequiresAuthAndCSRF` |
| Admin org (owner/admin vs member) | `TestAdminAPI_RBACAndParity`, `TestAdminIntegrations_RequiresOrgAdmin`, `TestAdminSMTP_OrgOwnerAllowed`, `TestAdminSMTP_MaskedPasswordAndOrgAdmin`, `TestAdminJira_MaskedTokenTestAndRBAC`, `TestAdminWebhooks_ConfigHMACAndSSRF` — `internal/api/v1/admin*_test.go` |
| Anti-SSRF webhooks / Jira | `TestWebhook_SSRF_Block`, `TestWebhook_SSRF_BlockPrivateDial`, `TestDispatcher_Drain_ReChecksSSRF` — `internal/integrations/webhooks/dispatcher_test.go` ; `TestAdminJira_RejectsPrivateURL` |

Manque connu (à ajouter dans une PR `area:auth` dédiée) : une matrice table-driven globale **rôle global × rôle org × accès sujet × route → status** couvrant toutes les lignes de la section « Matrice des actions ».

Toute nouvelle route documente sa ligne dans le tableau « Routes — contrôles requis » et ajoute un test IDOR/CSRF dans `internal/api/v1/<domaine>_test.go`.

---

## Évolution

Modifier ce fichier uniquement via PR dédiée `area:auth` avec validation produit.
