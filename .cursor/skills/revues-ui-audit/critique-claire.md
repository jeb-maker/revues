# Wave 2 — Cognitive walkthrough + présence — **claire** (reader)

**Date** : 2026-07-19  
**Persona** : `claire` / claire@example.com — `users.role=reader`, org Default `member`, accès sujets ungated via `org_member_legacy`  
**Caps UI typiques** : `SimpleUI=false` ; `ShowAssign` / `ShowMyTasks` / `ShowCollab` / `ShowSubjectColumn` = **true** ; pas d’onglet Modèles ; pas de `Can*` write  
**Sources** : `matrix-roles.md`, `inventory-screens.md`, `decisions.md`, templates `site_nav` / `runs_list` / `run_show` / `run_item_row` / `subject_show` / `checklist_templates_*` / `templates_index`, handlers `subjects` / `runs` / `checklisttemplates` / `mytasks`  
**Méthode** : walkthrough cognitif (objectifs lecteur) + critique de présence ; confirmation Confirmé / Partiel / Rejeté ; **read-only**.

---

## Synthèse

Claire est globalement **protégée en écriture** (CTAs launch/check/complete/manage absents ; POST → 404 ; admin → 403). La présence « Lecteur » est **faible** : pas de cadrage lecture, grille revue « morte », blocs collab vides avant les revues, et surtout un **écart UI↔serveur** confirmé sur le catalogue `/modeles` (nav masquée, deep link OK). Le copy « Gérez depuis l’onglet Modèles » sur le picker sujet **contredit** l’absence d’onglet pour reader.

---

## Profil surface attendu (matrice)

| Surface | Attendu claire | UI réelle | Serveur |
|---------|----------------|-----------|---------|
| Nav Revues | Voir | Voir | OK |
| Nav Mes tâches | Voir (≥2 membres) | Voir | Auth only — OK pour claire |
| Nav Modèles | **Masqué** | Masqué (`ne .User.Role "reader"`) | `GET /modeles` **200** → **Fuite #5** |
| Nav Organisation | Masqué | Masqué | `/admin*` → **403** aligné |
| CTA Lancer (`CanLaunch`) | Masqué | Masqué | Wizard GET → redirect `/revues` ; POST launch → **404** |
| Cocher / commenter | Masqué | Badges lecture | POST item → **404** |
| Colonne Assigné | Voir lecture (I2) | Colonne + login ou « — » ; pas de select | POST assign → **404** |
| Clôturer / Notion export | — | Absents | **404** / `CanExportNotion=false` |
| CSV / preuve (done) | Faire si Voir | Boutons si done (+ hash preuve) | Aligné lecture |
| Créer / modifier sujet ou modèle | Masqué | Masqué | **404** |
| Fiche sujet collab | Voir (ShowCollab) | Blocs Équipes/Membres lecture | Forms absents |

---

## Walkthrough cognitif

Objectifs typiques d’un **lecteur** seed (suivre une revue, comprendre l’état, exporter, ne pas agir).

### Objectif 1 — Arriver sur le hub et trouver une revue

| Étape | Action claire | Feedback UI | Goal? | Note |
|-------|---------------|-------------|-------|------|
| 1.1 | Post-login → `/revues` | Liste + onglets Tous / En cours / Terminées ; **pas** de CTA Lancer | Oui | Aligné décisions hub |
| 1.2 | Liste vide (filtre / pas de runs) | « Demandez à un éditeur d’en lancer… » | Oui | Message reader-aware Confirmé |
| 1.3 | Clic titre run | `/runs/{id}` | Oui | Deep métier OK |

**CTAs qui ne doivent pas apparaître** : toolbar « Lancer une revue », empty-state Lancer, « Gérer les emails ». **Confirmé absents** (`CanLaunch` / `CanManageOrgUsers` false).

### Objectif 2 — Comprendre une revue en cours (lecture seule)

| Étape | Action claire | Feedback UI | Goal? | Note |
|-------|---------------|-------------|-------|------|
| 2.1 | Ouvre run `in_progress` | H1 + progression sous Points ; **pas** de badge « en cours » | Partiel | Cohérent charte ; **aucun** libellé « lecture seule » / rôle |
| 2.2 | Lit statut / commentaire | Badges + texte ; **pas** de `<select>` / textarea HTMX | Oui | `CanCheck=false` |
| 2.3 | Colonne Assigné | Login ou « — » | Oui | ShowAssign on ; `CanAssign=false` — matrice I2 Confirmé |
| 2.4 | Cherche clôture | Section absente | Oui | `CanComplete=false` |
| 2.5 | « Lancer une autre revue » | Lien absent | Oui | `CanLaunch=false` |
| 2.6 | Lien Détails → item | Satellite PJ / Jira / historique | Partiel | Upload / Jira forms absents (`CanUpload` / `CanLinkJira` false) — OK ; section Jira vide reste visible |

**Piège présence** : la grille ressemble à l’UI éditeur sans contrôles — l’utilisateur doit **déduire** l’impossibilité d’agir. Pas de empty-state alternatif ni bandeau.

### Objectif 3 — Revue terminée : export

| Étape | Action claire | Feedback UI | Goal? | Note |
|-------|---------------|-------------|-------|------|
| 3.1 | Run `done` | Carte « Revue terminée », note, NOK | Oui | |
| 3.2 | Exporter CSV | Bouton secondary toujours si done | Oui | Matrice #8 — lecture large intentionnelle |
| 3.3 | Preuve ZIP | Si hash scellé | Oui | P3 capability |
| 3.4 | Export Notion | Absent (`CanExportNotion` = lead) | Oui | |

### Objectif 4 — Contexte sujet (deep link depuis colonne Sujet)

| Étape | Action claire | Feedback UI | Goal? | Note |
|-------|---------------|-------------|-------|------|
| 4.1 | `/subjects/{id}` | Pas de Lancer / Modifier | Oui | |
| 4.2 | Ligne « Vos accès » | Tags + **« rôle effectif : Contributeur »** (legacy) | **Non** | Conflit avec badge header **Lecteur** — voir Top issues |
| 4.3 | Blocs Équipes / Membres | Vides (seed) **avant** liste Revues | Partiel | ShowCollab on ; bruit cognitif pour reader |
| 4.4 | Liste revues du sujet | Liens runs | Oui | |

### Objectif 5 — Mes tâches

| Étape | Action claire | Feedback UI | Goal? | Note |
|-------|---------------|-------------|-------|------|
| 5.1 | Nav « Mes tâches » | Présente | Oui | ShowMyTasks — pas une fuite pour claire |
| 5.2 | Liste / empty | Liens item/run lecture ; empty → Voir les revues | Oui | Route auth only ; claire ne peut pas cocher depuis la tâche |

### Objectif 6 — Découvrir / lancer un modèle (hors parcours principal)

| Étape | Action claire | Feedback UI | Goal? | Note |
|-------|---------------|-------------|-------|------|
| 6.1 | Cherche onglet Modèles | **Absent** | Oui (décision) | |
| 6.2 | Deep `GET /modeles` | Index catalogue, **sans** Créer / Importer | **Fuite** | Serveur n’applique pas deny reader — matrice #5 Confirmé |
| 6.3 | Fiche `/modeles/{id}` | Points en lecture ; **pas** Lancer / Modifier | Partiel | Lecture catalogue OK côté flags ; découverte non voulue |
| 6.4 | Deep `/revues/nouvelle` | Redirect soft → `/revues` | Partiel | Matrice disait 404 ; code = SeeOther (pas de page erreur) |
| 6.5 | Deep `…/modeles?for_run=1` | **404** | Oui | `CanContributeAccess` deny |
| 6.6 | Deep `…/modeles` **sans** `for_run` | Liste + copy « Gérez les modèles depuis l’onglet **Modèles** » + liens `/modeles/{id}` | **Non** | Onglet inexistant pour claire — copy trompeur Confirmé |

### Objectif 7 — Admin / gouvernance

| Étape | Action claire | Feedback UI | Goal? | Note |
|-------|---------------|-------------|-------|------|
| 7.1 | Nav Organisation | Absente | Oui | |
| 7.2 | Deep `/admin` | **403** | Oui | Aligné (pas 404) |

---

## CTAs / copy qui ne devraient pas apparaître (claire)

| # | Élément | Où | Statut | Preuve |
|---|---------|-----|--------|--------|
| C1 | CTA Lancer (toolbar / empty / sujet / modèle / « autre revue ») | Listes & fiches | **Absent** — OK | `CanLaunch` / `CanContributeAccess` |
| C2 | Modifier sujet / modèle / Créer | Fiches & toolbars | **Absent** — OK | `CanManage` / `CanManageGlobal` |
| C3 | Selects statut / commentaire / assign / clôture / Jira / PJ upload | Run + item | **Absents** — OK | `CanCheck` / `CanAssign` / `CanComplete` / `CanLinkJira` / `CanUpload` |
| C4 | Primary empty « Voir les modèles » (wizard for_run) | `checklist_templates_list` | N/A (404 avant render for_run) | handler `List` for_run |
| C5 | Lien copy « onglet Modèles / Listes » | `checklist_templates_list` mode lecture | **Présent si deep link** — **à ne pas montrer** (ou reformuler) | L.24–26 template |
| C6 | Intro index modèles « lancez une revue… » | `templates_index` mono-sujet | N/A claire (multi-sujet) | L.26–28 ; hors cas claire seed |
| C7 | Empty admin « Gérer les emails » | `/revues` | **Absent** — OK | `CanManageOrgUsers` |

---

## Deep links (carte claire)

| URL | Nav? | Réponse claire | Risque UX |
|-----|------|----------------|-----------|
| `/revues`, `/runs/{id}`, `/runs/{id}/items/{id}` | Hub / métier | 200 lecture | Présence « shell éditeur » |
| `/mes-taches` | Oui | 200 | OK |
| `/subjects`, `/subjects/{id}` | Hors nav | 200 | BC ancêtre `/subjects` ; collab noise |
| `/subjects/{id}/modeles` | Non | 200 + copy onglet Modèles | **Trompeur** + pont vers Fuite #5 |
| `/subjects/{id}/modeles?for_run=1` | Non | **404** | Deny dur OK ; UX « n’existe pas » |
| `/revues/nouvelle` | Non | **302 → `/revues`** | Soft deny (≠ 404 matrice) |
| `/modeles`, `/modeles/{id}` | **Non** (reader) | **200** lecture | **Fuite #5** catalogue |
| `/modeles/new`, `…/edit`, notion-import | Non | **404** | Aligné |
| `/admin*` | Non | **403** | Aligné |
| POST check / assign / complete / launch | — | **404** | Aligné deny métier |

---

## Constats confirmés (passe transversale)

| # | Passe | Constat | Statut | Preuve |
|---|-------|---------|--------|--------|
| 1 | Parcours / RBAC | Nav Modèles masquée pour reader mais `IndexAll` sans deny | **Confirmé** Fuite #5 | `site_nav.html` L.13–16 ; `checklisttemplates.IndexAll` ; matrice #5 |
| 2 | Parcours | Copy picker sujet « Gérez depuis l’onglet Modèles » alors que l’onglet n’existe pas pour claire | **Confirmé** | `checklist_templates_list.html` L.24–26 |
| 3 | Présence | `DisplayRole` legacy → « Contributeur » sur fiche sujet vs badge header « Lecteur » | **Confirmé** | `subjects.DisplayRole` + `AccessSourceOrgMemberLegacy` ; `formatRole` ; `subject_show` L.23–28 |
| 4 | Présence | Fiche revue in_progress : aucun signal « lecture seule » | **Confirmé** | `run_show` + `run_item_row` branche else badges |
| 5 | Composants | Colonne Assigné visible en lecture (pas de contrôles) | **Confirmé** (choix I2) | `$showAssign` ; `CanAssign=false` |
| 6 | Parcours | Blocs Équipes/Membres vides avant Revues (ShowCollab) | **Confirmé** bruit | `subject_show` branche `ShowCollab` |
| 7 | Parcours | Empty `/revues` reader : message éditeur, pas de faux CTA | **Confirmé** OK | `runs_list.html` L.87 |
| 8 | RBAC | Write CTAs absents ; POST → 404 ; admin → 403 | **Confirmé** OK | service `Can*` ; `rbac_test` reader |
| 9 | Parcours | Wizard GET reader → redirect `/revues` (pas 404) | **Partiel** vs matrice | `WizardNouvelle` L.682–684 |
| 10 | A11y / i18n | Badge rôle header `formatRole` Lecteur | **Confirmé** OK | `base.html` role-badge |
| 11 | Décisions | Masquer Modèles pour reader = intentionnel | **Confirmé** (ne pas re-litiger) | `decisions.md` « Modèles pour lecteurs » |
| 12 | CSV lecture | Export done accessible reader | **Confirmé** OK / intention | matrice #8 ; `run_show` post-closure |

---

## Critique adverse

### Ce qui tient

- **Plafond reader côté write est réel** : UI et serveur alignés sur launch / check / assign / complete / manage / admin. Ce n’est pas du security-through-obscurity pour les mutations.
- **Messages empty hub** (« Demandez à un éditeur ») sont adaptés — meilleur que des CTAs morts.
- **Fuite #5 est une fuite de *lecture catalogue***, pas d’écriture : Créer / edit / import restent 404. Risque = découverte + copy trompeur, pas élévation de privilège.
- **Colonne Assigné lecture** (I2) : utile pour un lecteur qui suit qui fait quoi ; ne pas confondre avec contrôle éditable.
- **Deny 404 métier** : cohérent IDOR projet ; la confusion « page cassée » vs « pas le droit » est un trade-off déjà acté (matrice #6).

### Ce qui est downgradé / faux positif

- **« Reader ne devrait jamais voir `/modeles` »** comme bug sécurité P0 → **downgrade** : décisions = masquage nav ; lecture catalogue peut être un choix produit à trancher (fermer serveur **ou** assumer deep link et corriger le copy).
- **Absence de CTA Lancer** ≠ bug présence : c’est le contrat reader.
- **Mes tâches visible** pour claire ≠ Fuite ShowMyTasks (celle-ci vise SimpleUI P0).
- **Legacy « Contributeur »** : dette modèle d’accès v1, pas un faux CTA ; impact UX réel toutefois.

### Pièges de l’audit

- Ne pas exiger le même parcours que alice (lancer / cocher) : échec attendu.
- Ne pas traiter SimpleUI / Notion-import / Organisation header (fuites #1–4) comme problèmes claire — hors persona.
- Seed sans équipes : les blocs collab vides surexposent le bruit ; avec équipes peuplées la lecture collab peut être légitime pour un observateur.

---

## Top issues actionnables (claire)

| Rang | Issue | Type | Effort | Priorité |
|------|-------|------|--------|----------|
| 1 | **Fuite #5** : deny ou assumer `GET /modeles` pour `reader` (aligner nav ↔ serveur) | UI↔RBAC | S–M | P1 |
| 2 | Reformuler / conditionner le copy « onglet Modèles » sur `checklist_templates_list` si `User.Role==reader` (ou si `!CanManage`) | Copy / parcours | S | P1 |
| 3 | Afficher le plafond global sur fiche sujet : ne pas dire « Contributeur » si `users.role=reader` (ex. « Observateur (plafond lecteur) » / cacher DisplayRole legacy) | Présence / i18n | S | P1 |
| 4 | Signal lecture seule sur `/runs/{id}` in_progress (ligne muted / role=status) quand `!CanCheck && !CanComplete` | Présence | S | P2 |
| 5 | Ordre fiche sujet ShowCollab : **Revues d’abord**, Équipes/Membres ensuite (surtout si listes vides) — aide claire + lecteurs | Parcours | S | P2 |
| 6 | Harmoniser deny wizard GET reader : 404 vs redirect (doc matrice + UX) | Cohérence | XS | P3 |
| 7 | Section Jira « Aucune issue » toujours visible en lecture — optionnellement replier si `!CanLinkJira && !JiraLink` | Présence | XS | P3 |

---

## Décisions produit ouvertes

1. **Reader × catalogue modèles** : lecture deep-link **autorisée** (corriger copy + éventuellement entrée secondaire) **ou** `IndexAll` / `Show` → 404 pour reader (nav déjà masquée) ?
2. **Rôle effectif affiché** sous plafond `reader` : montrer le grant sujet, le plafond global, ou les deux ?
3. **Bandeau lecture seule** sur revue : oui / non (charte minimaliste vs clarté persona) ?

Si hors scope Wave 2 claire : fuites SimpleUI (#1–4), legacy lead alice (A–D).

---

## Plan PR suggéré (ne pas implémenter ici)

1. **PR A — Reader catalogue** : trancher décision 1 ; deny serveur **ou** copy + tests `GET /modeles` reader ; fix L.24–26 `checklist_templates_list`.
2. **PR B — Présence lecteur** : DisplayRole sous plafond reader ; optionnel bandeau run + reorder collab/revues sur `subject_show`.

---

Fin Wave 2 persona **claire** — critique seule ; aucune correction implémentée.
