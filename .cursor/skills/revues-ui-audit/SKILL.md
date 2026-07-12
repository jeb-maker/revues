---
name: revues-ui-audit
description: >-
  Audite l'UI/UX de Revues en passes structurées avec confirmation et critique
  adverse. Use when the user asks for a UX/UI review, incohérences interface,
  audit design, contrôle charte, passe UI, or similar quality checks on templates
  and CSS — read-only by default unless a P0 bug is confirmed.
---

# Audit UX/UI — Revues

Audit **read-only** par défaut. N'implémenter que si bug bloquant confirmé (`./scripts/check.sh` rouge).

## Références (lire en premier)

1. `.cursor/rules/design.md` — charte
2. `web/templates/**/*.html`, `web/static/css/app.css`
3. `internal/web/templates/templates.go` — `formatItemStatus`, `formatRunStatus`, `formatRole`
4. `internal/web/handlers/*_test.go` — spec implicite (souvent plus fiable que la charte seule)

Décisions produit actées : [decisions.md](decisions.md)

## Méthode — 5 passes + confirmation + critique adverse

### Passe 1 — Libellés & i18n
Statuts, rôles, vocabulaire métier, messages d'erreur, placeholders vs charte (`.field-hint`).

### Passe 2 — Composants
`.button*`, `.panel`/`.card`, `empty-state`, toolbars, messages succès/erreur, cohérence listes ↔ détails.

### Passe 3 — Parcours
Navigation, breadcrumbs, CTA, wizard lancer revue, onboarding / états vides.

### Passe 4 — Accessibilité
`scope="col"`, `aria-current`, `role="status"`/`alert`, colspan tableaux, labels/aria.

### Passe 5 — Confirmation
Pour **chaque** constat : **Confirmé** / **Partiel** / **Rejeté** + preuve (fichier, test, handler).

### Critique adverse (obligatoire)
Pour chaque point retenu : argumenter pour/contre ; séparer bug produit, dette charte, choix intentionnel, faux positif.

## Hors scope (sauf demande explicite)

- Refactor architecture, RBAC, sécurité, perf
- Harmonisation cosmétique sans impact utilisateur
- Implémentation large sans validation utilisateur

Dette doc (`docs/DESIGN.md`, `/styleguide`) : noter seulement, ne pas implémenter.

## Livrable

```markdown
# Audit UX/UI Revues — [date]

## Synthèse
[2–3 phrases]

## Constats par passe
| # | Passe | Constat | Statut | Preuve |
|---|-------|---------|--------|--------|

## Critique adverse
[Ce qui tient / ce qui est downgradé / pièges de l'audit]

## Top 10 actionnable
| Rang | Action | Effort | Priorité |

## Décisions produit ouvertes
[Si aucune : « Aucune »]

## Plan PR suggéré
[2–3 PR max, sans implémenter]
```

## Rappels projet

- Stack : Go + chi + `html/template` + HTMX — pas de SPA
- Budgets : CSS ≤ 20 Ko, HTML ≤ 50 Ko/page
- `./scripts/check.sh` avant tout push si code modifié

## Prompt utilisateur type

> Fais un audit UX/UI en plusieurs passes, confirme les problèmes, puis critique adverse. Pas de code sauf P0.
