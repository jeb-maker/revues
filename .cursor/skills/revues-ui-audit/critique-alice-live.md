# Critique live — persona **alice** (org editor)

**Date** : 2026-07-21  
**App** : `http://127.0.0.1:8080` · `REVUES_DEV_AUTH=1` · login DevAuth `user_id=2` (Alice Martin · Éditeur)  
**Caps live** : `simple_ui=false` · `show_assign/my_tasks/collab/subject_column=true` · `ui_run_label=revues` (org Default)  
**Méthode** : parcours HTTP authentifié (cookie session) ; pas de reset DB / kill serveur / commit  
**Périmètre** : matrice B1–B6 (walkthrough éditeur) + statut / clôture / CSV + sens des actions  
**Classes** : `OK palier` · `Fuite` · `Manque` · `Bug` · `Copy`

---

## Synthèse

Alice (P1/P2 collab) a un **cœur métier aligné palier** : nav Revues / Mes tâches / Modèles, wizard, feuille HTMX (statut + assign), clôture lead-like, CSV sur `done`, admin **403**. Les écarts live majeurs restent les **impasses collab** (copy « déjà affectées » + lien `/admin/teams` → 403), **Mes tâches → satellite sans saisie**, **rôle « Contributeur » vs pouvoirs lead**, et un **sens d’action faible à la clôture** (confirm générique alors que la revue peut rester à 40 % avec points « En attente »).

---

## Table parcours

| parcours | attendu | constat | classe | sens | preuve |
|----------|---------|---------|--------|------|--------|
| **B1** Hub `/revues` + nav | Voir Revues · Mes tâches · Modèles ; pas Organisation ; CTA Lancer ; colonnes Sujet+Statut ; onglets Tous/En cours/Terminées | Nav et caps OK ; CTA `Lancer une revue` ; Sujet visible ; Organisation absente | **OK palier** | Hub éditeur P1/P2 clair ; 1 primary toolbar | `GET /revues` 200 ; meta `ui_caps` show_* true ; nav sans `/admin` ; `mb-button` → `/revues/nouvelle` |
| **B1** Wizard `/revues/nouvelle` | Faire : multi-sujets → choix puis modèles `for_run=1` ; créer sujet possible | Liste ~100 sujets + « Créer et continuer » ; étape 2 H1 « Choisir un modèle », POST `/subjects/{id}/revues` | **OK palier** | Parcours lancer fluide sans admin | `GET /revues/nouvelle` 200 ; `GET /subjects/90/modeles?for_run=1` 200 |
| **B2** Feuille run `in_progress` | Faire cocher/commenter HTMX ; colonne Assigné + select ; pas CSV | Selects `status` + `assignee_id` ; hx-post item/assign ; progression sous Points ; pas d’export | **OK palier** | Surface saisie = hub ; Assigné utile (≥2 membres) | `GET /runs/101` (avant clôture) ; `POST …/items/444` + `/assign` → 200, alice sélectionnée |
| **B2** Assign seed | Persona ShowMyTasks : tâches alice démontrables | Seed assigne surtout `devadmin` ; empty Mes tâches jusqu’à assign manuelle | **Manque** | Nav « Mes tâches » trompeuse en démo seed | Empty `Aucune tâche assignée` avant assign ; select run : `devadmin` selected sur 1/5 points |
| **B3** Mes tâches | Voir liste ; lien vers traitement | Après assign : lien point → `/runs/{id}/items/{itemId}` (satellite) ; colonne Sujet texte **sans** lien `/subjects/{id}` | **Manque** | 1 hop de plus pour cocher ; sujet non navigable | `GET /mes-taches` : `href="/runs/99/items/436"` ; 0 `href="/subjects/…"` |
| **B3** Fiche point (depuis tâches) | Satellite PJ/Jira/histo (décision) ; saisie sur grille | Pas de `status`/`comment` ; « Retour à la revue » ; Jira « contactez un administrateur » | **OK palier**† / **Manque** entrée | †Satellite conforme décisions ; friction = entrée Mes tâches | `GET /runs/101/items/444` : pas de `name="status"` ; retour ghost |
| **B3** Périmètre tâches vs copy empty | Empty : points des revues **en cours** | Query liste : `cr.status != archived` → **inclut `done`** (ex. #101 OK encore listé) | **Copy** / **Manque** | Empty promet « en cours » ; liste garde le terminé | `store/run_items.go` `ListAssignedRunItems` ; empty `my_tasks.html` Labels.Run.Plural **en cours** ; live row `#101` |
| **B4** Modèles | Voir + CRUD ; CTA Lancer ; Notion si config | Index Créer ; show « Lancer avec ce modèle » → wizard `?template=` ; Notion absent (non config) | **OK palier** | Catalogue P2 « Modèles » cohérent | `GET /modeles` / `/modeles/9` / `/modeles/new` 200 |
| **B5** Fiche sujet collab | ShowCollab : Équipes/Membres ; Lancer/Modifier ; pas private | Lancer primary + Modifier secondary ; blocs Équipes(0)/Membres(0) **avant** Revues ; « Vos accès · Contributeur » | **Manque** | Métier enterré sous collab vide ; rôle sous-estime les pouvoirs | `GET /subjects/90` H2 order Équipes→Membres→Revues ; `DisplayRole` legacy → Contributeur |
| **B5** Ajouter équipe (0 équipes org) | Message honnête ; CTA admin seulement si org admin | Copy « Toutes les équipes… déjà affectées » + lien **Créer une équipe** → `/admin/teams` | **Bug** + **Copy** | Mensonge (0 équipes ≠ toutes affectées) + impasse 403 | `subject_show.html` L178 ; `GET /admin/teams` → **403** Forbidden |
| **B5** Inviter membre | Form si policy ; secondary | Form email/rôle + hint ; submit `.button-secondary` | **OK palier** | Collab membre utilisable sans admin teams | même fiche sujet |
| **B6** Admin profond | Masqué nav ; **403** deep | `/admin`, `/admin/teams`, `/admin/users`, labels, integrations, `/admin/subjects` → 403 ; pas d’onglet Organisation | **OK palier** | Plafond éditeur respecté (sauf fuite lien B5) | codes 403 ; nav sans Organisation |
| **B6** Lien UI → admin teams | Ne doit pas offrir une route interdite | Lien rendu pour non–org-admin | **Fuite** | Découverte admin + cul-de-sac | ancre `/admin/teams` dans UI alice |
| **STATUT** Liste `/revues` | Badge omis si `in_progress` ; « Terminée » si done ; colonne Statut visible (!SimpleUI) | Cellule statut vide en cours ; badge Terminée sur done ; filtre onglets OK | **OK palier** | Progression porte le statut « en cours » | `run-card__status` vide sur en cours ; `status-done` sur terminées |
| **STATUT** Fiche run | Badge omis si in_progress ; gardé si done | Avant clôture : pas de badge run ; après : Terminée + badges points | **OK palier** | Aligné décisions | `run_show` live #101/#100 |
| **CLOSE** Clôturer | Faire‡ lead/legacy ; primary + confirm uniquement à la clôture | Section « Clôturer » + `Terminer la revue` `.button` + `hx-confirm` ; POST complete OK (legacy) | **OK palier** | CTA clôture lisible ; confirm présent | form `hx-post="/runs/101/complete"` ; confirm « Confirmer la clôture de la revue ? » |
| **CLOSE** Sens si points ouverts | Signal si progression &lt; 100 % (attendu UX) | Clôture **acceptée à 40 %** (3× En attente) ; confirm **sans** mention des points restants ; redirect `?msg=Revue+terminée` | **Manque** | Action destructive métier trop « légère » vs état réel | `POST /runs/101/complete` → 303 ; fiche : progress 2/5 + badge Terminée + CSV |
| **CLOSE** Copy Labels | Libellés via `Labels.Run` | Hardcode « revue » dans confirm / H2 / bouton / flash | **Copy** | Casse preset `audits` / `listes_en_cours` | `run_show.html` L24, L190, L196 |
| **CSV** Export run `done` | Faire si Voir ; secondary | Bouton « Exporter CSV » `.button-secondary` ; download 200 `text/csv` | **OK palier** | Export au bon moment (pas sur in_progress) ; non primary | `GET /runs/100/export.csv` ; `Content-Disposition` attachment ; absent sur in_progress |
| **CSV** Preuve ZIP | Si hash scellé (P3) | Absent sur #100 (pas de hash) | **OK palier** | Capability-gated | pas de lien `preuve.zip` |
| **SENS** Hiérarchie CTA fiche sujet | 1 `.button` plein | Avec 0 équipes : seul Lancer primary (OK) ; si `AvailableTeams` : **Ajouter** aussi `.button` → double primary | **Copy**† | †Dette charte latente (seed 0 équipes ne la montre pas) | `subject_show.html` L124 vs L174 |
| **SENS** Accès affiché vs pouvoirs | Libellé cohérent avec CanLead (assign/clôturer) | « Contributeur » + assign/clôturer OK | **Copy** | Alice croit lecteur-éditeur alors qu’elle clôture | `subjects.DisplayRole` + `AccessSourceOrgMemberLegacy` ; actions B2/CLOSE |
| **SENS** Visibilité private | Masqué (legacy non CanSetVisibility) | Pas d’option private au edit sujet | **OK palier** | Pas de fausse promesse | `GET /subjects/90/edit` |

† Décision produit actée (ne pas « corriger » le satellite lui-même).

---

## Top issues (live)

| Rang | Issue | Classe | Priorité |
|------|-------|--------|----------|
| 1 | Lien « Créer une équipe » → `/admin/teams` **403** pour alice | Bug / Fuite | P0 |
| 2 | Copy « Toutes les équipes… déjà affectées » quand `Teams==0` et `AvailableTeams` vide | Copy | P0 |
| 3 | Clôture confirmée sans signal des points encore « En attente » (live : 40 % → Terminée + CSV) | Manque | P1 |
| 4 | Mes tâches : lien point → satellite sans saisie (hop obligatoire) | Manque | P1 |
| 5 | `DisplayRole` « Contributeur » vs pouvoirs lead-like legacy | Copy | P1 |
| 6 | Empty Mes tâches dit « en cours » mais liste inclut runs `done` | Copy / Manque | P2 |
| 7 | Hardcodes « revue » à la clôture (hors `Labels.Run`) | Copy | P2 |
| 8 | Seed : peu/pas d’assign à alice → empty trompeur pour ShowMyTasks | Manque | P2 démo |
| 9 | ShowCollab : Équipes/Membres au-dessus des Revues | Manque | P2 |
| 10 | Double primary Lancer + Ajouter équipe (quand équipes dispo) | Copy | P3 |

---

## Critique adverse (live)

| Tentation | Verdict |
|-----------|---------|
| Alice devrait voir Organisation | Non — 403 aligné matrice |
| Clôture &lt;100 % = bug serveur | Non prouvé comme invariant produit ; défaut = **manque de signal UI** |
| Empty Mes tâches = bug produit | D’abord seed/assign ; ensuite hop satellite |
| CSV pour alice = fuite | Non — matrice #8 lecture `CanView` intentionnelle |
| « Audits » vu une fois en nav | Transient / autre session labels ; DB live `ui_run_label=revues` — ne pas figer comme bug alice |

---

## Hors scope noté

- RBAC rewrite fin de `org_member_legacy`
- Implémentation / PR
- Cosmétique sans impact parcours

Fin audit live alice — read-only livrable.
