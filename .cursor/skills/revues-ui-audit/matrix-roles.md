# Matrice rôles × caps × actions — Revues

Wave 1 — audit UI read-only. Sources : `docs/RBAC.md`, `internal/web/middleware/simple_ui.go`, `header_data.go` / `organisation_nav.go` / `org_admin.go`, `cmd/seed/main.go`, `decisions.md`, handlers `Can*` + templates.

Légende cellules :

| Symbole | Sens |
|---------|------|
| **Voir** | Surface / donnée visible (lecture) |
| **Faire** | Action possible (UI + serveur alignés, sauf note) |
| **Masqué** | UI absente (progressive disclosure ou rôle) — route peut rester ouverte |
| **403** | `RequireOrgAdmin` → `Forbidden` |
| **404** | Deny métier / IDOR (souvent `http.NotFound`, pas 403) |
| **Fuite** | Masqué en UI mais serveur autorise (deep link / POST) — ou inverse |

Contexte **v1 seed Default** (sujets `normal` sans `subject_members` / teams) : accès membre = source `org_member_legacy` + rôle effectif `contributor`. `CanLeadAccess` traite ce legacy comme lead (assign / clôturer / gérer sujet) — **plus permissif** que la matrice cible RBAC « editor + contributor ».

---

## Personae seed (référence)

| persona | login / email | `users.role` | org context | membership org | sujets / accès typiques | UICaps typiques | vocabulaire (`ui_run_label`) |
|---------|---------------|--------------|-------------|----------------|-------------------------|-----------------|------------------------------|
| **particulier** | `particulier` / particulier@example.com | `editor` | org **Perso Camille** (`perso-camille`) seule | `owner` (seul membre) | 1 sujet privé « Chez moi » → direct **lead** | `SimpleUI=true` ; `ShowAssign/MyTasks/Collab=false` ; `ShowSubjectColumn=false` ; `SimpleSubjectID` = chez moi | `listes_en_cours` → Nav « Listes en cours » / short « En cours » ; Listes (pas Modèles) |
| **solo** | `solo` / solo@example.com | `editor` | org **Default** | `member` | sujet privé « Mon usage perso » (lead) **+** sujets normal ungated (legacy contributor) | `SimpleUI=false` (multi-membres Default) ; `ShowAssign/MyTasks/Collab=true` ; `ShowSubjectColumn=true` (≥2 visibles) | Default `revues` |
| **alice** | `alice` / alice@example.com | `editor` | Default | `member` | Portail / API… via **legacy** (contributor + lead-like legacy) | comme solo collab : assign/tasks/collab on ; colonne sujet on (seed multi-sujets) | `revues` |
| **bob** | `bob` / bob@example.com | `editor` | Default | `member` | idem alice (legacy) ; auteur d’une revue API | idem alice | `revues` |
| **claire** | `claire` / claire@example.com | `reader` | Default | `member` | legacy **visible** ; plafond reader (pas cocher / lancer) | collab caps on (membres≥2) mais pas d’onglet Modèles ; pas de Can* write | `revues` |
| **admin / devadmin** | seed `admin` (email bootstrap) ; DevAuth login **`devadmin`** (même user GitHub id 1 typiquement) | `admin` | Default **owner** (+ bypass global) | `owner` | tous sujets org active ; bypass `global_admin` | `SimpleUI` **jamais** (early return admin) ; Show* selon structure org | `revues` |

Notes personae :

- **Whitelist** particulier = 1 email → seuil SimpleUI / masquage onglet Organisation.
- **Pas d’équipes seed** : surfaces Équipes / `team_subject_roles` non exercées par les personae ; code + RBAC cible existent.
- Solo ≠ SimpleUI : solo est « membre isolé d’une grosse org », pas le palier P0.

---

## Matrice globale

Colonnes :

1. **reader** — `users.role=reader`, membre org, sujet visible (legacy ou grant viewer+)
2. **editor** — `editor` + membre, accès sujet **sans** org admin (legacy ou contributor/viewer selon grant)
3. **admin org** — `organization_members` owner/admin (+ rôle global indiqué si critique)
4. **admin global** — `users.role=admin`
5. **SimpleUI P0** — persona **particulier** (caps + org owner editor)
6. **org editor (alice)** — seed alice sur sujets ungated Default

Abréviations : L = lecture/export ; W = écriture métier ; — = non ; † = si `users.role` ≥ editor ; ‡ = si lead **ou** `org_member_legacy` (transition v1).

### Navigation & hubs

| Action / surface | reader | editor | admin org | admin global | SimpleUI P0 | alice |
|------------------|--------|--------|-----------|--------------|-------------|-------|
| Nav `/revues` (Labels.Run.Nav) | Voir | Voir | Voir | Voir | Voir (« Listes en cours ») | Voir (« Revues ») |
| Nav Mes tâches (`ShowMyTasks`) | Masqué si &lt;2 membres ; sinon Voir | idem | idem | idem | **Masqué** (1 membre) ; route `/mes-taches` **Fuite** (auth only) | Voir |
| Nav Modèles/Listes | **Masqué** (`role!=reader`) | Voir (Listes si !ShowSubjectColumn sinon Modèles) | Voir | Voir | Voir « Listes » | Voir « Modèles » (multi-sujet) |
| Nav Organisation (`ShowOrganisationNav`) | Masqué | Masqué | Voir si ≥2 membres **ou** whitelist&gt;1 **ou** multi-org ; sinon Masqué + lien header | Toujours Voir | **Masqué** (solo + SimpleUI) ; **pas** de lien header Organisation (`!SimpleUI` requis) | Masqué (pas org admin) |
| Lien header Organisation (solo admin) | — | — | Voir si !ShowOrganisationNav && !SimpleUI | — | **Masqué** | — |
| Hub `/admin` | **403** | **403** | Faire | Faire | Faire (deep link) ; UI nav absente → **Fuite** découverte | **403** |
| Liste `/subjects` (hors nav classique) | Voir (visibles) | Voir | Voir | Voir | Voir (souvent 1 sujet) | Voir |
| `/admin/subjects` | **403** | **403** | Faire | Faire | Faire (owner) | **403** |

### Sujets

| Action / surface | reader | editor | admin org | admin global | SimpleUI P0 | alice |
|------------------|--------|--------|-----------|--------------|-------------|-------|
| Voir fiche sujet | Voir / **404** hors accès | Voir / **404** | Voir tous org | Voir | Voir « Chez moi » | Voir Portail/API (legacy) |
| Créer sujet (`CanCreate` / `CanCreateSubject`) | **404** POST ; UI Masqué | Faire | Faire† | Faire | Faire | Faire |
| Modifier / archiver (`CanManage` / `CanManageAccess`) | — / **404** | Faire si lead **ou** legacy‡ | Faire† (source org_admin) | Faire | Faire (lead privé) | Faire‡ (legacy) |
| Visibilité private (`CanSetVisibility`) | — | lead seulement | Faire† | Faire | Faire (owner) | — (legacy non lead formel) |
| Domaines / étiquettes UI | Voir si ShowSubjectColumn | idem | idem | idem | **Masqué** (mono-sujet) | Voir (P2) |
| Bloc Équipes / Membres (`ShowCollab`) | Masqué si &lt;2 membres | Masqué / Voir selon caps | Voir | Voir | **Masqué** (layout revues d’abord) | Voir |
| Assigner équipe sujet (`CanAssignTeams`) | — | lead + politique | Faire | Faire | N/A UI | lead-like legacy + politique défaut |
| Inviter membre sujet (`CanManageMembers`) | — | lead + politique | Faire | Faire | N/A UI | idem |
| CTA Lancer depuis fiche (`CanLaunch` = contribute) | Masqué / **404** | Faire | Faire† | Faire | Faire | Faire |

### Modèles / listes

| Action / surface | reader | editor | admin org | admin global | SimpleUI P0 | alice |
|------------------|--------|--------|-----------|--------------|-------------|-------|
| Index `/modeles` | **Voir** deep link (**Fuite** vs nav Masqué) ; `CanManage=false` | Voir + CRUD UI | Voir + CRUD† | Voir + CRUD | Voir « Listes » + CRUD | Voir + CRUD |
| Créer / éditer / archiver (`CanManageGlobal` = editor+) | **404** | Faire | Faire† | Faire | Faire | Faire |
| Lien Notion import toolbar | Masqué | Voir si NotionConfigured && !SimpleUI | idem | idem | **Masqué** (`not .SimpleUI`) | Voir si Notion config |
| `/modeles/notion-import` serveur | **404** | Faire si Notion ready | Faire† | Faire | **Fuite** : serveur autorise editor même si UI Masqué | Faire si ready |
| Fiche modèle CTA Lancer (`CanLaunch` = `CanLaunchRun` org) | Masqué | Faire | Faire† | Faire | Faire (« Lancer cette liste ») | Faire |
| Liste modèles pour sujet `?for_run=1` | **404** | Faire | Faire† | Faire | Faire | Faire |

### Revues (runs)

| Action / surface | reader | editor | admin org | admin global | SimpleUI P0 | alice |
|------------------|--------|--------|-----------|--------------|-------------|-------|
| Liste `/revues` + onglets statut | Voir | Voir | Voir | Voir | Voir ; colonne Statut **Masquée** ; sujet **Masqué** | Voir ; Statut+Sujet visibles |
| CTA Lancer (`CanLaunch` liste) | Masqué + message « Demandez à un éditeur » | Faire si sujets ou CanCreate | Faire† | Faire | Faire | Faire |
| Wizard `/revues/nouvelle` | **404** contribute | Faire | Faire† | Faire | Faire | Faire |
| Voir run (`CanViewAccess`) | Voir | Voir | Voir | Voir | Voir | Voir |
| Cocher / commenter (`CanCheck` / `CanUpdateAccess`) | Masqué / **404** | Faire (contributor+) | Faire† | Faire | Faire | Faire |
| Colonne Assigné (`ShowAssign`) | Masqué si &lt;2 membres | Masqué / Voir | Voir | Voir | **Masqué** | Voir |
| Assigner point (`CanAssign` ∧ ShowAssign UI) | — | UI si ShowAssign ; serveur = `CanLeadAccess` ‡ | org_admin **seul** ≠ lead → **404** assign ; UI peut montrer colonne vide | Faire | UI Masqué ; serveur lead → **Fuite** POST | Faire‡ |
| Clôturer (`CanComplete` / `CanLeadAccess`) | — / **404** | Faire‡ lead/legacy | org_admin† sans lead → **404** ; avec lead OK | Faire | Faire (lead) | Faire‡ |
| Export CSV (run `done`) | Faire (si Voir) | Faire | Faire | Faire | Faire | Faire |
| Preuve ZIP (`CanExportEvidence` = done + hash) | Voir+Faire si scellée | idem | idem | idem | idem (P3 capability) | idem |
| Export Notion (`CanExportNotion` = lead-complete ∧ done ∧ pas d’URL) | — | si CanCompleteAccess | idem | Faire | si NotionConfigured | legacy OK |
| PJ upload / download | download si Voir ; upload = CanCheck ∧ in_progress | Faire | Faire† | Faire | Faire | Faire |
| Lien / créer Jira (`CanLinkJira` ∧ `JiraConfigured`) | — | Faire contribute | Faire† | Faire | Faire si Jira config (P3, pas masqué SimpleUI) | Faire |

### Admin org / intégrations

| Action / surface | reader | editor | admin org | admin global | SimpleUI P0 | alice |
|------------------|--------|--------|-----------|--------------|-------------|-------|
| Emails autorisés `/admin/users` | **403** | **403** | Faire | Faire | Faire (owner, deep) ; empty-state CTA si `CanManageOrgUsers` hors SimpleUI paths | **403** |
| Équipes `/admin/teams` | **403** | **403** | Faire | Faire | Faire deep | **403** |
| Libellés `/admin/settings/labels` (`ui_run_label`, `ui_subject_label`) | **403** | **403** | Faire | Faire | Faire deep | **403** |
| Politiques leads | **403** | **403** | Faire | Faire | Faire deep | **403** |
| SMTP / webhooks / Jira / Notion admin | **403** ; UI `CanEncrypt` | **403** | Faire (+ clé chiffrement) | Faire | Faire deep | **403** |
| Sous-nav Intégrations (`CanManageOrgUsers`) | — | — | Voir | Voir | Voir si page admin ouverte | — |

### Caps UI (rappel seuils)

| Flag | Seuil runtime | Effet UI principal |
|------|---------------|--------------------|
| `SimpleUI` | 1 org · 1 membre · ≤1 sujet visible · whitelist ≤1 · **pas** admin global | Nav P0 ; pas assign/tâches/collab/org tab ; Listes ; H1 run sans sujet ; pas colonne statut listes |
| `ShowAssign` / `ShowMyTasks` / `ShowCollab` | ≥2 membres org | Colonne assign ; nav tâches ; blocs équipes/membres fiche sujet |
| `ShowSubjectColumn` | ≥2 sujets visibles | Colonne Sujet `/revues` ; vocabulaire Modèles ; domaines UI |

---

## Caps UI vs RBAC

Écarts **UI hide ≠ server deny** (ou inverse) — confirmés ou fortement suspectés d’après le code.

### Confirmés

| # | Écart | UI | Serveur | Risque |
|---|-------|----|---------|--------|
| 1 | **SimpleUI masque Organisation** alors que particulier est **org owner** | Pas d’onglet ni lien header | `/admin*` → 200 si `RequireOrgAdmin` | **Fuite** découverte admin ; intention P0 possible (décisions) |
| 2 | **ShowAssign=false** mais assign POST non gated par caps | Colonne/contrôles absents | `CanAssignAccess` seul → lead/legacy OK | **Fuite** POST `/runs/{id}/items/{itemId}/assign` (ex. particulier lead) |
| 3 | **ShowMyTasks=false** | Nav absente | `GET /mes-taches` auth only | **Fuite** lecture tâches (souvent vide en P0) |
| 4 | **Notion import** masqué si `SimpleUI` | Bouton toolbar absent | `CanManageGlobal` + Notion ready → OK | **Fuite** deep link import |
| 5 | **Reader** : nav Modèles masquée | Pas d’onglet | `GET /modeles` (IndexAll) **sans** deny reader | **Fuite** lecture catalogue modèles |
| 6 | Deny métier → **404** (pas 403) | CTA Masqué | `http.NotFound` sur contribute/lead/manage | Pas de fuite droits ; UX « n’existe pas » |
| 7 | Admin org routes → **403** | Lien Masqué | `RequireOrgAdmin` | Aligné (pas 404) |
| 8 | **CSV / preuve** : pas de `Can*` role au-delà de `CanViewAccess` | Boutons si done (+ hash pour preuve) | Idem | Aligné lecture ; plus large que « lead only » |
| 9 | **Jira / Notion export run / preuve** : P3 capability-gated, **pas** SimpleUI | Visibles dès config/hash | Idem | Aligné décisions progressive disclosure |

### Transition v1 vs matrice RBAC cible

| # | Écart | RBAC.md cible | Code actuel (sujets ungated) | Impact UI alice/bob |
|---|-------|---------------|------------------------------|---------------------|
| A | Assign / clôturer | editor+**lead** (contributor —) | `org_member_legacy` ⇒ `CanLeadAccess=true` | Alice **peut** assigner et clôturer |
| B | Gérer sujet (edit/archive) | editor+lead (cible) / v1 editor membre | `CanManageAccess` true sur legacy | Alice voit « Modifier » |
| C | Org admin write | lecture libre ; write si editor+ | `CanContributeAccess` oui† ; **`CanLeadAccess` non** sans lead/legacy | Org owner editor peut cocher/lancer mais **pas** clôturer/assigner sans lead — UI `CanComplete`/`CanAssign` false |
| D | Matrice « Sujets v1 » en tête de RBAC | editor membre = lancer/cocher/clôturer | Aligné legacy lead-like pour clôturer | Cohérent v1 court, divergent section « Matrice des actions » cible |

### Inverses (serveur refuse, UI pourrait tromper)

| # | Cas | Note |
|---|-----|------|
| I1 | Org admin **reader** : voit Organisation / sujets | Write CTAs absents (`CanLaunch`/`CanCheck` false) ; POST → 404 — OK |
| I2 | Colonne Assign visible (`ShowAssign`) pour claire | `CanAssign=false` → pas de contrôles éditables ; colonne peut rester (headers via `$showAssign`) — vérifier Wave 2 inventaire |
| I3 | `CanExportNotion` false pour contributor pur (cible) | Sur seed legacy, alice a CanComplete → export Notion possible si config |

### Flags template ↔ helpers

| Flag UI / PageData | Helper / source | Gate serveur associé |
|--------------------|-----------------|----------------------|
| `SimpleUI`, `ShowAssign`, `ShowMyTasks`, `ShowSubjectColumn`, `ShowCollab` | `resolveUICaps` | **Aucun** (disclosure only) |
| `CanLaunch` (liste / wizard / sujet / modèle) | `CanLaunchRun` ou `CanContributeAccess` | même |
| `CanCheck` | `CanUpdateAccess` → contribute | POST item |
| `CanAssign` | `ShowAssign && CanAssignAccess` | POST assign (**sans** ShowAssign) |
| `CanComplete` | `CanLeadAccess` | POST complete |
| `CanManage` (sujet / modèle) | `CanManageAccess` / `CanManageGlobal` | POST update/archive |
| `CanManageOrgUsers` / `ShowOrganisationNav` | org owner/admin ou global admin | `RequireOrgAdmin` |
| `CanCreate` | `CanCreateSubject` | POST subjects |
| `CanManageMembers` / `CanAssignTeams` | policies + lead/org admin | POST members/teams |
| `NotionConfigured` / `CanExportNotion` / `CanExportEvidence` | config / hash / lead-complete | routes export |
| `CanLinkJira` / `JiraConfigured` | contribute + intégration | POST jira-* |
| `CanEncrypt` | clé env | disable submit admin intégrations |

---

## Liens vers inventaire

Wave 2 doit confronter **chaque élément d’inventaire** (nav, listes, fiches, wizard, partials HTMX, admin, empty states) à cette matrice :

1. Pour la persona de test (au minimum **claire**, **alice**, **particulier**, **devadmin**) : cellule attendue Voir / Faire / Masqué / 403|404.
2. Noter tout écart UI↔serveur comme **Fuite** ou **faux positif charte** ; ne pas confondre avec décisions actées dans `decisions.md` (P0–P3, vocabulaire, placement CTA).
3. Prioriser les lignes marquées **Fuite** et les écarts **legacy vs RBAC cible** (A–D) avant les écarts cosmétiques.
4. Inventaire fichier cible suggéré : `.cursor/skills/revues-ui-audit/inventory-*.md` (Wave 2) — référencer les lignes de la « Matrice globale » par action.

Fin Wave 1 — pas d’implémentation.
