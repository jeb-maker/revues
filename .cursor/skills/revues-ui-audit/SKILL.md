---
name: revues-ui-audit
description: >-
  Audite l'UI/UX de Revues en passes structurées avec confirmation et critique
  adverse. Use when the user asks for a UX/UI review, incohérences interface,
  audit design, contrôle charte, passe UI, or similar quality checks on templates
  and CSS — read-only by default unless a P0 UX bug is confirmed.
---

# Audit UX/UI — Revues

Audit **read-only** par défaut. N'implémenter que si **P0 UX** confirmé (voir ci-dessous).  
`./scripts/check.sh` = gate **après** patch, pas critère d'existence d'un bug UI.

## Références (lire en premier)

1. `.cursor/rules/design.md` — charte
2. [decisions.md](decisions.md) — décisions produit (ne pas re-signaler)
3. Pages : `frontend/src/routes/**/*.svelte` (+ `+layout.svelte` / `+layout.ts` racine : header commun, session unique, garde d'auth)
4. Lib front : `frontend/src/lib/**` — `components/` (`AdminNav.svelte`, `TemplateEditor.svelte`), `api/` (client OpenAPI généré + wrappers), `auth/session.ts` (session `page.data.boot`), `auth/messages.ts` (copy FR des erreurs login), `i18n/labels.ts` (libellés FR), `styles/app.css` (thème sur tokens), `mb/` (`ensureMb()`, `svelte.d.ts`)
5. CSS : `frontend/src/lib/styles/app.css` (classes hôte partagées) + blocs `<style>` résiduels des pages ; vendor `mb-*` sous `web/static/vendor/jeb-maker-mb/` (tokens `tokens-core.css`, `mb-bridge.css`, `mb-boot.js`)
6. Contrat : `api/openapi/openapi.yaml` (capabilities `can_*` par ressource) ; règles métier et tests : `internal/features/**`, `internal/api/v1/*_test.go` — spec implicite souvent plus fiable que la charte seule

## Avant les passes

1. Lire `decisions.md` en entier (surtout progressive disclosure + terminologie).
2. **Diff decisions ↔ code** : pour chaque écart, classer **Doc drift** (mettre à jour decisions) vs **Bug produit** (constat audit). Ne pas traiter un onglet/feature acté en code mais absent de decisions comme bug UX sans vérifier.
3. Si un audit récent existe (transcript / issue) : ne pas re-lister les points déjà actés dans decisions ; noter les régressions seulement.

## P0 UX (autorise un fix immédiat)

- Parcours cassé (CTA mort, wizard bloqué, empty state sans issue)
- Violation **systémique** de charte (ex. plusieurs primaires sur le même écran métier, primaire mb sans fond)
- A11y bloquante (pas de nom accessible sur action critique, tableau sans en-têtes)
- Vocabulaire **faux** sur un écran hub (nav ≠ page)

Sinon : livrer l'audit, proposer des PR, **attendre validation**.

## Méthode — passes + statut unifié

Pour **chaque** constat : un statut parmi :

| Statut | Sens |
|--------|------|
| **Confirmé** | Bug / dette réelle, preuve fichier |
| **Partiel** | Vrai écart, impact limité ou fix ambigu |
| **Doc drift** | Code OK / décisions périmées → proposer MàJ `decisions.md` |
| **Choix intentionnel** | Aligné decisions ou Basecamp ; ne pas top-10 |
| **Rejeté** | Faux positif |

Critique adverse **intégrée** : une ligne « pourquoi ça tient / downgrade » par point **Confirmé** ou **Partiel**.

### Passe 1 — Libellés & i18n

Checklist :

- [ ] Statuts / rôles affichés via des helpers de libellés front (pas de codes bruts `pending`, `nok`, `owner`…)
- [ ] Vocabulaire **Revues / Sujets / Modèles** cohérent nav, H1, empty states, CTA, confirms de clôture (marque produit « Revues » OK)
- [ ] Presets org `ui_run_label` / `ui_subject_label` (schéma `Organization`) : vérifier s'ils sont consommés par l'écran audité avant de signaler un écart
- [ ] UI 100 % FR (pas leads / Issue / direct non traduit)
- [ ] Info essentielle en texte d'aide visible, pas placeholder seul
- [ ] Domaines = matching modèles (pas d’étiquettes descriptives — retirées)

### Passe 2 — Composants

Checklist :

- [ ] Préférer `mb-*` (button, alert, badge, input, textarea, tag, empty-state, table…) aux `<button>` / `<input>` / `<select>` natifs restylés
- [ ] **Un seul** primaire (`mb-button variant="primary"`) **par écran** — empty states onboarding peuvent justifier plusieurs CTA
- [ ] Destructif : `variant="danger"` + `confirm()` (sauf exception actée row-action)
- [ ] Clôture revue : primaire + `confirm()`, **pas** danger
- [ ] Cohérence listes ↔ détails (badges, overdue, colonnes)
- [ ] Messages succès/erreur : `mb-alert` (`role="status"` / `role="alert"`)
- [ ] Tokens : couleurs via `var(--mb-*)` existants dans `tokens/*.css`, pas de hex en dur ni de token fantôme

### Passe 3 — Parcours

Checklist :

- [ ] Capabilities serveur (`can_launch`, `can_update_items`, `can_complete`, `can_assign`, `can_manage*`, `can_link`) : CTA masqué **ou** message explicite — jamais une erreur opaque
- [ ] Nav + hub org (`/org/select`, `/org/new`, `/admin`)
- [ ] Fil d'Ariane = ancêtres seulement ; H1 = courant ; absents sur pages racine
- [ ] CTA listes en toolbar ; formulaires : primaire en bas (danger-zone archiver = exception courante)
- [ ] Lancement revue `/subjects/{id}/launch` : choix du modèle, pas de stepper ; clic modèle = lancer
- [ ] Empty states onboarding vs « aucun résultat filtre »
- [ ] Intégrations (Jira / webhooks ; **pas** Notion) : CTA masqué **ou** message « non configuré » — pas d'erreur opaque
- [ ] Garde d'auth : redirection `/login` cohérente, pas de flash de contenu non autorisé

### Passe 4 — Accessibilité & budgets

Checklist :

- [ ] `lang="fr"` (`app.html`), `aria-current` sur actif, `scope="col"` sur `<table>` natif
- [ ] `role="status"` / `alert` (souvent via `mb-alert`)
- [ ] `mb-input` / `mb-textarea` : nom accessible (`label=` ou `aria-label`) — le `<label>` hôte ne traverse pas le shadow DOM
- [ ] Boutons icône : `aria-label` ; colonnes actions : libellé `visually-hidden` si `<th>` vide
- [ ] Budgets éco : seuils JS/CSS app de `scripts/check.sh` (voir `docs/PLAN.md`) ; signaler une marge < 10 % ou un relèvement de seuil dans la PR auditée

### Vérif live (optionnelle)

Si `REVUES_DEV_AUTH=1` / seed dispo : smoke sur `/runs`, `/subjects/{id}/launch`, fiche revue `/runs/{id}`. Sinon sources Svelte + tests API suffisent — le noter dans la synthèse.

## Hors scope (sauf demande explicite)

- Refactor architecture, RBAC, sécurité, perf
- Cosmétique pure sans impact compréhension / action (ex. espacement ±2px) — **a11y et libellés ne sont pas cosmétique**
- Implémentation large sans validation utilisateur

## Livrable

```markdown
# Audit UX/UI Revues — [date]

## Synthèse
[2–3 phrases]

## Constats par passe
| # | Passe | Constat | Statut | Preuve | Critique |
|---|-------|---------|--------|--------|----------|

## Top 10 actionnable
| Rang | Action | Effort | Priorité |

## MàJ decisions.md proposée
[Puces Doc drift à acter — ou « Aucune »]

## Décisions produit ouvertes
[Questions à trancher — ou « Aucune »]

## Plan PR suggéré
[2–3 PR max, sans implémenter]
```

## Rappels projet

- Stack : SvelteKit SPA (`frontend/`) + Web Components `@jeb-maker/mb` ; API Go JSON `/api/v1` (pas de `html/template` ni HTMX)
- Composants : `@jeb-maker/mb` (voir decisions — version vendorée)
- Si code modifié : `./scripts/check.sh` avant push

## Prompt utilisateur type

> Fais un audit UX/UI en plusieurs passes, confirme les problèmes, puis critique adverse. Pas de code sauf P0.
