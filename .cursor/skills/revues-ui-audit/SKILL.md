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
3. `web/templates/**/*.html`
4. CSS : `web/static/css/app.css` + `run.css` / `editor.css` à la demande ; vendor `mb-*` sous `web/static/vendor/jeb-maker-mb/`
5. `internal/web/templates/` — `templates.go` (`FormatItemStatus`, `FormatRunStatus`, `FormatRole`, `FormatAccessSource`), `labels.go`, `breadcrumbs.go`
6. Handlers / tests : `internal/features/**` et `internal/web/handlers/*_test.go` — spec implicite souvent plus fiable que la charte seule

## Avant les passes

1. Lire `decisions.md` en entier (surtout progressive disclosure + Labels).
2. **Diff decisions ↔ code** : pour chaque écart, classer **Doc drift** (mettre à jour decisions) vs **Bug produit** (constat audit). Ne pas traiter un onglet/feature acté en code mais absent de decisions comme bug UX sans vérifier.
3. Si un audit récent existe (transcript / issue) : ne pas re-lister les points déjà actés dans decisions ; noter les régressions seulement.

## P0 UX (autorise un fix immédiat)

- Parcours cassé (CTA mort, wizard bloqué, empty state sans issue)
- Violation **systémique** de charte (ex. plusieurs primaires sur le même écran métier)
- A11y bloquante (pas de nom accessible sur action critique, tableau sans en-têtes)
- Vocabulaire listUI/`Labels.*` **faux** sur un écran hub (nav ≠ page)

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

- [ ] Statuts / rôles via `FormatItemStatus` / `FormatRunStatus` / `FormatRole` (pas de codes bruts)
- [ ] **`Labels.Run`** sur surface actée : nav, H1, breadcrumbs, empty states, CTA, confirms de clôture (marque produit « Revues » OK)
- [ ] **`Labels.Subject`** dans copy / confirms (pas « sujet » hardcodé si preset ≠ sujet)
- [ ] Vocabulaire **Listes / Modèles** piloté par **`!ShowSubjectColumn`** (`listUI`), **pas** seulement `SimpleUI` — piège P1 mono-sujet
- [ ] UI 100 % FR (pas leads / Issue / direct non traduit)
- [ ] Info essentielle en `hint` / `.field-hint`, pas placeholder seul
- [ ] Domaines ≠ Étiquettes (terminologie decisions)

### Passe 2 — Composants

Checklist :

- [ ] Préférer `mb-*` (button, card, alert, badge, toolbar, breadcrumbs…)
- [ ] **Un seul** primaire (`mb-button` sans `variant`, ou variant primary) **par écran** — empty states onboarding peuvent justifier plusieurs CTA
- [ ] Destructif : `variant="danger"` + `confirm()` / `hx-confirm` (sauf exception actée row-action)
- [ ] Clôture revue : primaire + `hx-confirm`, **pas** danger
- [ ] Cohérence listes ↔ détails (badges, overdue, colonnes)
- [ ] Messages succès/erreur : `mb-alert` / flash partials

### Passe 3 — Parcours

Checklist :

- [ ] Flags P0–P3 : `SimpleUI`, `ShowAssign`, `ShowMyTasks`, `ShowCollab`, `ShowSubjectColumn`, `HasJira` / `HasNotion` / `HasWebhooks` / `HasEvidence`
- [ ] Nav + hub org solo (lien header Organisation si pas d'onglet)
- [ ] Breadcrumbs = ancêtres seulement ; H1 = courant ; absents sur pages racine
- [ ] CTA listes en toolbar ; formulaires : primaire en bas (danger-zone archiver = exception courante)
- [ ] Wizard `/revues/nouvelle` : 2 étapes, pas de stepper ; clic modèle = lancer
- [ ] Empty states onboarding vs « aucun résultat filtre »
- [ ] Matrice P3 : CTA masqué **ou** message « non configuré » — pas d'erreur opaque

### Passe 4 — Accessibilité & budgets

Checklist :

- [ ] `lang="fr"`, skip-link, `scope="col"`, `aria-current` sur actif
- [ ] `role="status"` / `alert` (souvent via `mb-alert`)
- [ ] `colspan` aligné sur colonnes conditionnelles
- [ ] Boutons icône : `aria-label` ; colonnes actions : libellé `visually-hidden` si `<th>` vide
- [ ] Budgets éco (signaler dépassement flagrant) : CSS core ≤ 24 Ko / total ≤ 40 Ko ; HTML ≤ 50 Ko/page ; JS app ≤ 15 Ko

### Vérif live (optionnelle)

Si DevAuth / seed dispo : smoke P0 (particulier) et P1 (duo) sur `/revues`, wizard, fiche run. Sinon templates + tests suffisent — le noter dans la synthèse.

## Hors scope (sauf demande explicite)

- Refactor architecture, RBAC, sécurité, perf
- Cosmétique pure sans impact compréhension / action (ex. espacement ±2px) — **a11y et Labels ne sont pas cosmétique**
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

- Stack : Go + chi + `html/template` + HTMX — pas de SPA
- Composants : `@jeb-maker/mb` (voir decisions — version vendorée)
- Si code modifié : `./scripts/check.sh` avant push

## Prompt utilisateur type

> Fais un audit UX/UI en plusieurs passes, confirme les problèmes, puis critique adverse. Pas de code sauf P0.
