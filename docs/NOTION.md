# Notion — mapping des champs

Documentation de la configuration Notion admin et du mapping prévu pour export/import (issues #26, #27).

## Configuration admin (`/admin/integrations/notion`)

Stockage chiffré dans la table `integrations` (`type = 'notion'`), même schéma que Jira.

### Payload JSON (déchiffré)

| Champ JSON | Formulaire admin | Obligatoire | Description |
|------------|------------------|-------------|-------------|
| `api_token` | Jeton d'intégration Notion | Oui | Token d'intégration interne (`secret_…`) ou jeton d'accès personnel (`ntn_…`). Jamais loggé. |
| `workspace_name` | Nom du workspace (libellé) | Non | Libellé admin pour identifier le workspace cible. |
| `default_database_id` | ID base Notion par défaut | Non | UUID (32 hex, avec ou sans tirets) de la base cible pour l'export de revues (#26). |

### Test de connexion

`POST /admin/integrations/notion` avec `action=test` appelle `GET https://api.notion.com/v1/users/me` :

- En-tête `Authorization: Bearer {api_token}`
- En-tête `Notion-Version: 2022-06-28`

## Mapping export revue → Notion (#26, futur)

| Revues | Propriété Notion | Type |
|--------|------------------|------|
| Titre revue | `Name` | `title` |
| Projet | `Projet` | `rich_text` / `select` |
| Date clôture | `Date` | `date` |
| Point statut | colonne | `select` |
| URL revue | `Lien Revues` | `url` |

## Mapping import modèle Notion → Revues (#27, futur)

| Notion | Revues template |
|--------|-----------------|
| `Name` | Nom du modèle |
| `Section` | `section` |
| `Point` | `label` |
| `Aide` | `help_text` |
| `Requis` | `required` |

## Sécurité

- RBAC : routes `/admin/integrations/notion` réservées au rôle `admin`.
- CSRF obligatoire sur tous les POST.
- Token chiffré avec `REVUES_ENCRYPTION_KEY`.
