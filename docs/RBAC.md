# Matrice RBAC — Revues

Document normatif. Toute PR touchant une route doit mettre à jour la matrice dans la description PR.

## Modèle cible (#310)

```
Compte
  └─ Organisation (appartenance + gouvernance)
       └─ Projet / sujet (droits métier)
```

Héritage **fixe** : org `owner` / `admin` voient tous les projets et peuvent écrire (contributor+) ;
org `member` n’a accès qu’avec un grant `subject_members` (ou équipe icebox).
Pas d’admin global produit ; `users.role` est un vestige technique (défaut `editor` au login).

## Projets (sujets)

| Entité | Accès |
|--------|----------|
| **Projet** | Org owner/admin **ou** grant `subject_members` / équipe |
| **Domaines** (`subject_domains`, `template_domains`) | Matching modèles ↔ sujet — **jamais** d'accès |

### Permissions projet (`internal/features/subjects/service.go`)

| Action | org owner/admin | lead | contributor | viewer | membre org sans grant |
|--------|-----------------|------|-------------|--------|------------------------|
| Voir projet | ✓ | ✓ | ✓ | ✓ | — |
| Créer projet | ✓ (membre org) | ✓ | ✓ | ✓ | ✓ (membre org) |
| Modifier / archiver | ✓ | ✓ | — | — | — |
| Lancer revue / cocher | ✓ | ✓ | ✓ | — | — |
| Gérer membres projet | ✓ | ✓* | — | — | — |

\* Selon politiques org `leads_may_invite_*`.

- Créateur d’un projet → `subject_members.role = lead` automatique.
- IDOR : projet hors org active ou sans accès → **404**.
- Libellé UI : preset org `ui_subject_label` (défaut `projet`).

### Routes projets (`/api/v1`)

| Route | Contrôle |
|-------|----------|
| `GET /subjects` | Auth + org active ; liste selon `ResolveSubjectAccess` |
| `GET /subjects/{id}` | Auth + org active + `CanViewAccess` (sinon 404) |
| `POST /subjects` | Auth + org active + membre org |
| `PATCH /subjects/{id}` · `POST /subjects/{id}/archive` | `CanManageAccess` |
| `GET /subjects/{id}/run-templates` | `CanViewAccess` |
| `POST /subjects/{id}/runs` | `CanContributeAccess` |
| `POST /subjects/{id}/members` | `CanInviteSubjectMember` |

---

## Modèle équipes (icebox)

Parcours produit nominal : **affectation personne → projet** (`subject_members`). Les équipes org /
`team_subject_roles` restent en base et dans `ResolveSubjectAccess`, mais **hors UI nominale**
(masquées admin, #295).

## Rôles

### Organisation (`organization_members.role`)

| Rôle | Description |
|------|-------------|
| `owner` | Gouvernance : membres, invitations, politiques, **intégrations** (SMTP, Jira, Confluence, webhooks) ; voit tous les projets |
| `admin` | Idem `owner` sauf réserves owner ultérieures |
| `member` | Dans l’org ; **pas** d’accès auto aux projets |

### Projet — direct (`subject_members.role`)

| Rôle | Description |
|------|-------------|
| `lead` | Chef de projet : membres du projet, tout faire sur le projet |
| `contributor` | Cocher, commenter, lancer revues |
| `viewer` | Lecture seule sur ce projet |

### Projet — via équipe (`team_subject_roles.role`)

Même sémantique que `subject_members.role` (icebox UI).

### Vestige `users.role`

Colonne encore présente (`admin` / `editor` / `reader`) mais **non utilisée** pour les décisions métier après #310. Login / register assignent `editor`. Bootstrap (`REVUES_BOOTSTRAP_ADMIN_EMAIL`) → `editor` + `owner` de l’org `default`. Suppression de colonne = follow-up éventuel.

---

## Chemins d'accès à un projet

Évalués par `ResolveSubjectAccess` :

| Chemin | Condition | Usage |
|--------|-----------|-------|
| **Org admin** | `organization_members.role ∈ {owner, admin}` | Supervision : voit tout dans l’org |
| **Membre direct** | Ligne `subject_members` | Parcours nominal |
| **Équipe** | user ∈ équipe ∧ `team_subject_roles` | Icebox |

### Hors périmètre accès

- **`subject_domains` / `template_domains`** : matching modèles ↔ sujet **uniquement**.

---

## Règle de composition

```
ResolveSubjectAccess(user, subject) :

  1. Si org owner/admin (org du sujet) → Visible=true
     (écriture métier via CanContributeAccess ; lead implicite non)

  2. Sinon calculer :
     role_direct  = subject_members.role ou ∅
     role_teams   = max(team_subject_roles.role) pour équipes du user
     role_effectif = max(role_direct, role_teams)   // lead > contributor > viewer

  3. Visible = (role_effectif ≠ ∅) OU org admin

  4. Action :
     - Org admin : lecture + contribution (cocher / lancer)
     - Sinon rôle projet suffisant
     - Assign / manage members : lead (ou org admin pour invite)

  5. Si ¬Visible → 404 (pas 403)
```

### Ordre des rôles (max)

```
lead > contributor > viewer
```

---

## Auth / invitations

| Route | Contrôle |
|-------|----------|
| `GET /bootstrap` · `GET /me` · `POST /auth/login\|register\|logout` | Session cookie ; CSRF sur les POST ; invite-only (`REVUES_LOGIN_REQUIRE_WHITELIST`) sur register |
| `POST /admin/invitations*` | `requireOrgAdmin` ; rôle invite `owner` réservé aux owners |
| `POST /orgs/invitations/{id}/accept` | Session ; email doit matcher |

Gate register/OAuth strict : membre org **ou** invitation pending **ou** bootstrap. Self-service (flag off) : compte + `/org/new`.

---

## Administration org

| Route | Contrôle |
|-------|----------|
| `/admin/members*` · `/admin/invitations*` · `/admin/teams*` · `/admin/settings/policies` | `requireOrgAdmin` (owner/admin org) |
| `/admin/settings/smtp` · `/admin/integrations*` · webhooks | `requireOrgAdmin` |
| `GET\|POST /runs/{runId}/confluence` | Accès sujet ; publish = contributeur+ (`CanPublishConfluenceAccess`) |

---

## Tests de référence

| Zone | Tests |
|------|-------|
| Accès sujet | `internal/store/subject_access_test.go` · `internal/features/subjects/*_test.go` |
| CSRF / session / invite-only | `internal/api/v1/auth_test.go` |
| Admin org | `internal/api/v1/admin_test.go` · `admin_invitations_test.go` |

Manque connu : matrice table-driven globale **rôle org × rôle projet × route → status** (follow-up `area:auth`).
