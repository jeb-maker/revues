# Matrice RBAC — Revues

Document normatif. Toute PR touchant une route doit mettre à jour la matrice dans la description PR.

## Rôles

### Globaux (`users.role`)

| Rôle | Description |
|------|-------------|
| `admin` | Tout + admin système (users, SMTP, intégrations) |
| `editor` | Créer modèles, lancer revues, cocher (tous projets où membre) |
| `reader` | Lecture seule |

### Locaux (`project_members.role`)

| Rôle | Description |
|------|-------------|
| `lead` | Gérer membres du projet, tout faire sur le projet |
| `contributor` | Cocher, commenter, lancer revues |
| `viewer` | Lecture seule sur ce projet |

## Règle de composition

```
admin     → bypass contrôles projet (sauf impersonation)
editor    → nécessite project_members pour actions projet
reader    → lecture si membre viewer+ OU admin

Action projet :
  1. Utilisateur authentifié
  2. Si admin → autorisé
  3. Sinon → membre du projet avec rôle local suffisant
  4. Sinon → 404 (pas 403, pour ne pas révéler l'existence)
```

## Matrice des actions

| Action | admin | editor + lead | editor + contributor | editor + viewer | reader + * |
|--------|-------|---------------|----------------------|-----------------|------------|
| Admin users/SMTP/intégrations | ✓ | — | — | — | — |
| Créer projet | ✓ | ✓ | ✓ | — | — |
| Gérer membres projet | ✓ | ✓ (lead) | — | — | — |
| CRUD modèles | ✓ | ✓ (membre) | — | — | — |
| Lancer revue | ✓ | ✓ (lead/contrib) | ✓ | — | — |
| Cocher / commenter point | ✓ | ✓ (lead/contrib) | ✓ | — | — |
| Assigner point | ✓ | ✓ (lead) | — | — | — |
| Clôturer revue | ✓ | ✓ (lead) | — | — | — |
| Lire revue / projet | ✓ | ✓ (membre) | ✓ | ✓ | ✓ (membre) |
| Export CSV | ✓ | ✓ (membre) | ✓ | ✓ | ✓ (membre) |
| Lier / créer ticket Jira | ✓ | ✓ (lead/contrib) | ✓ | — | — |
| Config intégrations | ✓ | — | — | — | — |

## Routes — contrôles requis

| Route | Contrôle |
|-------|----------|
| `GET /projects` | Auth ; liste filtrée par appartenance (sauf admin) |
| `GET /projects/{id}` | Auth + membre ou admin |
| `POST /projects/{id}/runs` | Auth + lead/contributor ou admin |
| `GET /runs/{id}` | Auth + membre projet ou admin |
| `PATCH /runs/{id}/items/{itemId}` | Auth + contributor+ ou admin |
| `GET /attachments/{id}` | Auth + membre projet de la revue liée |
| `POST /admin/*` | Auth + admin |

## Tests obligatoires

Chaque PR `area:auth` ou `area:core` ajoute ou maintient :

```go
// TestRBAC_Matrix — table-driven : role × route → status
// TestIDOR_CrossProject — user A n'accède pas ressources projet B
```

Fichier cible : `internal/web/rbac_test.go` (créé avec issue auth).

## Évolution

Modifier ce fichier uniquement via PR dédiée `area:auth` avec validation produit.
