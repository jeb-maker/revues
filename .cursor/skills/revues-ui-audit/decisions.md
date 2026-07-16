# Décisions produit UI — Revues

Ne pas remonter ces choix comme des problèmes UX. Mettre à jour ce fichier quand une décision change.

## Parcours

| Sujet | Décision |
|-------|----------|
| Page d'accueil connectée | **`/` → `/revues`** — revues = hub principal |
| Post-login (1 org) | **`/revues`** — plus `/subjects` |
| CTA « Lancer une revue » sur `/revues` | **Oui** — toolbar + empty states ; wizard `/revues/nouvelle` |
| CTA fiche sujet | **Conservé** — lancement depuis un sujet connu reste possible |
| CTA fiche modèle `/modeles/{id}` | **Oui** — « Lancer avec ce modèle » → wizard étape 1 (choix sujet), modèle présélectionné à l'étape 2 ; pas de lancement sans sujet (matching domaines) |
| Stepper wizard | **Supprimé** — fil d'Ariane ; **2 étapes** (sujet → modèle, clic = lancer) |
| Titre de page (H1) | **Visible** — `.page-title` = dernier crumb |
| Fil d'Ariane | **Ancêtres seulement** (≥ 2 niveaux) ; **absent** sur pages racine (1 crumb) — le courant = H1 |
| Saisie points (revue en cours) | **Sans confirm** sur changement de statut ; confirm **uniquement** à la clôture |
| Clôturer | Bouton **primary** (`.button`) + `hx-confirm` ; pas `.button-danger` |
| Fiche point | **Satellite** PJ / Jira / historique — saisie statut/commentaire dans la grille ; lien **Détails** discret |
| Statut revue à la création | **Directement `in_progress`** — pas d'étape brouillon ni CTA « Démarrer » (legacy `draft` encore démarable) |
| Liste `/revues` | **Pagination** — 25 par page (`?page=`), total affiché ; pas de SPA |
| Post-CRUD sujet (admin org) | **Créer** → redirect fiche `/subjects/{id}` (hub métier) ; **modifier / archiver** depuis admin → liste `/admin/subjects` ; « Modifier » fiche sujet → `/admin/subjects/{id}/edit` pour admin org |
| Colonne « Auteur » sur `/revues` | **Non** — placeholder sans « auteur » ; recherche SQL par login conservée |
| Titre de revue (UI) | **Supprimé** — pas de champ titre à la création ni en liste (issue parallèle) |

## Navigation

| Sujet | Décision |
|-------|----------|
| Onglet principal | **Revues** · Mes tâches · Modèles (éditeur+) · **Organisation** (admin org ; intégrations dans le sous-menu org) — **sans logo/marque** dans la barre (menu seul) |
| Solo sans onglet Organisation | Lien header **Organisation** → hub minimal (`/admin` org) : Inviter + Mes sujets ; onglet Organisation complet réapparaît après le 2ᵉ email whitelisté |
| Sujets dans nav principale | **Non** — déplacés sous Organisation → `/admin/subjects` ; solo : pas de liste dans nav principale, accès via hub org ou wizard |
| Route `/subjects` (liste membre) | **Conservée** (deep link) mais **hors nav** — usage quotidien = wizard + fiches ; liste admin = `/admin/subjects` |
| Modèles pour lecteurs | **Masqués** — rôle `reader` seul n'a pas l'onglet Modèles |
| Deep links `/subjects/{id}` | **Conservés** — accessibles à tous les membres org |

## Terminologie

| Code / interne | Affichage UI |
|----------------|--------------|
| `subject_domains` / `template_domains` | **Domaines** — matching modèle↔sujet (intersection ; modèle sans domaine = tous sujets) |
| `subject_tags` | **Étiquettes** — descriptif uniquement, pas de filtrage modèle |
| Colonne modèles index `/modeles` | **Domaines** (plus « Tags ») — aligner placeholder recherche |
| Libellé sujet (org) | Preset admin `ui_subject_label` ∈ {sujet, cible, entite, asset} — défaut `sujet` ; écran `/admin/settings/labels` |
| Item `nok` | **Non validé** (code `nok` inchangé) |
| Rôles (`lead`, `reader`…) | Français via `formatRole` |
| Statuts item / revue | Français via `formatItemStatus` / `formatRunStatus` |
| `project-meta` (CSS) | **`subject-meta`** — renommage charte |

## Charte

- Un seul `.button` plein par écran (sauf empty states onboarding à justifier)
- Destructif : `.button-danger` + `confirm()`
- Info essentielle : `.field-hint`, pas placeholder seul
- Domaines / étiquettes sujet : **`<details>` options avancées** (formulaire et fiche)

## Dette doc connue

- `docs/DESIGN.md` et `/styleguide` référencés dans la charte mais absents — issue doc dédiée si besoin
