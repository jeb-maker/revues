# Epic — Migration complète vers `@jeb-maker/mb` (broccoli)

Objectif : faire du design system `@jeb-maker/mb` (v0.3.0, vendoré sous
`web/static/vendor/jeb-maker-mb/`) la source unique des composants UI de Revues,
partout où un composant mb existe et où le rendu serveur progressif n'est pas dégradé.

Référence : inventaire des écarts `.cursor/skills/revues-ui-audit/broccoli-gap-inventory.md`,
décisions `.cursor/skills/revues-ui-audit/decisions.md`, gaps upstream
[miniature-broccoli#19](https://github.com/jeb-maker/miniature-broccoli/issues/19) (fermés en 0.3.0).

## Vagues

| Vague | Périmètre | Statut |
|-------|-----------|--------|
| 1 — Pilotes | login, /signaler, /revues (empty-state, segmented, pagination), progress fiche revue | ✓ |
| 2 — HTMX & grille | `htmx.js` : `FormData(form)` natif (FACE), écoute `mb-change`, dispatch `HX-Trigger` ; grille points : `mb-select` / `mb-textarea` ; `mb-toast` | ✓ |
| 3 — Boutons & flashes | `mb-button`, `mb-alert`, `mb-badge` | ✓ |
| 4 — Formulaires | `mb-input` / `textarea` / `select` / `checkbox` / `radio-group` | ✓ |
| 5 — Empty states & segmented | mes-taches, sujets, modèles | ✓ |
| 6 — Purge CSS | classes hôtes mortes ; budgets | ✓ |
| 7 — Vendor 0.3.0 + shell | `mb-card`, `mb-tag`, `mb-breadcrumbs`, `mb-nav`/`mb-nav-toggle`, `mb-avatar`, `mb-spinner`, `mb-toolbar`, `mb-select` placeholder | cette PR |

## Reste host (assumé, avec raison)

| Pattern | Raison |
|---------|--------|
| `.data-table` | décision gap-inventory : « host CSS forever » (tables + lignes HTMX) |
| formulaire noscript `/signaler` | contrôles natifs obligatoires sans JS (CE non upgradés) |
| `hx-confirm` natif | décision existante : pas de `mb-modal` pour les confirms |

## Bugs corrigés en chemin (découverts au test manuel)

- **Valeur FACE en retard d'une microtâche** : les `mb-*` synchronisent leur valeur
  de soumission (`ElementInternals.setFormValue`) dans le `updated()` asynchrone de Lit,
  APRÈS l'événement `mb-change`. `htmx.js` diffère désormais la requête d'une macrotâche,
  sinon le POST partait avec l'ancienne valeur (statut/assignation non persistés).
- **Submission value perdue au déplacement DOM** : déplacer une ligne de l'éditeur
  (`insertBefore`) déconnecte/reconnecte les custom elements et perd leur valeur de
  soumission → erreur « lignes incohérentes ». `template-editor.js` force un
  `requestUpdate('value')` sur les champs de la ligne déplacée.
- **`hx-target="closest …"`** : support `closest` ajouté au mini-client HTMX.
- **Redirections 303 sur requêtes HTMX** (upload PJ) : `resp.redirected` déclenche
  une navigation pleine page.
- **HX-Trigger ignoré** : dispatch client des événements serveur (toasts).
- **Upload multipart HTMX** : `FormData` natif quand `enctype` multipart.

## Notes d'implémentation

- Anti-FOUC 0.3.0 : les primitives de layout (`mb-card`, `mb-nav`, `mb-toolbar`, …)
  restent visibles avant `customElements.define` ; seuls les contrôles interactifs
  sont masqués (opt-in layout via `.mb-fouc`).
- `mb-select` : `placeholder` ou `<option value="">Libellé</option>` pour filtres
  « Tous / Toutes / Aller à… ».
- Assignation grille : `assignee_id=0` = « Non assigné ».
