# Audit live UI — persona **particulier** (SimpleUI P0)

**Date** : 2026-07-21  
**App** : `http://127.0.0.1:8080` · `REVUES_DEV_AUTH=1` · seed (DB non reset)  
**Login** : `POST /auth/dev/login` `user_id=6` + CSRF → Camille Particulier / org Perso Camille  
**Caps live** (`revues-reports-meta`) : `simple_ui=true` · `show_assign/my_tasks/collab/subject_column=false` · `ui_run_label=listes_en_cours`  
**Méthode** : curl + HTML ; parcours A1–A7 (matrice SimpleUI P0 / plan paliers) + fuites CSV/preuve/assign + transversal E léger  
**Note** : run seed `#103` a été coché, assigné (probe), puis **clôturé** pour exercer CSV/preuve — ne pas reset ici.

Légende **classe** : `OK palier` · `Fuite` · `Manque` · `Bug` · `Copy`  
Légende **sens** : `utile` · `trompeur` · `mort` · `inutile`

---

## Synthèse

Le palier P0 est **globalement tenu** en nav et sur le hub (2 onglets, pas d’assign/tâches/org, CTA lancer, cocher, clôturer, CSV). Les écarts actionnables restent : **collision Listes / liste**, **hardcodes « revue/modèle »** sur le parcours chaud, **carte Jira + « contactez un administrateur »** alors que l’user est owner sans nav Organisation, **hub `/admin` complet en deep link**, et **fuite POST assign** (UI masquée, serveur + DB OK). CSV/preuve sur run `done` = **OK palier** (P0 + P3 capability), pas une fuite de rôle.

---

## Table parcours

| parcours | attendu | constat | classe | sens | preuve |
|----------|---------|---------|--------|------|--------|
| **A1** Chrome nav P0 | Nav « Listes en cours · Listes » ; pas Mes tâches / Organisation / header Org ; marque Revues OK | Confirmé : 2 liens nav ; `ShowMyTasks`/Org absents ; badge **Éditeur** présent ; DevAuth switcher | OK palier (+ Copy badge) | utile (nav) / inutile (badge) | HTML `/revues` `site-nav` L42–47 ; meta `simple_ui:true` ; `base.html` role-badge |
| **A2** Hub `/revues` | H1 Listes en cours ; CTA Lancer ; onglets Tous/En cours/Terminées ; pas colonnes Sujet/Statut/Assigné | Confirmé ; CTA `mb-button` « Lancer une liste » ; th Listes en cours / Échéance / Progression ; recherche « une liste » | OK palier (+ Copy collision CTA↔onglet Listes) | utile / trompeur (collision) | `/revues` 200 6.5 Ko ; toolbar + `list-toolbar__action` ; matrice ShowSubjectColumn/SimpleUI |
| **A3** Wizard lancer | 1 sujet → skip étape 1 ; picker listUI ; vocabulaire liste | Redirect `303` → `/subjects/102/modeles?for_run=1` ; H1/BC **« Choisir un modèle »** ; muted « lancer **la revue** » | Copy | trompeur | `Location: /subjects/102/modeles?for_run=1` ; picker H1 ; `checklist_templates_list` |
| **A4** Run sheet cocher / clôturer | Cocher HTMX ; pas colonne Assigné ; clôturer lead ; confirm | Selects statut OK ; Assigné absent ; « Clôturer/**Terminer la revue** » + `hx-confirm` « …**la revue** » ; flash post-clôture « **Revue** terminée » ; H1 SimpleUI sans sujet | OK palier (actions) + Copy (libellés) | utile / trompeur | `/runs/103` ; POST items 200 ; POST complete → `HX-Redirect` `msg=Revue+terminée` ; DB status `done` |
| **A5** Item satellite | Détails PJ/historique ; Jira si config (P3) ; vocabulaire preset | Meta « **Revue** : … » + BC avec sujet+#id ; carte **Issue Jira** + « contactez un **administrateur** » (Jira off) ; PJ upload OK ; « Retour à **la revue** » | Copy / Manque (CTA config) | trompeur | `/runs/103/items/452` ; cartes Jira/PJ ; user = org owner sans lien `/admin/integrations/jira` |
| **A6** Catalogue `/modeles` | Nav « Listes » ; CRUD ; pas Notion toolbar ; CTA lancer liste | H1 Listes ; Créer ; intro « lancez une **revue**… » ; fiche « Lancer cette liste » ; form « case… dans **la revue** » ; Notion toolbar absente ; deep `/modeles/notion-import` → redirect non configuré | OK palier (caps) + Copy | utile / trompeur | `/modeles` 200 ; show `#29` ; `new` ; notion `303` → `msg=Notion+n'est+pas+configuré` |
| **A7** Deep hors nav P0 | Org/mes-taches masqués nav ; owner peut deep `/admin*` (disclosure) ; sujet hub sans collab | `/mes-taches` 200 empty assign ; `/admin` 200 hub **complet** (Inviter, Équipes, Politiques, Intégrations) + doublon H1/H2 Organisation ; BC intégrations parent **« Admin »**→`/admin/integrations` ; sujet « Chez moi » sans Équipes/Membres/Domaines ; CTA Lancer+Modifier | Fuite (découverte) / Manque (hub minimal) / Copy | trompeur / inutile (équipes/policies solo) | GET 200 admin/* ; `admin_org_hub` CTAs ; subject layout `not ShowCollab` ; matrice Fuite #1/#3 |
| **F-assign** | UI Assign masquée (P0) ; serveur idéalement aligné caps **ou** fuite documentée | Colonne/select absents ; `POST …/assign` **200** + `assigned_to=6` persisté | Fuite | mort (UI) / trompeur (API) | row sans `assign-select` ; POST assign 200 ; SQL `run_items` id 452 `assigned_to=6` ; `AssignItem` gate `CanAssignAccess` only |
| **F-CSV** | Export CSV si run `done` + voir (P0 cœur) | Bouton « Exporter CSV » ; GET 200 attachment CSV (2 lignes ok) | OK palier | utile | `/runs/103/export.csv` ; `Content-Disposition` ; matrice #8 / PLAN P0 |
| **F-preuve** | Preuve si `done` + hash (P3 capability, pas masqué SimpleUI) | Bouton « Télécharger la preuve » ; ZIP 200 ; `evidence_csv_sha256` présent | OK palier | utile | `/export/preuve.zip` ; filename `preuve-revue-103.zip` ; décisions P3 |
| **E1** Badge rôle header | Mute/masquer en Solo P0 (peu informatif) | « Éditeur » toujours affiché | Manque | inutile | `role-badge">Éditeur` sur toutes pages |
| **E2** Colonne Échéance hub | Optionnel si vide | th Échéance présente ; cellules vides/`—` sur seed | Manque | inutile | `/revues` th Échéance |
| **E3** A11y léger | `scope=col`, `aria-current`, flash `role=status` | Hub : 3× `scope=col` ; nav `aria-current` ; flash clôture `role="status"` | OK palier | utile | HTML hub + run done |

---

## Critique adverse (court)

- **Ne pas** traiter CSV/preuve comme fuite SimpleUI : decisions/PLAN = P0 CSV + P3 capability-gated.  
- **Fuite assign** = vrai écart UI↔serveur (confirmé live) ; impact P0 faible (1 membre, pas de UI) mais porte ouverte.  
- **Deep `/admin`** = disclosure volontaire (« unlock, don’t fork ») ; le problème UX est le **hub complet** (Équipes/Politiques) + copy Jira, pas l’existence de la route.  
- Collision **Listes** (catalogue) ↔ **liste / Listes en cours** (instances) = défaut produit ouvert, pas un mauvais preset.  
- Marque footer/title « Revues » = intentionnel.

---

## Top 5 issues → parent

| # | Issue | Classe | Sens |
|---|-------|--------|------|
| 1 | Hardcodes **revue/modèle** sur parcours chaud (clôture, flash, item, picker H1, intros `/modeles`, form liste) malgré `listes_en_cours` | Copy | trompeur |
| 2 | Collision lexicale **Listes** (nav catalogue) ↔ **Lancer une liste** / Listes en cours (instances) | Copy | trompeur |
| 3 | Carte Jira non configurée + « contactez un administrateur » pour owner sans lien config Organisation | Manque / Copy | trompeur |
| 4 | Deep `/admin` = hub org **complet** (pas hub minimal solo) + surfaces Équipes/Politiques inutiles P0 | Fuite / Manque | trompeur |
| 5 | `POST …/assign` autorisé alors que `ShowAssign=false` (DB écrite) | Fuite | mort UI / trompeur API |

---

Fin audit live particulier — read-only code ; pas de commit ; DB seed altérée sur run `#103` uniquement.
