# Notion — mapping des champs

Documentation de la configuration Notion admin et du mapping export/import (WP-021 / issues historiques #25–#27).

## Configuration admin

| Surface | Chemin |
|---------|--------|
| SPA | `/admin/integrations/notion` |
| API | `GET\|PUT\|DELETE /api/v1/admin/integrations/notion` |
| Test | `POST /api/v1/admin/integrations/notion/test` |

Stockage chiffré dans la table `integrations` (`type = 'notion'`), même schéma que Jira. RequireOrgAdmin + CSRF sur mutations.

### Payload JSON (déchiffré)

| Champ JSON | Formulaire admin | Obligatoire | Description |
|------------|------------------|-------------|-------------|
| `api_token` | Jeton d'intégration Notion | Oui | Token d'intégration interne (`secret_…`) ou jeton d'accès personnel (`ntn_…`). Jamais loggé ni renvoyé (`has_api_token`). |
| `workspace_name` | Nom du workspace (libellé) | Non | Libellé admin pour identifier le workspace cible. |
| `default_database_id` | ID base Notion par défaut | Non | UUID (32 hex, avec ou sans tirets) de la base cible pour l'export de revues. |

### Test de connexion

`POST /api/v1/admin/integrations/notion/test` appelle `GET https://api.notion.com/v1/users/me` :

- En-tête `Authorization: Bearer {api_token}`
- En-tête `Notion-Version: 2022-06-28`

## Export revue → Notion

| Surface | Chemin |
|---------|--------|
| SPA | fiche revue `/runs/{id}` (bouton si `can_export_notion`) |
| API | `POST /api/v1/runs/{runId}/notion-export` |

RBAC : `CanCompleteAccess` + revue `done`. 409 si déjà exportée. Prérequis : `ExportReady` (token + `default_database_id`).

| Revues | Propriété Notion | Type |
|--------|------------------|------|
| Titre revue | `Name` | `title` |
| Projet | `Projet` | `rich_text` / `select` |
| Date clôture | `Date` | `date` |
| Point statut | colonne | `select` |
| URL revue | `Lien Revues` | `url` |

## Import modèle Notion → Revues

| Surface | Chemin |
|---------|--------|
| SPA | `/modeles/notion-import` |
| API | `POST /api/v1/templates/notion-import` |

Actions : `fetch` → `preview` → `import` (crée un nouveau modèle, version 1). RBAC : `CanManageGlobal` (editor+).

Chemin collection (pas `{templateId}`) : l'import crée un modèle ; équivalent documenté de `/templates/{templateId}/notion-import`.

| Notion | Revues template |
|--------|-----------------|
| `Name` | Nom du modèle / libellé point |
| `Section` | `section` |
| `Point` | `label` |
| `Aide` | `help_text` |
| `Requis` | `required` |

## Sécurité

- RBAC : config admin réservée aux org owner/admin (ou admin global) ; import editor+ ; export CanCompleteAccess + IDOR sujet.
- CSRF obligatoire sur tous les POST/PUT/DELETE.
- Token chiffré avec `REVUES_ENCRYPTION_KEY`.
