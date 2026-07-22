# Epic — Migration complète vers `@jeb-maker/mb` (broccoli)

Objectif : faire du design system `@jeb-maker/mb` (v0.2.0, vendoré sous
`web/static/vendor/jeb-maker-mb/`) la source unique des composants UI de Revues,
partout où un composant mb existe et où le rendu serveur progressif n'est pas dégradé.

Référence : inventaire des écarts `.cursor/skills/revues-ui-audit/broccoli-gap-inventory.md`,
décisions `.cursor/skills/revues-ui-audit/decisions.md`, gaps upstream
[miniature-broccoli#19](https://github.com/jeb-maker/miniature-broccoli/issues/19).

## Vagues

| Vague | Périmètre | Statut |
|-------|-----------|--------|
| 1 — Pilotes | login, /signaler, /revues (empty-state, segmented, pagination), progress fiche revue | ✓ mergée |
| 2 — HTMX & grille | `htmx.js` : `FormData(form)` natif (FACE), écoute `mb-change`, dispatch `HX-Trigger` ; grille points : `mb-select` statut/assigné (sentinelle `0` = non assigné), `mb-textarea` commentaire ; `mb-toast` | cette PR |
| 3 — Boutons & flashes | `.button*` → `mb-button` (variants primary/secondary/ghost/danger, `size="sm"`, `icon-only`, `href`) ; flashes `.success`/`.error` → `mb-alert` ; badges statut/rôle → `mb-badge` (helper `badgeVariant`) | cette PR |
| 4 — Formulaires | `mb-input` (text/search/email/password/url/number/file), `mb-textarea`, `mb-select` (options slottées), `mb-checkbox`, `mb-radio-group` sur toutes les pages formulaire + toolbars recherche | cette PR |
| 5 — Empty states & segmented restants | `mb-empty-state` + `mb-segmented-control` sur mes-taches, sujets, modèles | cette PR |
| 6 — Purge CSS | retirer de `app.css`/`run.css`/`editor.css` les classes devenues mortes ; vérifier budgets | cette PR |

## Reste host (assumé, avec raison)

| Pattern | Raison |
|---------|--------|
| `.card` sections structurelles | anti-FOUC `mb-card:not(:defined)` masquerait tout le contenu tant que `mb-boot.js` n'est pas exécuté ; contraire au rendu progressif éco. À revisiter si mb livre un mode SSR/DSD. |
| `.data-table` | décision gap-inventory : « host CSS forever » (tables + lignes HTMX) |
| nav / breadcrumb / toolbar layout / tags / avatar / timeline / spinner | pas de composant mb en 0.2.0 — gaps upstream #19 |
| selects à option vide signifiante (filtres « Tous/Toutes », switcher org, jump admin) | `mb-select` 0.2.0 ne supporte pas de label sur l'option vide (placeholder) — gap upstream |
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
- **`hx-target="closest .card"` jamais supporté** par le mini-client (sélecteur invalide
  silencieux, préexistant) — support `closest` ajouté.
- **Redirections 303 sur requêtes HTMX** (upload PJ) : le client suivait la redirection
  et swappait la page entière dans la carte — `resp.redirected` déclenche maintenant
  une navigation.
- **HX-Trigger ignoré** (préexistant) : les toasts serveur (`toast:success`) n'étaient
  jamais déclenchés côté client — le client dispatch maintenant les événements HX-Trigger,
  relayés vers `mb-toast`.
- **Upload multipart HTMX** (préexistant) : le client sérialisait tout en URL-encoded ;
  il envoie maintenant le `FormData` natif quand `hx-encoding`/`enctype` est multipart.

## Notes d'implémentation

- Les composants mb sont form-associated (ElementInternals) : soumission native
  (POST classique et GET filtres) et validation `required` fonctionnent sans JS applicatif.
- `htmx.js` collecte désormais les champs via `new FormData(form)` (couvre les FACE),
  écoute `mb-change` (composed) au niveau document, et dispatch les événements
  `HX-Trigger` renvoyés par le serveur (répare les toasts, jamais câblés côté client).
- Assignation grille : `assignee_id=0` = « Non assigné » (l'option vide du composant
  reste sélectionnable et équivaut aussi à une désassignation).
- Sauvegarde commentaire : `mb-change` (change natif au blur) remplace le trigger
  `blur` historique qui ne remontait pas au document (non-bubbling) — bug corrigé de fait.
