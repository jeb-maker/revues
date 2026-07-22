# Critique présence — persona **particulier** (Wave 2)

**Date** : 2026-07-19  
**Persona** : `particulier` / Camille Particulier — org Perso Camille, `SimpleUI=true`, `ui_run_label=listes_en_cours`  
**Caps** : `ShowAssign/MyTasks/Collab/SubjectColumn=false` ; owner org ; 1 sujet « Chez moi »  
**Méthode** : walkthrough inventaire × matrice P0 ; preuve template/handler + HTML rendu (DevAuth user_id=6, 127.0.0.1:8080)  
**Read-only** — aucune correction

Légende **sens** : `utile` · `inutile` · `trompeur` · `mort` (route/UI morte)  
Légende **présence?** : Oui / Non / Deep (reachable sans nav) — *attendu P0* entre parenthèses si écart

---

## Synthèse

Le cœur P0 (nav **Listes en cours · Listes**, cocher, clôturer, CSV) est bien gated. Les fuites majeures de **présence cognitive** sont lexicales : collision **Listes** (templates) vs **liste** (instances `listes_en_cours`) + chaînes hardcodées **revue/modèle** sur le parcours chaud. Second plan : chrome satellite (Jira non configuré, badge Éditeur, hub `/admin` complet en deep link, `/mes-taches` vide).

---

## Chrome global

| écran | élément | présence? | sens | statut | preuve |
|-------|---------|-----------|------|--------|--------|
| layout | Nav Listes en cours + Listes | Oui (attendu) | utile | Confirmé | `site_nav.html` L3–6 ; HTML nav 2 liens |
| layout | Nav Mes tâches | Non (attendu) | — | Confirmé | `site_nav.html` branche `SimpleUI` ; matrice ShowMyTasks |
| layout | Nav Organisation | Non (attendu) | — | Confirmé | idem ; `ShowOrganisationNav` false |
| layout | Lien header Organisation | Non (attendu) | — | Confirmé | `base.html` L49 `not .SimpleUI` |
| layout | Org switcher | Non (1 org) | — | Confirmé | `base.html` L35 `gt len UserOrganizations 1` |
| layout | Badge rôle « Éditeur » | Oui (attendu Non pour P0) | inutile | Confirmé | `base.html` L66 ; HTML `role-badge">Éditeur` |
| layout | DevAuth user switcher | Oui (dev only) | utile (seed) | Confirmé | `base.html` L52–62 — hors prod |
| layout | Déconnexion | Oui | utile | Confirmé | `base.html` L67–70 |
| layout | Footer « Revues — check-lists… » | Oui | utile (marque) | Confirmé | décisions « Marque produit inchangée » ; `base.html` footer |
| layout | Title suffixe « — Revues » | Oui | utile | Confirmé | `base.html` L9 |

---

## Hub `/revues` (Listes en cours)

| écran | élément | présence? | sens | statut | preuve |
|-------|---------|-----------|------|--------|--------|
| `/revues` | H1 « Listes en cours » | Oui | utile | Confirmé | Labels preset ; HTML H1 |
| `/revues` | CTA « Lancer une liste » | Oui | trompeur | Confirmé | `LaunchRunCTA` → « Lancer une liste » ; collision avec onglet **Listes** (templates) — `labels.go` L174–176 + `templates_index` nav |
| `/revues` | Colonne Sujet | Non | — | Confirmé | `runs_list.html` L62 `ShowSubjectColumn` |
| `/revues` | Colonne Statut | Non | — | Confirmé | `runs_list.html` L63 `not .SimpleUI` ; décisions |
| `/revues` | Colonne Échéance | Oui | partiel/inutile si vide | Partiel | HTML th Échéance ; cellules `run-card__due--empty` sur seed |
| `/revues` | Progression | Oui | utile | Confirmé | `runs_list.html` L65–78 |
| `/revues` | Onglets Tous / En cours / Terminées | Oui | utile | Confirmé | décisions accès terminées ; `runs_list.html` L51–55 |
| `/revues` | Recherche + Filtrer | Oui | utile (échelle) | Partiel | OK dès >quelques listes ; bruit si 1–2 lignes |
| `/revues` | Empty → Gérer emails / Mes tâches | Non (SimpleUI branch) | — | Confirmé | `runs_list.html` L15–20 gated `not .SimpleUI` / `ShowMyTasks` |

---

## Wizard lancer + picker

| écran | élément | présence? | sens | statut | preuve |
|-------|---------|-----------|------|--------|--------|
| `/revues/nouvelle` | Étape 1 sujets | Skip auto (1 sujet) | utile | Confirmé | inventaire + redirect → `/subjects/102/modeles?for_run=1` |
| picker `for_run=1` | H1 / crumb « Choisir un modèle » | Oui | trompeur | Confirmé | `breadcrumbs.go` L227 hardcodé ; HTML H1 ; body dit pourtant « liste » |
| picker | Intro « …lancer **la revue** » | Oui | trompeur | Confirmé | `checklist_templates_list.html` L10 |
| picker | Clic liste → POST create run | Oui | utile | Confirmé | inventaire ; `$listUI` path |

---

## Fiche run `/runs/{id}`

| écran | élément | présence? | sens | statut | preuve |
|-------|---------|-----------|------|--------|--------|
| run | H1 sans sujet / sans `#id` | Oui | utile | Confirmé | `handlers.go` L651–654 SimpleUI ; HTML `Check rapide · date` |
| run | Colonne Assigné | Non | — | Confirmé | `ShowAssign=false` ; `run_show.html` L108–135 |
| run | Select statut + commentaire HTMX | Oui | utile | Confirmé | P0 cœur métier |
| run | Lien Détails | Oui | utile | Confirmé | décisions satellite ; `run_item_row_fragment.html` L5 |
| run | Colonne PJ | Oui | utile | Confirmé | décisions P0 « cocher ; CSV » + PJ |
| run | Filtres section/statut | Oui | utile si sections | Partiel | utile dès multi-sections ; léger pour 2 points |
| run | Carte « Clôturer **la revue** » / « Terminer **la revue** » | Oui | trompeur | Confirmé | `run_show.html` L186–196 hardcodé ; preset = liste |
| run | Confirm clôture « …de **la revue** » | Oui | trompeur | Confirmé | `hx-confirm` L190 |
| run | « Lancer une autre **revue** » | Oui | trompeur | Confirmé | `run_show.html` L206 |
| run | « Retour au sujet » | Oui | utile | Confirmé | deep hub sujet ; décisions |
| run | Export CSV / Notion / preuve | CSV si done ; Notion/preuve gated | utile | Confirmé | décisions P3 capability ; pas masqués SimpleUI |
| run | CTA Démarrer (draft) | Non (seed in_progress) | mort (legacy) | Confirmé | décisions ; `run_show.html` L59–67 |

---

## Fiche point `/runs/.../items/...`

| écran | élément | présence? | sens | statut | preuve |
|-------|---------|-----------|------|--------|--------|
| item | Meta « **Revue** : … » + sujet dans label | Oui | trompeur | Confirmé | `run_item_show.html` L5–6 ; BC garde sujet+#id vs H1 run SimpleUI |
| item | Carte Issue Jira (non configuré) | Oui | trompeur | Confirmé | carte toujours rendue ; « contactez un administrateur » alors que user = owner sans nav Organisation — `run_item_show.html` L21–40 |
| item | Carte Pièce jointe | Oui | utile | Confirmé | P0 |
| item | Historique des statuts | Oui | utile | Confirmé | décisions satellite |
| item | « Retour à **la revue** » | Oui | trompeur | Confirmé | `run_item_show.html` L103 |

---

## Listes (templates) `/modeles*`

| écran | élément | présence? | sens | statut | preuve |
|-------|---------|-----------|------|--------|--------|
| `/modeles` | Nav + H1 « Listes » | Oui | utile | Confirmé | `site_nav` SimpleUI ; `!ShowSubjectColumn` |
| `/modeles` | Intro « lancez une **revue** à partir d'une liste » | Oui | trompeur | Confirmé | `templates_index.html` L27 |
| `/modeles` | Colonnes Domaines / Version | Non | — | Confirmé | `templates_index.html` L51–52 |
| `/modeles` | Bouton Importer Notion | Non | — | Confirmé | `not .SimpleUI` L40 ; deep → redirect msg non configuré |
| `/modeles/{id}` | CTA « Lancer cette liste » | Oui | trompeur (collision) | Confirmé | `checklist_template_show.html` L10 SimpleUI ; même mot que instance |
| `/modeles/new|edit` | Form compact listUI | Oui | utile | Confirmé | `checklist_template_form.html` `$listUI` |
| form liste | Hints « dans **la revue** » / « prochaines **revues** » | Oui | trompeur | Confirmé | `checklist_template_form.html` L11, L23 |

---

## Sujet `/subjects*`

| écran | élément | présence? | sens | statut | preuve |
|-------|---------|-----------|------|--------|--------|
| `/subjects/{id}` | Layout revues d’abord (pas Équipes/Membres) | Oui | utile | Confirmé | `subject_show.html` L64–65 `not .ShowCollab` |
| `/subjects/{id}` | Domaines/étiquettes details | Non | — | Confirmé | `ShowSubjectColumn` L30 |
| `/subjects/{id}` | Badge Privé (meta) | Non sans ShowCollab | — | Confirmé | L8 `and private ShowCollab` |
| `/subjects/{id}` | Desc « Vos checklists (**revues**) » | Oui | trompeur | Confirmé | L21 ; HTML |
| `/subjects/{id}` | Colonne Statut runs | Non | — | Confirmé | L74 `not SimpleUI` |
| `/subjects/{id}` | CTA Lancer une liste + Modifier | Oui | utile / collision CTA | Confirmé | LaunchRunCTA |
| `/subjects` liste | Deep hors nav | Deep | partiel | Confirmé | décisions ; BC sujet → `/subjects` ; intro « revues / checklists » L42 |
| form sujet | Domaines (options avancées) | Oui (deep edit) | inutile (mono-sujet) | Partiel | `subject_form` details ; matching domaines peu pertinent P0 |

---

## Deep links hors nav P0

| écran | élément | présence? | sens | statut | preuve |
|-------|---------|-----------|------|--------|--------|
| `/mes-taches` | Page entière | Deep (nav Non) | inutile | Confirmé | matrice Fuite #3 ; empty « points assignés… listes en cours » ; pas d’assign P0 |
| `/admin` hub | Accès owner | Deep (nav Non) | trompeur | Confirmé | matrice Fuite #1 ; hub **complet** Équipes/Politiques/Intégrations — pas hub minimal décisions solo |
| `/admin/users` | Inviter emails | Deep | utile (unlock P1) | Partiel | intentional pour passer whitelist>1 ; discovery faible sans lien |
| `/admin/teams` | Équipes | Deep | inutile P0 | Confirmé | 0 équipe seed ; collab masqué |
| `/admin/settings/policies` | Politiques leads | Deep | inutile P0 | Confirmé | solo |
| `/admin/integrations` | SMTP/Jira/Notion/Webhooks | Deep | partiel (P3) | Partiel | capability P3 OK en principe ; surface org lourde pour particulier |
| `/admin/settings/labels` | Presets sujet + run | Deep | utile (power) | Partiel | seul moyen de changer `listes_en_cours` ; inaccessible sans URL |
| `/modeles/notion-import` | Import | Deep UI masquée | mort/fuite | Confirmé | toolbar masquée ; serveur redirect « pas configuré » |

---

## Top issues (actionnable)

| Rang | Issue | Sens | Effort | Priorité |
|------|-------|------|--------|----------|
| 1 | Collision lexicale **Listes** (templates) ↔ **liste / Listes en cours** (runs) + CTA « Lancer une liste » | trompeur | M | P0 |
| 2 | Hardcode **revue/modèle** sur parcours chaud (clôture, item, picker H1, intros `/modeles`, fiche sujet) | trompeur | S–M | P0 |
| 3 | Carte Jira toujours visible si non configuré + copy « contactez un administrateur » | trompeur | S | P1 |
| 4 | Hub `/admin` deep link = surface org complète (Équipes, Politiques, Intégrations) pour SimpleUI | trompeur | M | P1 |
| 5 | Badge **Éditeur** permanent en header solo | inutile | S | P2 |
| 6 | `/mes-taches` deep link (empty assign) | inutile / fuite | S | P2 |
| 7 | Colonne Échéance toujours affichée (souvent « — ») | inutile | S | P3 |
| 8 | BC / H1 picker « Choisir un modèle » malgré `$listUI` | trompeur | S | P0 (sous-ensemble de #2) |

---

## Critique adverse

### Ce qui tient (ne pas « corriger »)

- **Nav P0 à 2 onglets** et masquage Assign / Mes tâches / Collab / colonne Sujet·Statut : aligné décisions progressive disclosure ; preuves template + HTML.
- **Skip wizard étape 1** à 1 sujet : réduit la friction ; pas un trou.
- **Marque « Revues »** (title, footer) : décisions explicites — ne pas traiter comme incohérence `listes_en_cours`.
- **Fuite routes admin / mes-taches** : disclosure volontaire (« unlock, don’t fork ») ; le bug UX est surtout le **contenu** du hub admin et le copy Jira, pas l’existence de la route.
- **Détails / PJ / historique** : satellite voulu ; pas du bruit collab.
- **Onglets Terminées** : décision produit accès historique — utile même en solo.

### Downgrades (faux positifs potentiels)

| Constat tentant | Pourquoi downgrader |
|-----------------|-------------------|
| « Org owner sans Organisation = bug RBAC » | Non — P0 UI hide ; serveur owner OK (`matrix-roles` Fuite #1 intentionnelle). |
| « Échéance vide = bug » | Colonne stable pour quand une due date existera ; plutôt polish. |
| « Recherche sur hub avec 1 ligne » | Pattern liste uniforme ; coût faible. |
| « Domaines dans form sujet » | Repliés en `<details>` ; coût cognitif limité. |
| « DevAuth switcher » | Outil seed ; hors audit prod. |
| « Collision Listes = mauvais preset » | Preset `listes_en_cours` est **décidé** ; le vrai défaut est l’absence de distinction templates vs instances dans la copy, pas le preset lui-même. |

### Pièges d’audit

- Ne pas confondre **marque produit** Revues et **libellé d’instances** (`Labels.Run`).
- Ne pas exiger masquage serveur des routes admin pour « valider » P0 — mesurer la **découvrabilité** et le **copy** des deep links.
- Seed particulier a déjà un run : empty states SimpleUI (`/subjects/new`) non rejoués ici — statut Partiel si on généralise aux empty states.

### Décisions produit ouvertes (issues #1–2)

1. Comment différencier UI **catalogue** (aujourd’hui « Listes ») vs **instances** (`listes_en_cours`) sans revenir à « Modèles/Revues » en P0 ?  
2. Faut-il un hub `/admin` **minimal** aussi sous `SimpleUI` (Inviter + Mes sujets + Libellés), ou garder power-user deep link complet ?

---

## Périmètre non couvert

- Parcours empty (0 sujet / 0 liste) — seed déjà peuplé  
- Export CSV / preuve / Notion export run (pas de run `done` exercé ici)  
- Browser visuel (MCP tab indisponible) — preuve via HTML HTTP DevAuth  

Fin Wave 2 — persona particulier ; pas d’implémentation.
