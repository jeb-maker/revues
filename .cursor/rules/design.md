---
description: Charte design Revues
alwaysApply: true
---

# Design Revues

Référence normative : [docs/DESIGN.md](../../docs/DESIGN.md).

## Invariants

- Esprit **Basecamp** : lisible, accessible, hiérarchie typographique, chrome minimal.
- Réutiliser les composants CSS existants : `.button*`, `.data-table`, `.status-badge`, `.card`, `.stack-form`, `.field-hint`/`.field-error`, `.inline-actions`, `.table-scroll`, `.pagination`, `.empty-state`, `.warning-panel`.
- Un seul `.button` plein par écran ; destructif = `.button-danger` + `confirm()` ; pas d'info essentielle en `placeholder` (utiliser `.field-hint`).
- Budgets éco : CSS core ≤ 24 Ko / 8 Ko gzip ; CSS total ≤ 40 Ko / 12 Ko gzip cumulé ; JS ≤ 15 Ko ; HTML ≤ 50 Ko/page — feuilles `run.css` / `editor.css` à la demande ; pas d'animation décorative, emoji, webfont ni image décorative.
- UI **100 % en français** ; libellés via `formatItemStatus`, `formatRunStatus`, `formatRole`.
- Banc d'essai : `/styleguide` (admin).

## Accessibilité

- `aria-current` sur l'élément actif
- `aria-live` sur les mises à jour dynamiques non critiques
- `scope="col"` sur les en-têtes de tableau
- `aria-label` sur les boutons symboles sans texte visible
- `role="status"` / `role="alert"` sur les messages de retour
