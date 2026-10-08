# Décisions produit UI — Revues

Ne pas remonter ces choix comme des problèmes UX. Mettre à jour ce fichier quand une décision change.

Stack (octobre 2026) : SvelteKit SPA + `@jeb-maker/mb`, API JSON `/api/v1`. Routes SPA : `/runs`, `/runs/{id}`, `/runs/{id}/items/{itemId}`, `/subjects`, `/subjects/{id}`, `/subjects/{id}/launch`, `/subjects/new`, `/modeles`, `/modeles/{id}`, `/modeles/new`, `/mes-taches`, `/admin/*`, `/org/select`, `/org/new`, `/login`, `/register`. Les décisions ci-dessous datent pour partie du legacy HTMX ; quand le code SPA diverge, une ligne **À trancher** le signale sans trancher.

## Parcours

| Sujet | Décision |
|-------|----------|
| Accès projet | **Personnes → projets** (`subject_members`) — pas d’équipes dans le parcours nominal ; admin Équipes masqué (icebox), API/schéma conservés (#295). |
| Page d'accueil connectée | **`/` → `/runs`** — revues = hub principal. La SPA redirige `/` authentifié (org active) vers `/runs`. |
| Post-login (1 org) | **`/runs`** — `bootstrap.redirect` / `PostLoginRoute` = `/runs` (plus `/`). |
| CTA « Lancer une revue » sur `/runs` | **Oui** — toolbar + empty states ; lien vers `/subjects` (choix projet) puis `/subjects/{id}/launch`. |
| CTA fiche sujet | **Conservé** — lancement depuis un sujet connu reste possible |
| CTA fiche modèle `/modeles/{id}` | **Oui** — « Lancer » en bas de la carte Points → `/subjects` avec `?template_id=` puis launch avec modèle présélectionné (`?template_id=` sur `/subjects/{id}/launch`). |
| CTA liste `/modeles` | **Oui** — row-action « Lancer » (lien, pas primary) → `/subjects?template_id=` — même flux que la fiche. |
| Stepper wizard | **Supprimé** — fil d'Ariane ; **2 étapes** (sujet → modèle via `/subjects/{id}/launch`, clic = lancer) |
| Titre de page (H1) | **Visible** — `.page-title` = dernier crumb |
| Fil d'Ariane | **Ancêtres seulement** (≥ 2 niveaux) ; **absent** sur pages racine (1 crumb) — le courant = H1. **Toujours pleine largeur de `.page`** — ne pas l’embarquer dans `.page--narrow` (le formulaire seul se centre). **Bande réservée** : `.page` garde la hauteur crumbs même sans fil (`padding-top` + `--page-crumbs-band`) pour ne pas faire sauter le H1. |
| Saisie points (revue en cours) | **Sans confirm** sur changement de statut ; confirm **uniquement** à la clôture |
| Clôturer | `mb-button` **primary** + `confirm()` ; pas `variant="danger"` |
| Fiche point | **Satellite** PJ / Jira / historique — saisie **statut + commentaire (+ assign)** sur la fiche revue en tableau ; lien row-action pour le reste |
| Fiche revue — explications | `help_text` **toujours affiché en entier** sous le libellé du point |
| Fiche revue — grille saisie | Pleine largeur (`page--wide`) : **ligne titre** (libellé + Détails) puis **une ligne** Explication · Statut (boutons verticaux compacts = hauteur de ligne) · Commentaire (même hauteur) · Assigné si dispo. Explication trop longue → scroll. Autosave ; **NOK sans commentaire** = erreur inline |
| Fiche revue — filtre points | Si > 4 points : filtre local Tous · En attente · Non validés · Obligatoires (`mb-segmented-control`) |
| Fiche revue — lede | Date (`created_at`) · version · progression discrète `done/total · %` + fine barre 3px · badge hors `in_progress` · échéance |
| Statut revue à la création | **Directement `in_progress`** — pas d'étape brouillon ni CTA « Démarrer » (legacy `draft` encore démarable) |
| Liste `/runs` | **Pagination** — 25 par page, total affiché ; filtres `?status=` / `?q=` / `?offset=` dans l’URL ; filtre statut **appliqué à la sélection** (pas de bouton Filtrer) ; colonnes triables (client, page courante) dont Date (`created_at`) |
| Post-CRUD sujet | **Créer** (`/subjects/new`) → redirect fiche `/subjects/{id}` (hub métier) ; **modifier / archiver** inline sur la fiche `/subjects/{id}` (capabilities `can_manage`) — plus de routes `/admin/subjects*` |
| Colonne « Auteur » sur `/runs` | **Non** — placeholder sans « auteur » ; recherche SQL par login conservée |
| Titre de revue (UI) | **Supprimé** — pas de champ titre à la création ni en liste (issue parallèle) |
| H1 fiche revue | **Sans `#id`** ; sans nom de sujet en mono-sujet (évite répétition). Listes / exports gardent le label avec `#id` si besoin. **À trancher** : la SPA affiche `run.title` = « modèle · sujet · date · #id » |
| Flash « Revue créée » | **Non** — redirect sans message ; la page elle-même suffit |
| Statut sur fiche revue | Badge omis si `in_progress` (évident) ; garder pour `done` / autres + échéance |
| Progression fiche revue | **Discrète dans le lede** (compteur + barre 3px) — pas sous le H2 Points |
| Liste `/runs` — sujet | Titre = modèle · `#id` (**sans** sujet ni date — la date a sa colonne) ; colonne Sujet **masquée** quand un seul sujet |
| Colonne Assigné (grille points) | **Désactivée** — `FEATURE_ASSIGN_TASKS=false` (UI assignation non fiable) ; API `assigned_to` conservée |
| Colonne Sujet `/runs` | ≥2 sujets visibles (P2) |
| Nav « Mes tâches » | **Désactivée** — même flag ; `/mes-taches` redirige vers `/runs` |
| Fiche sujet Équipes/Membres | **`ShowCollab`** — ≥2 membres ; sinon layout « revues d’abord » |
| Placement CTA | **Listes** : primaire en haut à droite face au H1 (`.page-header__row` + `.page-header__actions`) ; pas sous le titre. Libellés courts face au H1 : **« Nouveau »** (création projet/modèle), **« Lancer »** (création revue — `launchRunCTA`) — le H1 / la carte portent déjà le type. Compteur de résultats (ex. « 2 revues ») sur la **même ligne** que les filtres, aligné à droite (`.filters__count`). **Fiches** : H1 + méta hors cartes ; contenu en `mb-card` / `.card-stack` ; primaire en bas **dans** le footer de la dernière carte métier utile (Points / Revues / Clôture) ; Archiver = carte danger-zone séparée. **Formulaires / édition** (`/subjects/new`, `/modeles/new`, `/modeles/{id}/edit`, `/org/new`, édition inline fiche sujet) : champs en `mb-card` ; primaire en bas dans le footer de la dernière carte métier éditable. **Launch** `/subjects/{id}/launch` : composition unique en `mb-card` (H1 + champs) ; footer `.actions--stack` = Lancer primary + Annuler secondary **pleine largeur**. Satellites lecture/PJ/Jira/historique **sans** primaire. **Un seul** `mb-button` primary par écran ; secondaires en `variant="secondary"` / `ghost`. Si export CSV un jour : secondary (pas preuve). |
| Ledes (sous-titres) | **Minimaux** — supprimer les ledes qui paraphrasent le H1 sur les listes ; garder seulement une instruction utile (ex. mode `?template_id=` sur `/subjects`). |
| Vocabulaire technique UI | **Interdit** dans l’appli : snapshot, matching, whitelist, slug, HMAC, SSRF, « l’API », icebox, parcours nominal, etc. Dire le métier en français courant. OK en docs / OpenAPI / code. |
| Fiche point `/runs/{id}/items/{itemId}` | **`mb-card`** : Saisie (footer = Enregistrer primary) · Pièce jointe · Issue Jira (secondary) · Historique si événements. `help_text` en callout **au-dessus** du `.card-stack`. Historique = `row-list` plat (pas de sous-cartes). |
| Statut vs progression (cartes revue) | **Option 1+5** : badge omis si `in_progress` (la progression suffit) ; colonne Statut **absente en SimpleUI**. Badge conservé pour brouillon / terminée / archivée hors SimpleUI. |
| Libellé runs (instances) | Preset org `ui_run_label` : `revues` (défaut) · `listes_en_cours` · `audits` · `checklists`. Surface : nav, H1, breadcrumbs, empty states, CTA. Particulier (seed) = `listes_en_cours` ; mobile nav short = « En cours ». Pas de marque « Revues » dans le chrome (onglet nav seulement). |
| Accès revues terminées / filtres | Liste `/runs` : filtre **Tous · En cours · Terminées · En retard** (`?status=` ; `overdue` = en retard). Clôture : `POST /runs/{id}/complete` puis rechargement de la fiche côté SPA. |
| Preuve ZIP / hash CSV (produit) | **Abandonnée** — le pack « preuve » (SHA256 scellé + ZIP) est du théâtre conformité ; la traçabilité métier (`run_item_events`, snapshot, `completed_at`, `closing_note`, auteur) suffit. Ne pas re-exposer de CTA « Télécharger la preuve » ni traiter `HasEvidence` comme surface P3. Colonne `evidence_csv_sha256` : interne / legacy, pas un argument produit. |
| Attestation de clôture | **Oui (cible)** — sur fiche revue `done` : encart « Clôturée par X le … » + note de clôture ; optionnellement confirmation explicite à la clôture. Responsabilité humaine visible, pas crypto. |
| Export CSV | **Utilitaire** éventuel (tableur / archive) en `secondary` — **sans** le vendre comme preuve. Pas prioritaire tant que l’attestation (et plus tard Confluence) couvrent l’archive lisible. |
| Notion (intégration) | **Retirée du périmètre produit** — pas d’API ni d’UI (admin / import modèles / export revue / `can_export_notion`). Package et routes retirés ; colonnes legacy en base non lues. Réintro seulement sur signal d’usage. |
| Archive doc | **Confluence Cloud** — publier revue `done` livré (#337). Auth cible : **OAuth 3LO Atlassian** (user-as-self, #340) ; token org = transitoire. Pas de second silo Notion. |
| Jira auth | **OAuth 3LO préféré** au token org partagé (#340) — create/link en tant que l’utilisateur ; hybride lecture/statut org possible jusqu’à bascule. |
| Colonne Sujet `/mes-taches` | **Toujours visible** quand `ShowMyTasks` (P1) — pas gated par `ShowSubjectColumn` (utile même en mono-sujet pour distinguer les tâches) |
| Breadcrumb fiche sujet | Ancêtre → **`/subjects`** |

## Progressive disclosure (paliers)

Objectif : **même produit**, complexité révélée par le contexte d’usage — pas un second produit « lite ».

Flags d'origine (legacy `middleware.resolveUICaps` → `PageData`, **supprimés** avec la stack HTMX) :

| Flag | Seuil | Palier |
|------|--------|--------|
| `SimpleUI` | 1 org · 1 membre · ≤1 sujet | P0 |
| `ShowAssign` / `ShowMyTasks` | off (`FEATURE_ASSIGN_TASKS`) ; seuil historique ≥2 membres | P1 icebox |
| `ShowCollab` | ≥2 membres org | P1 |
| `ShowSubjectColumn` | ≥2 sujets visibles | P2 |
| `HasJira` | intégration Jira **configurée** (org active) | P3 |
| `HasNotion` | **Obsolète (produit)** — Notion hors périmètre | — |
| `HasWebhooks` | webhooks **activés** (URLs + secret) | P3 |
| `HasEvidence` | **Obsolète (produit)** — ne plus utiliser ; voir décision « Preuve ZIP / hash CSV » | — |

| Palier | Déclencheur | Surface |
|--------|-------------|---------|
| **P0 — Particulier** | `SimpleUI` | Listes en cours (nav mobile « En cours ») · Listes ; cocher ; pas assign / tâches / collab |
| **P1 — Duo** | 2ᵉ **membre** (pas seulement invitation pending) | + collab fiche sujet · onglet Organisation si membres · *(Assignation / Mes tâches icebox — `FEATURE_ASSIGN_TASKS`)* |
| **P2 — Multi-sujet** | ≥2 sujets | + Colonne Sujet · domaines · vocabulaire « Modèles » |
| **P3 — Conformité** | Intégration configurée | Jira/webhooks **capability-gated** (config), pas masqués par SimpleUI ; **pas** Notion ni « preuve » |

**Acté (SPA minimale)** : pas de flags legacy `SimpleUI` côté API. Heuristiques front : **Modèles** masqués si `!can_edit` (reader) ; **Mes tâches / Assigné** off via `FEATURE_ASSIGN_TASKS` (réactiver + `member_count ≥ 2`) ; **Personnes** fiche sujet si `show_collab` (`member_count ≥ 2`) ; colonne sujet `/runs` si `visible_subject_count ≥ 2`. Capabilities `can_*` par ressource inchangées. Plein SimpleUI (vocabulaire listes, etc.) reste icebox hors presets `ui_*`.

**Partiel livré** : les presets `ui_run_label` / `ui_subject_label` de l'org active sont consommés par la nav, le hub `/` et les H1 `/runs` · `/subjects` · `/modeles` (`frontend/src/lib/i18n/uiLabels.ts`). Vocabulaire « Listes » pour les modèles si `ui_run_label=listes_en_cours` (heuristique particulier, en attendant `ShowSubjectColumn`). Défaut produit libellé conteneur = **projet** (fallback front si preset vide/inconnu) ; orgs déjà en base avec `sujet` gardent « Sujet » jusqu’à migration/admin. Pas encore d'écran admin pour changer les presets.

### Matrice capability P3

P3 n’est **pas** un unlock structurel : flags org legacy (`HasJira` / `HasWebhooks` ; ~~`HasNotion`~~ retiré produit). `HasEvidence` abandonné. Dans la SPA, P3 = `can_link` (Jira) + écrans `/admin/integrations/*` (sans Notion).

| Surface | Gate PageData | Gate page-spécifique (héritage B4b/B4c) | Visible si SimpleUI ? |
|---------|---------------|------------------------------------------|------------------------|
| Lien / création Jira (point NOK) | `HasJira` | `JiraConfigured` (= `HasJira`) | Oui, si Jira configuré |
| Import / export / admin Notion | ~~`HasNotion`~~ | **Non** — hors périmètre | — |
| Admin webhooks / overview | `HasWebhooks` | `Configured` / overview Enabled | N/A (admin org) |
| Attestation de clôture (fiche `done`) | — | Toujours si `done` (qui / quand / note) | Oui |
| Hash + ZIP preuve | ~~`HasEvidence`~~ | **Non** — ne pas réintroduire | — |
| Archive Confluence | — | **Cible future**, pas livré | — |

Règles :
1. **SimpleUI ne masque jamais** une surface P3 dont la capability est vraie.
2. **Absence de config** → CTA masqué ou message « non configuré », pas une erreur opaque.
3. Admin « configurer l’intégration » reste RBAC org owner/admin (voir `docs/RBAC.md`) — orthogonal aux flags UI.

Principes :
1. **Unlock, don’t fork** — routes et schéma stables ; surface UI seulement.
2. **Vocabulaire suit le palier** — Listes (P0/P1 mono-sujet) → Modèles (P2+).
3. **Déclencheur structurel** — membres / sujets, pas un toggle « mode pro ».
4. **Invitation ≠ collab** — inviter sans acceptation / login n’ouvre pas encore assignation.

## Navigation

| Sujet | Décision |
|-------|----------|
| Onglet principal | **Sans marque / logo** dans la barre. Menu à gauche : **Revues** (preset `ui_run_label` ; particulier Listes en cours) · Projets/Sujets · Modèles (éditeur+) · Mes tâches (≥2 membres) · Admin — recherche + compte à droite. |
| Solo sans onglet Organisation | Lien header **Organisation** → hub minimal (`/admin` org) : Inviter + Mes sujets ; onglet Organisation complet réapparaît après le 2ᵉ **membre** |
| Sujets dans nav principale | **Non** (org classique) — accès via le hub org ou le lancement de revue ; la liste membre est `/subjects` |
| Route `/subjects` (liste membre) | **Conservée** (deep link) mais **hors nav** en org classique — plus de liste admin séparée |
| Mode SimpleUI (particulier) | **Oui** — 1 org / 1 membre / ≤1 sujet / pas admin global. Nav = **Listes en cours · Listes** via preset `ui_run_label=listes_en_cours` (routes `/modeles` / `/runs` inchangées). Vocabulaire « liste » à la place de « modèle ». Fiche sujet = hub checklists ; pas d'Équipes / Membres / domaines. Voir **À trancher** (progressive disclosure). |
| Vocabulaire Listes / Modèles | Piloté par **`!ShowSubjectColumn`** (mono-sujet = Listes), pas seulement `SimpleUI` — évite schisme P1 nav vs formulaires |
| Formulaire `/modeles/new` (listUI) | **Tableau** desktop une ligne (Titre · Catégorie · Explication · Obligatoire · actions) / **cartes** mobile ; Explication compacte (ellipsis), multi-lignes en overlay au focus ; catégorie = `section` ; **nouvelle ligne hérite** de la catégorie précédente ; DnD desktop ; flèches mobile ; CTA hors carte (`lib/components/TemplateEditor.svelte`) |
| Versionnement modèle (save) | **Diff serveur** : structurel (titre, catégorie, obligatoire, ajout/suppression/ordre) → nouvelle `template_version` ; éditorial seul (`help_text`, nom, domaines) → **pas** de nouvelle version (UPDATE sur la dernière version / métadonnées). Snapshots `run_items` inchangés. `POST /templates/{id}/versions` = forcer une nouvelle version. UI edit : CTA « Enregistrer » vs « Publier une nouvelle version ». |
| CSS assets | Styles par page dans les blocs `<style>` Svelte (bundle Vite) ; budget CI = seuils JS/CSS app de `scripts/check.sh` ; Compress gzip middleware côté Go |
| Modèles pour lecteurs | **Masqués** si pas membre d’org active ; droits métier lecture/écriture = **par projet** (viewer/contributor), pas rôle global |
| Deep links `/subjects/{id}` | **Conservés** — accessibles à tous les membres org |

## Terminologie

| Code / interne | Affichage UI |
|----------------|--------------|
| `subject_domains` / `template_domains` | **Domaines** — matching modèle↔sujet (intersection ; modèle sans domaine = tous sujets) |
| `subject_tags` | **Retiré** — étiquettes descriptives supprimées (table droppée, pas de surface API/UI) |
| Colonne modèles index `/modeles` | **Domaines** (plus « Tags ») — aligner placeholder recherche |
| Libellé conteneur (org) | Preset `ui_subject_label` ∈ {**projet** (défaut produit), sujet, cible, entite, asset}. Code/API restent `subjects` ; l’UI dit **Projet(s)** par défaut. `sujet` reste disponible (ton audit). **Schéma** : CHECK/DEFAULT SQL encore sur `sujet` sans `projet` — migration `area:data` requise pour persister le preset. Consommation front partielle (`uiLabels.ts`) ; pas encore d'écran admin. |
| Item `nok` | **Non validé** (code `nok` inchangé) |
| Rôles (`lead`, `reader`…) | Français via un helper de libellés front (`formatRole`) |
| Statuts item / revue | Français via helpers de libellés front (`formatItemStatus` / `formatRunStatus`) |
| `project-meta` (CSS) | **`subject-meta`** — renommage charte |

## Charte

- Un seul `mb-button` primary par écran (sauf empty states onboarding à justifier)
- Destructif (archiver, actions irréversibles de formulaire) : `variant="danger"` + `confirm()`
- Row-actions « Retirer » (membre / équipe / email) : **`ghost` + `confirm()`** — exception Basecamp ; pas `danger` en masse dans les tableaux
- Info essentielle : `hint` / `.field-hint`, pas placeholder seul
- Domaines sujet : **`<details>` options avancées** (formulaire et fiche)
- **Icônes d’action** : SVG inline minimal (paths courts, classe `.icon`) — **pas** de lib npm, webfont, emoji ni composants icon lourds. CTA header = **icône + libellé** ; row-actions denses = lien/bouton `icon-only` / `.row-action` + `aria-label` (+ `title`) ; actions dangereuses / clôture **gardent le texte**. Pas d’icônes dans la nav, badges ou empty states.

## Design system (`@jeb-maker/mb`)

| Sujet | Décision |
|-------|----------|
| Version consommée | **`0.5.1`** (tag Git `v0.5.1`) |
| Tokens | **`tokens-core.css`** (+ `mb-bridge.css`), injectés par `ensureMb()` depuis le layout racine — pas de `tokens.css` (évite reset `html`/`body`) ni `typography.css`/woff2 (budget) |
| JS | `mb-boot.js` sous `web/static/vendor/jeb-maker-mb/` (Lit bundlé) — vendor mesuré hors budget app (`check.sh`) ; chargé par `ensureMb()` |
| Shell + formulaires | Cible `mb-*` (button, alert, badge, input, textarea, tag, empty-state, spinner, table, **card**…) — fiches (`/modeles/{id}`, `/subjects/{id}`, `/runs/{id}`, fiche point) en `mb-card` + `.card-stack` ; listes = tableau + header row ; nav / breadcrumbs / toolbar / modal / toast encore à venir |
| `mb-button` busy / gated | **`loading={busy}`** pour l’en-vol (pas `disabled={busy}`). Ne pas gater un submit avec `disabled` lié à un champ vide : préférer `required` + garde submit. Patch vendor local : ignore `formDisabledCallback` sticky (voir README vendor). `disabled=` réservé aux états structurels (`!configured`, pagination). |
| Listes | **`mb-table`** (cible : admin, sujets, modèles, revues, tâches, sections de la fiche revue) |
| Reste host | éditeur de modèle (`TemplateEditor.svelte`, `<table>` natif + DnD) ; `confirm()` natif |
| Tracking gaps | https://github.com/jeb-maker/miniature-broccoli/issues/40–44 (fermés en 0.4.1) |

## Dette doc connue

Points encore ouverts : format H1 fiche revue (sans #id) ; écran admin presets `ui_*` ; schéma de couleur (fond sombre vs thème clair mb) ; SimpleUI vocabulaire listes complet ; Confluence archive ; CSV utilitaire.

**Livré (session)** : `/` + post-login → `/runs` ; CTA lancer `/runs` + modèle ; pagination + filtres URL ; historique revues fiche sujet ; grille statut inline + attestation clôture ; disclosure nav (modeles/readers, mes-tâches ≥2, colonne sujet ≥2) ; garde `/admin` ; webhooks dans AdminNav ; migration `projet` + `completed_by` ; Notion retiré.
