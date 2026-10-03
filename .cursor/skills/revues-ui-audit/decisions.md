# Décisions produit UI — Revues

Ne pas remonter ces choix comme des problèmes UX. Mettre à jour ce fichier quand une décision change.

Stack (octobre 2026) : SvelteKit SPA + `@jeb-maker/mb`, API JSON `/api/v1`. Routes SPA : `/runs`, `/runs/{id}`, `/runs/{id}/items/{itemId}`, `/subjects`, `/subjects/{id}`, `/subjects/{id}/launch`, `/subjects/new`, `/modeles`, `/modeles/{id}`, `/modeles/new`, `/mes-taches`, `/admin/*`, `/org/select`, `/org/new`, `/login`, `/register`. Les décisions ci-dessous datent pour partie du legacy HTMX ; quand le code SPA diverge, une ligne **À trancher** le signale sans trancher.

## Parcours

| Sujet | Décision |
|-------|----------|
| Page d'accueil connectée | **`/` → `/runs`** — revues = hub principal. **À trancher** : la SPA sert sur `/` une page hub (liens Sujets · Modèles · Revues · Mes tâches · Administration) au lieu de rediriger vers `/runs` |
| Post-login (1 org) | **`/runs`** — plus `/subjects`. **À trancher** : `bootstrap.redirect` renvoie `/` quand une org est active |
| CTA « Lancer une revue » sur `/runs` | **Oui** — toolbar + empty states ; lancement via `/subjects/{id}/launch` (choix du modèle). **À trancher** : la page SPA `/runs` n'a pas encore ce CTA (seule la fiche sujet le porte) |
| CTA fiche sujet | **Conservé** — lancement depuis un sujet connu reste possible |
| CTA fiche modèle `/modeles/{id}` | **Oui** — « Lancer avec ce modèle » → choix du sujet puis `/subjects/{id}/launch` avec modèle présélectionné ; pas de lancement sans sujet (matching domaines). **À trancher** : absent de la fiche modèle SPA |
| Stepper wizard | **Supprimé** — fil d'Ariane ; **2 étapes** (sujet → modèle via `/subjects/{id}/launch`, clic = lancer) |
| Titre de page (H1) | **Visible** — `.page-title` = dernier crumb |
| Fil d'Ariane | **Ancêtres seulement** (≥ 2 niveaux) ; **absent** sur pages racine (1 crumb) — le courant = H1 |
| Saisie points (revue en cours) | **Sans confirm** sur changement de statut ; confirm **uniquement** à la clôture |
| Clôturer | `mb-button` **primary** + `confirm()` ; pas `variant="danger"` |
| Fiche point | **Satellite** PJ / Jira / historique — saisie statut/commentaire dans la grille ; lien **Détails** discret |
| Statut revue à la création | **Directement `in_progress`** — pas d'étape brouillon ni CTA « Démarrer » (legacy `draft` encore démarable) |
| Liste `/runs` | **Pagination** — 25 par page, total affiché. **À trancher** : la page SPA `/runs` filtre par statut (`?status=`) sans pagination visible |
| Post-CRUD sujet | **Créer** (`/subjects/new`) → redirect fiche `/subjects/{id}` (hub métier) ; **modifier / archiver** inline sur la fiche `/subjects/{id}` (capabilities `can_manage`) — plus de routes `/admin/subjects*` |
| Colonne « Auteur » sur `/runs` | **Non** — placeholder sans « auteur » ; recherche SQL par login conservée |
| Titre de revue (UI) | **Supprimé** — pas de champ titre à la création ni en liste (issue parallèle) |
| H1 fiche revue | **Sans `#id`** ; sans nom de sujet en mono-sujet (évite répétition). Listes / exports gardent le label avec `#id` si besoin. **À trancher** : la SPA affiche `run.title` = « modèle · sujet · date · #id » |
| Flash « Revue créée » | **Non** — redirect sans message ; la page elle-même suffit |
| Statut sur fiche revue | Badge omis si `in_progress` (évident) ; garder pour `done` / autres + échéance |
| Progression fiche revue | **Sous le H2 Points**, pas dans un bandeau meta au-dessus |
| Liste `/runs` — sujet | Titre = modèle · date · `#id` (**sans** sujet) ; colonne Sujet **masquée** quand un seul sujet |
| Colonne Assigné (grille points) | **`ShowAssign`** — ≥2 membres org (P1), pas seulement `!SimpleUI` |
| Colonne Sujet `/runs` | ≥2 sujets visibles (P2) |
| Nav « Mes tâches » | **`ShowMyTasks`** — ≥2 membres org (P1) |
| Fiche sujet Équipes/Membres | **`ShowCollab`** — ≥2 membres ; sinon layout « revues d’abord » |
| Placement CTA | **Listes** : primaire dans la toolbar de la carte (pas sous le H1). **Formulaires** : primaire en bas **dans** la dernière carte (carte Archiver / danger-zone = exception). **Un seul** `mb-button` primary par écran ; secondaires en `variant="secondary"` / `ghost`. Export CSV revue terminée = secondary. |
| Statut vs progression (cartes revue) | **Option 1+5** : badge omis si `in_progress` (la progression suffit) ; colonne Statut **absente en SimpleUI**. Badge conservé pour brouillon / terminée / archivée hors SimpleUI. |
| Libellé runs (instances) | Preset org `ui_run_label` : `revues` (défaut) · `listes_en_cours` · `audits` · `checklists`. Surface : nav, H1, breadcrumbs, empty states, CTA. Particulier (seed) = `listes_en_cours` ; mobile nav short = « En cours ». Marque produit « Revues » inchangée. |
| Accès revues terminées / filtres | Liste `/runs` : filtre **Tous · En cours · Terminées · En retard** (`?status=` ; `overdue` = en retard). Clôture : `POST /runs/{id}/complete` puis rechargement de la fiche côté SPA. |
| Colonne Sujet `/mes-taches` | **Toujours visible** quand `ShowMyTasks` (P1) — pas gated par `ShowSubjectColumn` (utile même en mono-sujet pour distinguer les tâches) |
| Breadcrumb fiche sujet | Ancêtre → **`/subjects`** |

## Progressive disclosure (paliers)

Objectif : **même produit**, complexité révélée par le contexte d’usage — pas un second produit « lite ».

Flags d'origine (legacy `middleware.resolveUICaps` → `PageData`, **supprimés** avec la stack HTMX) :

| Flag | Seuil | Palier |
|------|--------|--------|
| `SimpleUI` | 1 org · 1 membre · ≤1 sujet · whitelist ≤1 · pas admin global | P0 |
| `ShowAssign` / `ShowMyTasks` / `ShowCollab` | ≥2 membres org | P1 |
| `ShowSubjectColumn` | ≥2 sujets visibles | P2 |
| `HasJira` | intégration Jira **configurée** (org active) | P3 |
| `HasNotion` | intégration Notion **configurée** (token présent) | P3 |
| `HasWebhooks` | webhooks **activés** (URLs + secret) | P3 |
| `HasEvidence` | preuve **scellée** sur la fiche revue (`done` + hash CSV) — page-scoped | P3 |

| Palier | Déclencheur | Surface |
|--------|-------------|---------|
| **P0 — Particulier** | `SimpleUI` | Listes en cours (nav mobile « En cours ») · Listes ; cocher ; CSV ; pas assign / tâches / collab |
| **P1 — Duo** | 2ᵉ **membre** (pas seulement whitelist) | + Assignation · Mes tâches · collab fiche sujet · onglet Organisation si whitelist/membres |
| **P2 — Multi-sujet** | ≥2 sujets | + Colonne Sujet · domaines · vocabulaire « Modèles » |
| **P3 — Conformité** | Intégration configurée / preuve scellée | Notion/Jira/webhooks/preuve restent **capability-gated** (config ou hash), pas masqués par SimpleUI |

**À trancher** : la SPA n'implémente pas ces paliers (aucun flag `SimpleUI` / `ShowAssign` / `ShowSubjectColumn` exposé par l'API ; seules des `capabilities.can_*` par ressource existent dans `openapi.yaml`). Réimplémenter la progressive disclosure côté front ou acter son abandon.

**Partiel livré** : les presets `ui_run_label` / `ui_subject_label` de l'org active sont consommés par la nav, le hub `/` et les H1 `/runs` · `/subjects` · `/modeles` (`frontend/src/lib/i18n/uiLabels.ts`). Vocabulaire « Listes » pour les modèles si `ui_run_label=listes_en_cours` (heuristique particulier, en attendant `ShowSubjectColumn`). Pas encore d'écran admin pour changer les presets.

### Matrice capability P3

P3 n’est **pas** un unlock structurel : dans le legacy, les flags org (`HasJira` / `HasNotion` / `HasWebhooks`) étaient résolus via la config chiffrée et `HasEvidence` posé par la fiche revue (hash scellé). Dans la SPA, l'équivalent est porté par les capabilities par ressource (`can_link`, `can_export_notion`) et les écrans `/admin/integrations/*`.

| Surface | Gate PageData | Gate page-spécifique (héritage B4b/B4c) | Visible si SimpleUI ? |
|---------|---------------|------------------------------------------|------------------------|
| Lien / création Jira (point NOK) | `HasJira` | `JiraConfigured` (= `HasJira`) | Oui, si Jira configuré |
| Import Notion (`/modeles/…`) | `HasNotion` | `NotionConfigured` (= `HasNotion`) | Oui, si Notion configuré |
| Export Notion (revue `done`) | `HasNotion` (prérequis token) | `NotionConfigured` = **ExportReady** (token + base) | Oui, si export ready |
| Admin webhooks / overview | `HasWebhooks` | `Configured` / overview Enabled | N/A (admin org) |
| Hash + ZIP preuve (revue `done`) | `HasEvidence` | `CanExportEvidence` (= `HasEvidence`) | Oui, si hash scellé |

Règles :
1. **SimpleUI ne masque jamais** une surface P3 dont la capability est vraie.
2. **Absence de config / hash** → CTA masqué ou message « non configuré », pas une erreur opaque.
3. Admin « configurer l’intégration » reste RBAC org owner/admin (voir `docs/RBAC.md`) — orthogonal aux flags UI.

Principes :
1. **Unlock, don’t fork** — routes et schéma stables ; surface UI seulement.
2. **Vocabulaire suit le palier** — Listes (P0/P1 mono-sujet) → Modèles (P2+).
3. **Déclencheur structurel** — membres / sujets, pas un toggle « mode pro ».
4. **Whitelist ≠ collab** — inviter sans login n’ouvre pas encore assignation.

## Navigation

| Sujet | Décision |
|-------|----------|
| Onglet principal | **Revues** (preset org `ui_run_label` ; particulier Listes en cours) · Mes tâches · Modèles (éditeur+) · **Organisation** (admin org ; intégrations dans le sous-menu org) — **sans logo/marque** dans la barre (menu seul). **À trancher** : la SPA n'a pas encore de barre de navigation commune (chaque page porte son en-tête) |
| Solo sans onglet Organisation | Lien header **Organisation** → hub minimal (`/admin` org) : Inviter + Mes sujets ; onglet Organisation complet réapparaît après le 2ᵉ email whitelisté |
| Sujets dans nav principale | **Non** (org classique) — accès via le hub org ou le lancement de revue ; la liste membre est `/subjects` |
| Route `/subjects` (liste membre) | **Conservée** (deep link) mais **hors nav** en org classique — plus de liste admin séparée |
| Mode SimpleUI (particulier) | **Oui** — 1 org / 1 membre / ≤1 sujet / pas admin global. Nav = **Listes en cours · Listes** via preset `ui_run_label=listes_en_cours` (routes `/modeles` / `/runs` inchangées). Vocabulaire « liste » à la place de « modèle ». Fiche sujet = hub checklists ; pas d'Équipes / Membres / domaines. Voir **À trancher** (progressive disclosure). |
| Vocabulaire Listes / Modèles | Piloté par **`!ShowSubjectColumn`** (mono-sujet = Listes), pas seulement `SimpleUI` — évite schisme P1 nav vs formulaires |
| Formulaire `/modeles/new` (listUI) | **Tableau** desktop une ligne (Case · Catégorie · Aide · Obligatoire · actions) / **cartes** mobile ; catégorie = `section` ; **nouvelle ligne hérite** de la catégorie précédente ; Aide sans « optionnel » ; DnD desktop ; flèches mobile ; CTA hors carte (`lib/components/TemplateEditor.svelte`) |
| CSS assets | Styles par page dans les blocs `<style>` Svelte (bundle Vite) ; budget CI = seuils JS/CSS app de `scripts/check.sh` ; Compress gzip middleware côté Go |
| Modèles pour lecteurs | **Masqués** — rôle `reader` seul n'a pas l'onglet Modèles |
| Deep links `/subjects/{id}` | **Conservés** — accessibles à tous les membres org |

## Terminologie

| Code / interne | Affichage UI |
|----------------|--------------|
| `subject_domains` / `template_domains` | **Domaines** — matching modèle↔sujet (intersection ; modèle sans domaine = tous sujets) |
| `subject_tags` | **Étiquettes** — descriptif uniquement, pas de filtrage modèle |
| Colonne modèles index `/modeles` | **Domaines** (plus « Tags ») — aligner placeholder recherche |
| Libellé sujet (org) | Preset admin `ui_subject_label` ∈ {sujet, cible, entite, asset} — défaut `sujet` (champ `Organization.ui_subject_label`). **À trancher** : plus d'écran d'édition ni de consommation front |
| Item `nok` | **Non validé** (code `nok` inchangé) |
| Rôles (`lead`, `reader`…) | Français via un helper de libellés front (`formatRole`) |
| Statuts item / revue | Français via helpers de libellés front (`formatItemStatus` / `formatRunStatus`) |
| `project-meta` (CSS) | **`subject-meta`** — renommage charte |

## Charte

- Un seul `mb-button` primary par écran (sauf empty states onboarding à justifier)
- Destructif (archiver, actions irréversibles de formulaire) : `variant="danger"` + `confirm()`
- Row-actions « Retirer » (membre / équipe / email) : **`ghost` + `confirm()`** — exception Basecamp ; pas `danger` en masse dans les tableaux
- Info essentielle : `hint` / `.field-hint`, pas placeholder seul
- Domaines / étiquettes sujet : **`<details>` options avancées** (formulaire et fiche)

## Design system (`@jeb-maker/mb`)

| Sujet | Décision |
|-------|----------|
| Version consommée | **`0.4.1`** (tag Git `v0.4.1`) |
| Tokens | **`tokens-core.css`** (+ `mb-bridge.css`), injectés par `ensureMb()` depuis le layout racine — pas de `tokens.css` (évite reset `html`/`body`) ni `typography.css`/woff2 (budget) |
| JS | `mb-boot.js` sous `web/static/vendor/jeb-maker-mb/` (Lit bundlé) — vendor mesuré hors budget app (`check.sh`) ; chargé par `ensureMb()` |
| Shell + formulaires | Cible `mb-*` (button, alert, badge, input, textarea, tag, empty-state, spinner, table…) — la SPA n'utilise pas encore nav / breadcrumbs / card / toolbar / modal / toast |
| Listes | **`mb-table`** (cible : admin, sujets, modèles, revues, tâches, sections de la fiche revue) |
| Reste host | éditeur de modèle (`TemplateEditor.svelte`, `<table>` natif + DnD) ; `confirm()` natif |
| Tracking gaps | https://github.com/jeb-maker/miniature-broccoli/issues/40–44 (fermés en 0.4.1) |

## Dette doc connue

Points **À trancher** (code SPA ≠ décision, laissés en l'état volontairement) : redirection `/` → `/runs` et post-login ; CTA « Lancer une revue » sur `/runs` et `/modeles/{id}` ; pagination `/runs` ; format H1 fiche revue ; progressive disclosure (paliers P0–P3 sans flags API) ; preset `ui_subject_label` sans écran ; barre de navigation commune ; schéma de couleur (fond sombre dégradé actuel vs thème clair mb « Basecamp » de la charte).
