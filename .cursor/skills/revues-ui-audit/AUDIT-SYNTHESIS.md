# Audit UX/UI Revues — inventaire × rôles × walkthrough — 2026-07-19

## Synthèse

Vague 1–2 confirme un cœur métier solide (hub → wizard → feuille HTMX → clôture) et un plafond **reader** write-safe, mais une dette de **présence** : collision lexicale P0 « Listes » (catalogue vs instances), impasses collab éditeur (lien équipes → 403 + copy mensongère), et chrome admin / lecteur mal cadré (BC « Admin », DisplayRole legacy, copy « onglet Modèles »). Les fuites UI↔serveur restantes sont surtout de **découverte lecture** (catalogue reader, hub `/admin` SimpleUI) — à trancher produit, pas à confondre avec les gates write déjà alignés.

Hors synthèse (déjà traités / reportés hors vague) : gates Wave2 CanLaunch / wizard reader / complete OOB / Notion importer ; bug parse HTMX row ; CSV/preuve rôle — **P0 différé** (voir décisions ouvertes).

---

## Constats consolidés

Dédup across personae. Statut = confirmation Wave 2 (code + walkthrough). Préférer présence confirmée aux fuites disclosure intentionnelles.

| # | Source persona(s) | Constat | Classe | Statut | Preuve |
|---|-------------------|---------|--------|--------|--------|
| 1 | particulier | Collision lexicale **Listes** (templates `/modeles`) ↔ **liste / Listes en cours** (instances) + CTA « Lancer une liste » | Copy | Confirmé | `labels.go` LaunchRunCTA ; `site_nav` SimpleUI ; `templates_index` ; critique-particulier §Top #1 |
| 2 | particulier | Hardcode **revue / modèle** sur parcours chaud malgré preset `listes_en_cours` (clôture, item, picker H1, intros modèles, fiche sujet) | Copy | Confirmé | `run_show.html` L186–206 ; `run_item_show` ; `breadcrumbs.go` L227 ; `checklist_templates_list` L10 ; `checklist_template_form` |
| 3 | alice | Lien « Créer une équipe » → `/admin/teams` pour non–org-admin → **403** (impasse ShowCollab) | Bug | Confirmé | `subject_show.html` ; `RequireOrgAdmin` ; critique-alice CW5 / #1 |
| 4 | alice | Copy « Toutes les équipes… déjà affectées » quand `len(Teams)==0` et `AvailableTeams` vide | Copy | Confirmé | même branche `subject_show` ; critique-alice #2 |
| 5 | admin (+ inventaire) | Fil d’Ariane intégrations : parent **« Admin »** → `/admin/integrations` au lieu d’**Organisation** → `/admin` | Fuite* / Copy | Confirmé | `breadcrumbs.go` PathAdmin + BCAdmin* ; inventory §5 ; critique-admin A1 |
| 6 | particulier + admin (+ claire partiel) | Messages Jira/Notion « contactez / demandez à un administrateur » sans lien config ; carte Jira visible même non configurée ; pour org admin = cul-de-sac cognitif | Manque / Copy | Confirmé | `run_item_show.html` L21–40 ; `checklist_template_notion_import` ; critique-particulier #3 ; critique-admin A5 |
| 7 | claire (+ matrice #5) | Nav Modèles masquée pour `reader` mais `GET /modeles` (+ show) **200** lecture catalogue | Fuite | Confirmé | `site_nav` `ne Role reader` ; `checklisttemplates.IndexAll` ; matrix Fuite #5 |
| 8 | claire | Copy picker sujet « Gérez depuis l’onglet Modèles » alors que l’onglet n’existe pas pour reader | Copy | Confirmé | `checklist_templates_list.html` L24–26 |
| 9 | claire + alice | `DisplayRole` legacy → **Contributeur** vs badge header Lecteur (claire) ou pouvoirs lead-like (alice) | Copy / Manque | Confirmé | `subjects.DisplayRole` + `AccessSourceOrgMemberLegacy` ; matrix écarts A–B ; critiques claire #3 / alice #5 |
| 10 | alice | Mes tâches → fiche point **satellite** sans saisie statut (hop obligatoire vers feuille) | Manque | Confirmé | `run_item_show` ; décisions satellite ; critique-alice CW3 |
| 11 | alice | Seed : aucun point assigné à alice → empty Mes tâches trompeur pour persona ShowMyTasks | Manque | Confirmé | `cmd/seed/main.go` `populateActiveRun` → `adminID` |
| 12 | particulier (+ matrice #1) | Deep `/admin` SimpleUI = hub **complet** (Équipes, Politiques, Intégrations) sans entrée nav ; ≠ hub minimal décisions solo | Fuite / Manque | Confirmé | matrix Fuite #1 ; `admin_org_hub` ; critique-particulier #4 ; décisions header Organisation |
| 13 | admin | Hub CTAs **Inviter** / **Mes sujets** / **Libellé {sujet}** sous-décrivent ou trompent (whitelist, catalogue org, dual preset `ui_run_label`) | Copy | Confirmé | `admin_org_hub` ; `admin_users` ; `admin_subject_labels` ; critique-admin A2–A4 |
| 14 | alice + claire | Layout ShowCollab : Équipes/Membres **au-dessus** des Revues (bruit reader ; but éditeur retardé) | Manque | Confirmé | `subject_show` branche ShowCollab |
| 15 | claire | Revue `in_progress` : aucun signal « lecture seule » (shell éditeur sans contrôles) | Manque | Confirmé | `run_show` + `run_item_row` branche badges |
| 16 | alice | Deux `.button` plein (Lancer + Ajouter équipe) sur fiche sujet | Copy† | Confirmé | charte « un primary » ; critique-alice #7 |
| 17 | particulier | Badge header **Éditeur** permanent en Solo/P0 (peu informatif) | Manque | Confirmé | `base.html` role-badge ; critique-particulier #5 |
| 18 | particulier (+ matrice #3) | `/mes-taches` deep reachable sans nav ShowMyTasks (souvent empty P0) | Fuite | Confirmé | matrix #3 ; route auth only — **disclosure volontaire** (downgrade contenu) |
| 19 | matrice (+ particulier) | Notion import toolbar masquée SimpleUI ; serveur autorise editor si Notion ready | Fuite | Confirmé | matrix #4 ; UI `not .SimpleUI` — P3 capability vs disclosure |
| 20 | matrice | Assign POST non gated par `ShowAssign` (UI hide only) | Fuite | Confirmé | matrix #2 — API deep ; impact UX faible vs #3–4 |
| 21 | claire | Wizard GET reader → redirect `/revues` (matrice disait 404) | Manque | Partiel | `WizardNouvelle` SeeOther ; cohérence doc seulement |
| 22 | particulier | Colonne Échéance toujours affichée (souvent « — ») | Manque | Partiel | polish ; downgrade adverse |
| 23 | inventaire | `home.html` branche connectée morte ; BC `BCRunWizardLaunch` orphelin ; `/styleguide` absent | Manque | Confirmé | inventory §5 — dette doc / dead code, hors top présence |
| 24 | — | CSV / preuve : `CanView` large (reader inclus) ; pas de gate lead-only | — | Confirmé (intentionnel) | matrix #8 — **P0 produit différé**, pas bug UI vague 2 |

\*Fuite au sens « navigation mentale cassée », pas élévation de privilège.  
†Dette charte, pas bug fonctionnel.

---

## Critique adverse globale

### Downgrader / ne pas « corriger »

| Tentation | Pourquoi |
|-----------|----------|
| Org owner SimpleUI sans onglet Organisation = bug RBAC | Disclosure P0 actée ; serveur owner OK — bug = **contenu** hub deep, pas l’absence de nav |
| Reader ne doit jamais voir `/modeles` = faille sécu P0 | Fuite **lecture catalogue** seulement ; write reste 404 — trancher produit (deny **ou** assumer + copy) |
| Alice « trop puissante » (assign / clôturer) | Legacy `org_member_legacy` documenté (matrice A–B / RBAC v1 court) — pas régression UI locale |
| ShowCollab « inutile » parce que seed 0 équipes | Surface conçue pour orgs avec teams ; bugs = lien 403 + copy, pas le bloc |
| Empty Mes tâches alice = bug produit | D’abord seed (`adminID`) ; ensuite hop satellite |
| Marque footer / title « Revues » vs preset listes | Décision explicite — ne pas confondre marque et `Labels.Run` |
| Colonne Assigné visible pour claire | Matrice I2 : utile en lecture ; pas un contrôle mort |
| Deny métier → 404 | Trade-off IDOR acté (matrice #6) |
| Chemins `/admin/settings/*` vs `/integrations/*` | Dette URL ; préférer BC/labels avant rename breaking |
| A11 hub `CanManageOrgUsers` sur Intégrations | Défense template OK — ne pas ouvrir PR « simplifier le if » |
| Recherche hub / domaines form sujet / DevAuth | Bruit faible ou hors prod |

### Disclosure intentionnelle (ne pas lister comme P0 sécu)

- Routes admin / mes-taches / modèles accessibles en deep alors que nav masquée (« unlock, don’t fork »).
- CSV / preuve / Jira / Notion export : capability ou `CanView`, **pas** masqués par SimpleUI (P3).
- `/subjects` hors nav org classique.

### Faux positifs / pièges d’audit

- Relitiger décisions actées (satellite fiche point, 2 étapes wizard, pas flash création, onglets Terminées, CTA placement).
- Exiger le parcours alice chez claire (échec write attendu).
- Confondre personae : solo ≠ SimpleUI ; particulier ≠ alice ; Fuite ShowMyTasks = P0 seulement.
- Rehasher gates Wave2 déjà fixés (CanLaunch, wizard reader, complete OOB, Notion importer) ou bug parse HTMX row — hors livrable présence actuel.
- Attribuer à l’UI le P0 CSV/preuve rôle tant que le produit ne tranche pas (#24).

---

## Top 10 actionnable

| Rang | Action | Effort | Priorité | Personas touchés |
|------|--------|--------|----------|------------------|
| 1 | Différencier copy catalogue vs instances en P0 (ex. Listes ≠ Listes en cours) + remplacer hardcodes revue/modèle par `Labels.*` sur parcours chaud | M | P0 | particulier (alice/claire secondaires sur hardcodes) |
| 2 | Collab dead-ends : copy 0 équipes ; CTA « Créer une équipe » seulement si `CanManageOrgUsers`, sinon « Demandez à un admin org » | S | P0 | alice |
| 3 | BC intégrations : parent **Organisation** → `/admin` ; drop libellé « Admin » | S | P0 | admin |
| 4 | Deep link « Configurer Jira/Notion » depuis messages non-config **si** `CanManageOrgUsers` ; sinon garder copy actuelle ; optionnellement replier carte Jira vide | S | P0–P1 | admin + particulier (+ claire polish) |
| 5 | Reader × catalogue : deny `IndexAll`/`Show` **ou** assumer deep link + fix copy « onglet Modèles » + tests | S–M | P1 | claire |
| 6 | DisplayRole sous plafond `reader` / legacy lead : ne pas afficher « Contributeur » trompeur (plafond Lecteur ou badge transition) | S | P1 | claire + alice |
| 7 | Mes tâches éditeur : seed assigne ≥1 point à alice ; lien → feuille run (`#run-item-…`) ou CTA « Traiter dans la revue » | S–M | P1 | alice |
| 8 | Hub admin copy : Inviter → Emails autorisés ; Mes sujets → Sujets ; H1/nav Libellés (sujet + instances) ; option SimpleUI hub minimal | S–M | P1 | admin + particulier |
| 9 | Fiche sujet ShowCollab : **Revues d’abord** ; signal lecture seule sur run si `!CanCheck` | S | P2 | claire + alice |
| 10 | Polish P0 chrome : badge Éditeur (masquer/mute SimpleUI) ; colonne Échéance vide ; double primary équipes | S | P2–P3 | particulier + alice |

---

## Décisions produit ouvertes

1. **Catalogue vs instances en P0** — Comment nommer UI le catalogue (`/modeles`) vs les instances (`listes_en_cours`) sans revenir à Modèles/Revues ? (critique-particulier)
2. **Hub `/admin` sous SimpleUI** — Hub **minimal** (Inviter + Mes sujets + Libellés) vs deep link power-user complet actuel ?
3. **Reader × `/modeles`** — Lecture deep-link autorisée (corriger copy ± entrée secondaire) **ou** 404 serveur aligné sur nav masquée ?
4. **CSV / preuve (P0 différé)** — Restreindre export aux leads / contributeurs, ou garder lecture large (`CanView`) pour readers ?
5. **Mes tâches × satellite** — Ancrer vers la feuille (`#item`) vs permettre la saisie sur la fiche point ?
6. **Rôle effectif affiché** — Grant sujet, plafond global `users.role`, ou les deux (surtout reader + legacy) ?
7. **Bandeau lecture seule** sur revue — Oui / non (minimalisme charte vs clarté persona) ?
8. **Ordre blocs fiche sujet** — Revues d’abord même quand `ShowCollab` ?
9. **Hub Organisation** — Reste lanceur plat ou devient mini-dashboard (compteurs, statut intégrations) ?
10. **Labels admin** — Une entrée « Libellés » ou deux (sujet / instances) dans `admin_nav` ?

---

## Plan PR suggéré

2–4 PRs max, ordonnées, **sans implémentation** ici.

1. **PR A — Dead-ends collab + BC admin** (S, P0)  
   Copy 0 équipes ; CTA Créer équipe gated `CanManageOrgUsers` ; BC Organisation sur intégrations.  
   *Personas : alice, admin. Risque faible, gains walkthrough immédiats.*

2. **PR B — Labels P0 + hardcodes Labels.*** (M, P0)  
   Distinguer catalogue / instances ; brancher clôture, item, picker, intros sur presets.  
   *Persona : particulier. Bloque sur décision produit #1 (sinon copy provisoire documentée).*

3. **PR C — Reader + présence accès** (S–M, P1)  
   Trancher `/modeles` reader (#3) + fix copy onglet ; DisplayRole sous plafond ; option bandeau lecture seule + reorder Revues/collab.  
   *Persona : claire (+ alice legacy). Dépend décisions #3, #6, #7.*

4. **PR D — Admin CTAs + Mes tâches démo** (S–M, P1)  
   Liens Configurer Jira/Notion pour org admin ; hub Inviter/Sujets/Libellés ; seed assign alice + lien tâche → feuille.  
   *Personas : admin, alice, particulier (Jira). SimpleUI hub minimal = sous-ensemble si #2 tranché.*

Hors vague PR (ne pas mélanger) : CSV/preuve rôle (#4) ; deny ShowAssign serveur ; Notion SimpleUI deep ; dette `/styleguide` / `home.html` morte.

---

## Index des livrables vague 1–2

| Livrable | Fichier |
|----------|---------|
| Skill / format | [SKILL.md](SKILL.md) |
| Décisions produit actées | [decisions.md](decisions.md) |
| Inventaire écrans & HTMX | [inventory-screens.md](inventory-screens.md) |
| Matrice rôles × caps × fuites | [matrix-roles.md](matrix-roles.md) |
| Walkthrough particulier (SimpleUI P0) | [critique-particulier.md](critique-particulier.md) |
| Walkthrough alice (éditeur legacy) | [critique-alice.md](critique-alice.md) |
| Walkthrough claire (reader) | [critique-claire.md](critique-claire.md) |
| Walkthrough admin (devadmin) | [critique-admin.md](critique-admin.md) |
| **Synthèse (ce fichier)** | [AUDIT-SYNTHESIS.md](AUDIT-SYNTHESIS.md) |

Fin synthèse vague 1–2 — read-only ; aucune correction implémentée.
