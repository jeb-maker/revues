# Notion — retiré

L’intégration Notion (import de modèles, export de revues, configuration admin) a été **retirée du produit** : plus d’API `/api/v1`, plus d’UI SPA, plus de package `internal/integrations/notion`.

Les lignes `integrations` de type `notion` et la colonne `checklist_runs.notion_url` peuvent subsister en base pour des données historiques ; l’application ne les lit ni ne les écrit.
