# Audit parcours multi-personas — synthèse live (2026-07-21)

## Synthèse

Cœur métier et caps tiennent (nav paliers, écritures lecteur refusées, statut HTMX + fragment `<tr>`). Dette de présence : dead-ends collab (équipes), breadcrumbs Intégrations, copy admin/Jira, fuites deep-link `/modeles` lecteur, hardcodes « revue ».

## Fixes appliqués dans cette vague

| Issue | Fix |
|-------|-----|
| Ligne run disparaît après statut/assign | `htmx.js` parse via `<template>` |
| « Créer une équipe » → 403 (éditeur) | Copy honnête + lien seulement si `CanManageOrgUsers` |
| Copy « toutes affectées » si 0 équipes | Message « Aucune équipe… » |
| BC Intégrations « Admin » | Parent **Organisation** → `/admin` |
| Jira « contactez un admin » pour admin | Lien **Configurer Jira** si `CanManageIntegrations` |

## Reporté (décision / PR UI)

- Hardcodes revue/modèle vs `ui_run_label`
- Lecteur `GET /modeles` sans nav
- DisplayRole « Contributeur » vs pouvoirs lead
- Confirm clôture sans signal points restants
- POST assign avec `ShowAssign=false` (disclosure vs deny)
- CSV/preuve P0

## Livrables agents

- `critique-particulier-live.md`
- `critique-alice-live.md`
- `critique-claire-live.md`
- `critique-admin-live.md`
