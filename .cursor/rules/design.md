---
description: Charte design Revues
alwaysApply: true
---

# Design Revues

Stack UI : SvelteKit (SPA) + Web Components `@jeb-maker/mb` (voir `docs/FRONTEND.md`). Pages dans `frontend/src/routes/**/*.svelte`, styles dans les blocs `<style>` Svelte.

## Invariants

- Esprit **Basecamp** : lisible, accessible, hiérarchie typographique, chrome minimal.
- Composants UI : préférer `mb-*` (button, alert, badge, input/textarea, tag, empty-state, spinner, **table**, **card**) avant du HTML natif restylé ; CSS hôte pour le layout seulement, sur tokens `var(--mb-*)` (jamais de hex en dur). Classes hôte partagées (feuille commune `frontend/src/lib/styles/app.css`, importée par le layout racine) : `.page`, `.page-header`, `.page-header__row`, `.page-header__actions`, `.lede`, `.muted`, `.actions`, `.filters`, `.filters__count`, `.crumbs`, `.section`, `.stack-form`, `.card-stack`, `.field-hint` / `.field-error`, `.table-scroll`, `.row-list`, `.card-list`, `.page--auth`, `.auth-card`. Un `<table>` natif n'est admis que pour l'éditeur de modèle (`TemplateEditor.svelte`, DnD). Listes : CTA face au H1 ; fiches : contenu en `mb-card`, actions en bas.
- Un seul bouton primaire plein par écran ; destructif = `variant="danger"` + `confirm()` ; pas d'info essentielle en `placeholder` (utiliser un `hint` / texte d'aide visible).
- Budgets éco : seuils JS/CSS **app** appliqués par `./scripts/check.sh` (détail dans `docs/PLAN.md`) ; vendor mb mesuré hors fail ; pas d'animation décorative, emoji, webfont ni image décorative.
- Icônes d'action : SVG inline (classe `.icon`), pas de lib npm ; CTA = icône + libellé ; row-actions denses = `.row-action` / `icon-only` + `aria-label` (voir decisions).
- UI **100 % en français** ; statuts, rôles et libellés métier passent par des helpers de libellés côté front (`frontend/src/lib/**`), jamais de code brut (`pending`, `nok`, `owner`…) affiché à l'utilisateur.

## Accessibilité

- `aria-current` sur l'élément actif
- `aria-live` sur les mises à jour dynamiques non critiques
- `scope="col"` sur les en-têtes de tableau natif
- `aria-label` sur les boutons symboles sans texte visible
- `role="status"` / `role="alert"` sur les messages de retour (ou `mb-alert`)
- Les `mb-input` n'héritent pas du `<label>` hôte (shadow DOM) : fournir `label=` ou `aria-label`
