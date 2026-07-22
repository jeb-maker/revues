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
| H1 fiche revue | **Sans `#id`** ; SimpleUI sans nom de sujet (évite répétition). Listes / exports gardent le label avec `#id` si besoin |
| Flash « Revue créée » | **Non** — redirect sans message ; la page elle-même suffit |
| Statut sur fiche revue | Badge omis si `in_progress` (évident) ; garder pour `done` / autres + échéance |
| Progression fiche revue | **Sous le H2 Points**, pas dans un bandeau meta au-dessus |
| Liste `/revues` — sujet | Titre = modèle · date · `#id` (**sans** sujet) ; colonne Sujet **masquée en SimpleUI** (un seul sujet) |
| Colonne Assigné (grille points) | **`ShowAssign`** — ≥2 membres org (P1), pas seulement `!SimpleUI` |
| Colonne Sujet `/revues` | **`ShowSubjectColumn`** — ≥2 sujets visibles (P2) |
| Nav « Mes tâches » | **`ShowMyTasks`** — ≥2 membres org (P1) |
| Fiche sujet Équipes/Membres | **`ShowCollab`** — ≥2 membres ; sinon layout « revues d’abord » |
| Placement CTA | **Listes** : primaire dans la toolbar de la carte (pas sous le H1). **Formulaires** : primaire en bas **dans** la dernière carte. **Un seul** `.button` plein par écran ; secondaires en `.button-secondary` / ghost. Export CSV revue terminée = secondary. |
| Statut vs progression (cartes revue) | **Option 1+5** : badge omis si `in_progress` (la progression suffit) ; colonne Statut **absente en SimpleUI**. Badge conservé pour brouillon / terminée / archivée hors SimpleUI. |
| Libellé runs (instances) | Preset org `ui_run_label` : `revues` (défaut) · `listes_en_cours` · `audits` · `checklists`. Surface : nav, H1, breadcrumbs, empty states, CTA. Particulier (seed) = `listes_en_cours` ; mobile nav short = « En cours ». Marque produit « Revues » inchangée. |
| Accès revues terminées | Liste `/revues` : onglets **Tous · En cours · Terminées** (`?status=`). Clôture HTMX via `HX-Redirect` (client minimal). |

## Progressive disclosure (paliers)

Objectif : **même produit**, complexité révélée par le contexte d’usage — pas un second produit « lite ».

Flags runtime (`middleware.resolveUICaps` → `PageData`) :

| Flag | Seuil |
|------|--------|
| `SimpleUI` | 1 org · 1 membre · ≤1 sujet · whitelist ≤1 · pas admin global |
| `ShowAssign` / `ShowMyTasks` / `ShowCollab` | ≥2 membres org |
| `ShowSubjectColumn` | ≥2 sujets visibles |

| Palier | Déclencheur | Surface |
|--------|-------------|---------|
| **P0 — Particulier** | `SimpleUI` | Listes en cours (nav mobile « En cours ») · Listes ; cocher ; CSV ; pas assign / tâches / collab |
| **P1 — Duo** | 2ᵉ **membre** (pas seulement whitelist) | + Assignation · Mes tâches · collab fiche sujet · onglet Organisation si whitelist/membres |
| **P2 — Multi-sujet** | ≥2 sujets | + Colonne Sujet · domaines · vocabulaire « Modèles » |
| **P3 — Conformité** | Intégration configurée / preuve scellée | Notion/Jira/webhooks/preuve restent **capability-gated** (config ou hash), pas masqués par SimpleUI |

Principes :
1. **Unlock, don’t fork** — routes et schéma stables ; surface UI seulement.
2. **Vocabulaire suit le palier** — Listes (P0/P1 mono-sujet) → Modèles (P2+).
3. **Déclencheur structurel** — membres / sujets, pas un toggle « mode pro ».
4. **Whitelist ≠ collab** — inviter sans login n’ouvre pas encore assignation.

## Navigation

| Sujet | Décision |
|-------|----------|
| Onglet principal | **{{Labels.Run.Nav}}** (défaut Revues ; particulier Listes en cours) · Mes tâches · Modèles (éditeur+) · **Organisation** (admin org ; intégrations dans le sous-menu org) — **sans logo/marque** dans la barre (menu seul) |
| Solo sans onglet Organisation | Lien header **Organisation** → hub minimal (`/admin` org) : Inviter + Mes sujets ; onglet Organisation complet réapparaît après le 2ᵉ email whitelisté |
| Sujets dans nav principale | **Non** (org classique) — sous Organisation → `/admin/subjects` ; accès via hub org ou wizard |
| Route `/subjects` (liste membre) | **Conservée** (deep link) mais **hors nav** en org classique — liste admin = `/admin/subjects` |
| Mode SimpleUI (particulier) | **Oui** — 1 org / 1 membre / ≤1 sujet / pas admin global. Nav = **Listes en cours · Listes** via preset `ui_run_label=listes_en_cours` (routes `/modeles` / `/revues` inchangées). Vocabulaire « liste » à la place de « modèle ». Fiche sujet = hub checklists ; pas d'Équipes / Membres / domaines. |
| Vocabulaire Listes / Modèles | Piloté par **`!ShowSubjectColumn`** (mono-sujet = Listes), pas seulement `SimpleUI` — évite schisme P1 nav vs formulaires |
| Formulaire `/modeles/new` (listUI) | **Compact** : 1 card, 1 point vide, libellé seul + Options repliées ; CTA « Créer la liste » ; domaines en `<details>` côté org |
| CSS assets | **Découpage** : `app.css` (core) + `run.css` + `editor.css` à la demande ; budget CI = gzip cumulé ≤ 12 Ko ; Compress gzip middleware |
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

## Design system (`@jeb-maker/mb`)

| Sujet | Décision |
|-------|----------|
| Version consommée | **`0.2.0`** (tag Git `v0.2.0`) — pas de tag `v0.3.0` publié au moment de l'adoption ; si « 0.3.0 » est cité ailleurs, vérifier les tags avant upgrade |
| Tokens | **`tokens-core.css`** (+ `mb-bridge.css`) — pas de `tokens.css` (évite reset `html`/`body`) ni `typography.css`/woff2 (budget) |
| JS | `mb-boot.js` sous `web/static/vendor/jeb-maker-mb/` (Lit bundlé) — hors budget 15 KiB app |
| Vague 1 (démarrée) | Login / signaler flashes+boutons ; `/revues` empty-state + segmented + pagination ; progress fiche revue |
| Grille points (status/assign) | **Reporté** — `mb-select` SSR OK en 0.2.0, mais HTMX écoute encore `change` natif ; migrer vers `mb-change` dans une PR dédiée |
| Tracking gaps | https://github.com/jeb-maker/miniature-broccoli/issues/19 |

## Dette doc connue

- `docs/DESIGN.md` et `/styleguide` référencés dans la charte mais absents — issue doc dédiée si besoin
