# Critique présence × parcours — persona **alice**

**Date** : 2026-07-19  
**Wave** : 2 — cognitive walkthrough + presence (read-only)  
**Persona** : `alice` — org editor Default, `users.role=editor`, membership `member`, accès sujets ungated via `org_member_legacy` (lead-like)  
**Caps** : `SimpleUI=false` · `ShowAssign` / `ShowMyTasks` / `ShowCollab=true` · `ShowSubjectColumn=true` · `ShowOrganisationNav=false`  
**Vocabulaire** : Revues · Modèles (P2)  
**Sources** : `inventory-screens.md`, `matrix-roles.md`, `decisions.md`, templates / handlers cités  
**Méthode** : inventaire × matrice (colonne alice) + walkthrough buts éditeur ; hors scope RBAC rewrite / cosmétiques sans impact

---

## Synthèse

Alice a une **surface éditeur complète et cohérente** pour le cœur métier (hub Revues → wizard → feuille de run → clôture / assign). Les caps P1/P2 (Mes tâches, Assigné, Modèles, collab fiche sujet) sont **présents comme attendu**. Les frictions majeures sont : (1) **impasse collab → `/admin/teams` (403)**, (2) **Mes tâches → satellite sans saisie**, (3) **rôle UI « Contributeur » vs pouvoirs lead-like legacy**, (4) **seed qui n’assigne rien à alice** (empty Mes tâches). Admin profond correctement **403** (pas de fuite UI vers Organisation).

---

## 1. Profil présence attendu (matrice)

| Surface | Attendu alice | Preuve matrice |
|---------|---------------|----------------|
| Nav Revues / Mes tâches / Modèles | Voir | Nav & hubs · alice |
| Nav Organisation / header Organisation | Masqué | pas org admin |
| Hub `/admin*`, `/admin/subjects`, teams, policies, intégrations | **403** | admin org |
| `/revues` Statut + colonne Sujet | Voir | ShowSubjectColumn · !SimpleUI |
| CTA Lancer / wizard | Faire | CanLaunch |
| Assign colonne + POST | Voir + Faire‡ | ShowAssign ∧ CanLeadAccess legacy |
| Clôturer | Faire‡ | CanComplete legacy |
| Modèles CRUD + CTA Lancer | Faire | editor |
| Notion import toolbar | Voir si Notion config | !SimpleUI |
| Fiche sujet Modifier / collab | Faire‡ / Voir | CanManage legacy · ShowCollab |
| Créer sujet | Faire | CanCreate |
| Visibilité private formulaire | Masqué | CanSetVisibility false (legacy) |
| `/subjects` liste | Voir deep, hors nav | décisions |

‡ = lead-like via `org_member_legacy` (écarts matrice A–B).

---

## 2. Table présence inventaire × alice

Légende : **OK** aligné · **Gap** présence/parcours · **Dette** connu / transition · **N/A** hors persona.

| # | Écran / élément (inventaire) | Présence UI alice | Serveur | Statut | Note |
|---|------------------------------|-------------------|---------|--------|------|
| N1 | `site_nav` Revues | Voir « Revues » | auth | **OK** | |
| N2 | Nav Mes tâches | Voir | auth | **OK** | ShowMyTasks |
| N3 | Nav Modèles | Voir « Modèles » | IndexAll | **OK** | ShowSubjectColumn |
| N4 | Nav Organisation | Masqué | 403 | **OK** | |
| N5 | Header Organisation ghost | Masqué | — | **OK** | `CanManageOrgUsers=false` |
| R1 | `/revues` onglets Tous/En cours/Terminées | Voir | OK | **OK** | |
| R2 | Colonne Sujet + Statut | Voir | OK | **OK** | |
| R3 | CTA Lancer toolbar | Faire | CanLaunch | **OK** | |
| R4 | Empty → `/admin/users` | Masqué | 403 | **OK** | gated `CanManageOrgUsers` |
| R5 | Lien sujet → fiche | Voir | OK | **OK** | |
| W1 | Wizard `/revues/nouvelle` | Faire | OK | **OK** | multi-sujets → pas de skip auto |
| W2 | Étape 2 `for_run=1` clic=POST | Faire | OK | **OK** | BC « Choisir un modèle » |
| W3 | Créer sujet dans wizard | Faire | OK | **OK** | découverte création OK |
| RS1 | Feuille `/runs/{id}` cocher HTMX | Faire | CanCheck | **OK** | hub saisie |
| RS2 | Colonne Assigné + select | Voir + Faire | CanAssign‡ | **OK** | |
| RS3 | Clôturer primary + confirm | Faire | CanComplete‡ | **OK** | |
| RS4 | Export CSV si done | Faire | CanView | **OK** | |
| RS5 | Retour sujet / Lancer autre | Voir | OK | **OK** | |
| IT1 | Fiche point satellite | Voir | OK | **Gap** | pas de statut/commentaire (décision) mais entrée principale depuis Mes tâches |
| IT2 | Jira / PJ | Faire si config / CanCheck | OK | **OK** | P3 capability |
| T1 | `/mes-taches` nav + liste | Voir | OK | **Gap** | seed : 0 assign alice → empty trompeur |
| T2 | Lien point → item show | Voir | OK | **Gap** | 1 hop de plus pour cocher |
| T3 | Colonne Sujet (texte) | Voir | — | **Gap** léger | pas de lien `/subjects/{id}` |
| M1 | `/modeles` index + Créer | Faire | CanManage | **OK** | |
| M2 | Importer Notion | Voir si config | OK / Fuite SimpleUI N/A | **OK** | alice !SimpleUI |
| M3 | Fiche modèle « Lancer avec ce modèle » | Faire | OK | **OK** | → wizard `?template=` |
| S1 | Fiche sujet CTA Lancer + Modifier | Faire | CanLaunch / CanManage‡ | **OK** | EditPath `/subjects/{id}/edit` |
| S2 | Domaines / étiquettes `<details>` | Voir | OK | **OK** | P2 |
| S3 | Bloc Accès + rôle effectif | Voir | — | **Dette** | « Contributeur » + « membre organisation » ≠ pouvoirs lead |
| S4 | Équipes / Membres avant Revues | Voir (ShowCollab) | OK | **Gap** | hiérarchie collab > métier |
| S5 | Form Ajouter équipe (`.button`) | Voir si policy | CanAssignTeams‡ | **Gap** | 2ᵉ primary vs CTA Lancer ; 0 équipes seed |
| S6 | Lien « Créer une équipe » → `/admin/teams` | Voir si 0 équipes | **403** | **Gap** | **impasse** |
| S7 | Inviter membre | Voir (policy défaut) | CanManageMembers‡ | **OK** / partiel | empty members + hint OK |
| S8 | Message « Toutes les équipes… déjà affectées » si 0 équipes | Voir | — | **Gap** | copy fausse |
| S9 | Breadcrumb ancêtre → `/subjects` | Voir | liste OK | **OK*** | *hors nav ; deep intentionnel |
| A1 | Deep `/admin`, `/admin/teams`, labels, policies | Absent nav | **403** | **OK** | sauf fuite lien S6 |
| A2 | `/admin/subjects` | Absent | **403** | **OK** | gestion sujet via `/subjects/{id}/edit` |
| X1 | Visibilité private au create/edit | Masqué | deny | **OK** | CanSetVisibility |
| X2 | Créer sujet hors wizard | Deep `/subjects/new` / liste | OK | **Gap** léger | pas d’entrée nav sujets (décision) |

---

## 3. Cognitive walkthrough (buts alice)

Scénarios seed Default (Portail, API, revue release en cours, 2ᵉ revue sécu).

### CW1 — Lancer une revue (but principal)

| Étape | Action | Feedback | Issue |
|-------|--------|----------|-------|
| 1 | Nav **Revues** | Hub + CTA Lancer | — |
| 2 | CTA → wizard sujet | Liste sujets + recherche | OK multi-sujet |
| 3 | Choisir sujet → modèles | Clic = création | OK (pas d’étape titre — décision) |
| 4 | Atterrissage feuille run `in_progress` | Grille + progression | OK flash omis (décision) |

**Verdict** : parcours fluide. Alice réussit sans admin.

### CW2 — Cocher / assigner / clôturer (run sheet)

| Étape | Action | Feedback | Issue |
|-------|--------|----------|-------|
| 1 | Ouvrir revue en cours | Points + Assigné | Caps OK |
| 2 | Changer statut / commentaire HTMX | Row + progress OOB | OK |
| 3 | Assigner un membre | Select HTMX | OK lead-like |
| 4 | Terminer + confirm | Redirect done + CSV | OK |

**Verdict** : surface la plus solide. Alignée décisions (confirm clôture seulement).

### CW3 — « Mes tâches » (P1 unlock)

| Étape | Action | Feedback | Issue |
|-------|--------|----------|-------|
| 1 | Nav Mes tâches | Empty « Aucune tâche assignée » (seed) | **Seed** : assigns → admin, pas alice |
| 2 | Si tâches : clic point | Fiche satellite (Jira/PJ/histo) | **Pas de cocher ici** |
| 3 | « Retour à la revue » | Feuille pour saisir | Friction 2 écrans |

**Verdict** : concept P1 bon ; **démo seed + entrée satellite** cassent le walkthrough « je traite mes points ».

### CW4 — Préparer un modèle puis lancer

| Étape | Action | Feedback | Issue |
|-------|--------|----------|-------|
| 1 | Nav Modèles → Créer / éditer | Form P2 domaines | OK |
| 2 | Lancer avec ce modèle | Wizard étape 1 (sujet) | OK décision matching |
| 3 | Choisir sujet compatible | Étape 2 | Domaines visibles P2 |

**Verdict** : OK. Notion import seulement si config (P3).

### CW5 — Collaborer sur un sujet (Équipes / Membres)

| Étape | Action | Feedback | Issue |
|-------|--------|----------|-------|
| 1 | Fiche sujet (lien colonne Revues) | **Équipes puis Membres puis Revues** | Métier enterré |
| 2 | Lire « Vos accès » | tags + rôle **Contributeur** | Ment pouvoirs (clôturer/assigner OK) |
| 3 | Ajouter une équipe | « Toutes déjà affectées » + **Créer une équipe** | 0 équipes ; copy fausse |
| 4 | Clic Créer une équipe | **/admin/teams → 403** | **Impasse** |
| 5 | Inviter membre (email) | Form secondary | OK si compte existe ; pas de création équipe |

**Verdict** : ShowCollab révèle une surface **quasi inutilisable** pour alice sans admin org / équipes seed. Pire écart présence de la persona.

### CW6 — Admin profond (négatif)

| Tentative | Résultat | Verdict |
|-----------|----------|---------|
| Nav Organisation | Absent | OK |
| URL `/admin` | 403 | OK |
| Lien UI `/admin/teams` (CW5) | 403 | **Fuite découverte** (lien UI) |

---

## 4. Top issues (actionnable)

| Rang | Issue | Impact alice | Effort | Priorité | Type |
|------|-------|--------------|--------|----------|------|
| 1 | Lien « Créer une équipe » → `/admin/teams` pour non–org-admin | Impasse 403 ; collab P1 morte | S | **P0 UX** | Bug présence |
| 2 | Copy « Toutes les équipes… déjà affectées » quand `len(Teams)==0` et `AvailableTeams` vide | Mensonge UI | S | **P0 UX** | Bug copy |
| 3 | Mes tâches → fiche point **sans** saisie statut | Parcours P1 incomplet | M | **P1** | Parcours / produit |
| 4 | Seed : aucun point assigné à alice | Empty Mes tâches pour persona ShowMyTasks | S | **P1** démo | Seed |
| 5 | `DisplayRole` Contributeur alors que `CanLeadAccess` legacy | Confusion gouvernance | S–M | **P1** | Dette legacy A–B |
| 6 | Layout ShowCollab : Équipes/Membres **au-dessus** des Revues | But éditeur retardé | M | **P2** | Hiérarchie |
| 7 | Deux `.button` plein (Lancer + Ajouter équipe) sur fiche sujet | Charte CTA | S | **P2** | Charte |
| 8 | Colonne Sujet Mes tâches non cliquable | Navigation sujet plus longue | S | **P3** | Polish |
| 9 | Breadcrumb `/subjects` hors nav | Deep OK mais orphelin | — | **P3** | Décision (ne pas « fixer » sans produit) |
| 10 | Accès sujets : pas de hub « mes sujets » en nav | Décision ; wizard + colonne compensent | — | Note | Intentionnel |

---

## 5. Critique adverse

### Ce qui tient (ne pas « corriger »)

| Point | Pourquoi ce n’est pas un bug |
|-------|------------------------------|
| Pas d’onglet Organisation | Alice n’est pas org admin — matrice + décisions |
| Sujets hors nav principale | Décision actée ; accès via colonne / wizard / BC |
| Fiche point = satellite (pas de grille) | Décision explicite ; problème = **entrée** Mes tâches, pas la fiche |
| Admin 403 (pas 404) | Aligné RequireOrgAdmin |
| Legacy lead-like (assign/clôturer) | Aligné RBAC.md « Sujets v1 » court ; divergence = matrice cible équipes |
| Visibilité private masquée | `CanSetVisibility` legacy — correct |
| Empty `/revues` sans CTA admin users | Correctement gated |

### Downgradé / faux positifs

| Tentation d’audit | Pourquoi downgrade |
|-------------------|-------------------|
| « Alice devrait voir Organisation » | Non — 403 serveur ; pas de besoin métier éditeur |
| « Fuite `/mes-taches` sans ShowMyTasks » | N/A alice (flag on) — reste pour particulier |
| « Reader voit `/modeles` » | Hors persona |
| « BC `/subjects` orphelin » | Intentionnel ; Alice peut lister — friction faible vs hub |
| « Deux primaries empty-state onboarding » | Hors alice seed (HasSubjects) |

### Pièges de l’audit

1. **Confondre v1 legacy et matrice cible lead** — Alice « trop puissante » vs contributor pur est **transition documentée** (écarts A–B), pas une régression UI locale.
2. **Juger ShowCollab sur seed sans équipes** — la surface est conçue pour orgs avec teams ; le vrai bug est le **lien admin**, pas l’existence du bloc.
3. **Attribuer l’empty Mes tâches au produit** — d’abord seed (`populateActiveRun` assigne `adminID`) ; ensuite le hop satellite.
4. **Vouloir un nav Sujets pour alice** — contredit décisions ; mesurer plutôt découverte wizard/colonne.

### Ce qui reste ouvert (produit)

| Question | Pourquoi |
|----------|----------|
| Entrée Mes tâches : run sheet ancré `#item` vs saisie sur satellite ? | Décision satellite vs efficacité P1 |
| Éditeur lead-like : libellé rôle / masquer « Contributeur » ? | Fin de vie `org_member_legacy` |
| Collab sans org admin : CTA « demander à un admin » vs cacher Ajouter équipe ? | Évite 403 |
| Ordre blocs fiche sujet (Revues d’abord même en ShowCollab) ? | Tradeoff P1 unlock vs but éditeur |

---

## 6. Confirmation des constats retenus

| # | Constat | Statut | Preuve |
|---|---------|--------|--------|
| 1 | Lien Créer équipe → 403 pour alice | **Confirmé** | `subject_show.html` L178 ; `RequireOrgAdmin` / matrice alice **403** |
| 2 | Copy « déjà affectées » si 0 équipes | **Confirmé** | même branche `AvailableTeams` vide + `len .Teams == 0` |
| 3 | Item show sans contrôles statut | **Confirmé** | `run_item_show.html` ; décisions fiche satellite |
| 4 | Seed n’assigne pas alice | **Confirmé** | `cmd/seed/main.go` `populateActiveRun` → `adminID` |
| 5 | Rôle affiché contributor + lead powers | **Confirmé** | `DisplayRole` + `CanLeadAccess` legacy ; template AccessSources |
| 6 | Ordre Équipes → Membres → Revues | **Confirmé** | `subject_show.html` branche `ShowCollab` |
| 7 | Double primary Lancer + Ajouter | **Confirmé** | `.button` meta + form teams |
| 8 | Nav alice sans Organisation | **Confirmé** (attendu) | `site_nav.html` ; matrice |
| 9 | Parcours lancer/cocher/clôturer OK | **Confirmé** | inventaire + handlers Can* |

---

## 7. Plan PR suggéré (ne pas implémenter ici)

1. **PR collab dead-ends** — copy 0 équipes ; CTA « Créer une équipe » seulement si `CanManageOrgUsers`, sinon message « Demandez à un administrateur de l’organisation ».
2. **PR Mes tâches éditeur** — seed assigne ≥1 point à alice ; lien tâche → `/runs/{id}` (optionnellement `#run-item-…`) ou CTA « Traiter dans la revue » au-dessus du satellite.
3. **PR libellé accès legacy** (optionnel, area access) — ne pas afficher « Contributeur » seul si `CanLeadAccess` via legacy ; ou badge « Accès organisation (transition) ».

---

Fin Wave 2 persona **alice** — read-only ; aucun code modifié.
