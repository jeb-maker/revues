# Décisions produit UI — Revues

Ne pas remonter ces choix comme des problèmes UX. Mettre à jour ce fichier quand une décision change.

## Parcours

| Sujet | Décision |
|-------|----------|
| CTA « Lancer une revue » sur `/revues` | **Non** — le lancement part d'un projet (CTA fiche projet via `PageActions`) |
| Stepper wizard | **Supprimé** — fil d'Ariane suffit (Projet → Modèle → Confirmation) |
| Colonne « Auteur » sur `/revues` | **Non** — placeholder sans « auteur » ; recherche SQL par login conservée |

## Terminologie

| Code / interne | Affichage UI |
|----------------|--------------|
| Item `nok` | **Non validé** (code `nok` inchangé) |
| Rôles (`lead`, `reader`…) | Français via `formatRole` |
| Statuts item / revue | Français via `formatItemStatus` / `formatRunStatus` |

## Charte

- Un seul `.button` plein par écran (sauf empty states onboarding à justifier)
- Destructif : `.button-danger` + `confirm()`
- Info essentielle : `.field-hint`, pas placeholder seul

## Dette doc connue

- `docs/DESIGN.md` et `/styleguide` référencés dans la charte mais absents — issue doc dédiée si besoin
