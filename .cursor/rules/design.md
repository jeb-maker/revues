---
description: Charte design Revues
alwaysApply: true
---

# Design Revues

## Invariants

- Esprit **Basecamp** : lisible, accessible, hiérarchie typographique, chrome minimal.
- Composants UI : préférer `mb-*` (card, button, alert, badge, toolbar, breadcrumbs, input/select/textarea) ; CSS hôte pour layout (`.data-table`, `.stack-form`, `.field-hint`/`.field-error`, `.table-scroll`, `.page-header`).
- Un seul bouton primaire plein par écran ; destructif = variante danger + `confirm()` ; pas d'info essentielle en `placeholder` (utiliser `.field-hint`).
- Budgets éco : CSS core ≤ 24 Ko / 8 Ko gzip ; CSS total ≤ 40 Ko / 12 Ko gzip cumulé ; JS ≤ 15 Ko ; HTML ≤ 50 Ko/page — feuilles `run.css` / `editor.css` à la demande ; pas d'animation décorative, emoji, webfont ni image décorative.
- UI **100 % en français** ; libellés via `formatItemStatus`, `formatRunStatus`, `formatRole` et `{{.Labels.*}}`.

## Accessibilité

- `aria-current` sur l'élément actif
- `aria-live` sur les mises à jour dynamiques non critiques
- `scope="col"` sur les en-têtes de tableau
- `aria-label` sur les boutons symboles sans texte visible
- `role="status"` / `role="alert"` sur les messages de retour
