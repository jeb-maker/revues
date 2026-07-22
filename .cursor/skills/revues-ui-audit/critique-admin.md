# Critique Wave 2 — persona **devadmin** (org + global admin)

**Date** : 2026-07-19  
**Méthode** : cognitive walkthrough + critique de présence (read-only)  
**Persona** : seed / DevAuth `devadmin` — `users.role=admin`, org Default owner, `SimpleUI=false`, `ShowOrganisationNav=true`, vocabulaire défaut `ui_run_label=revues`  
**Focus** : hub `/admin`, users, subjects admin, labels (`ui_run_label` / `ui_subject_label`), gating intégrations, sens des CTA Notion / Jira / SMTP  
**Sources** : `inventory-screens.md`, `matrix-roles.md`, `decisions.md`, templates `admin_*` / `subjects_list` / `run_*` / `templates_index`, `breadcrumbs.go`, handlers admin + notifications  
**Hors scope** : RBAC serveur (sauf écart UI↔serveur déjà en matrice), polish cosmétique, autres personae

---

## Synthèse

Pour un global admin, l’entrée **Organisation** est claire et le sous-espace admin est dense et actionnable. La présence se fissure sur trois axes : (1) double vocabulaire **Organisation** vs **Admin** dans le fil d’Ariane des intégrations ; (2) libellés hub **Inviter** / **Mes sujets** / **Libellé** qui sous-décrivent ou trompent la tâche réelle ; (3) CTA métier Notion/Jira « contactez un administrateur » sans lien retour config, alors que **devadmin est** cet administrateur. SMTP est bien gated (`CanEncrypt`) mais son effet produit est invisible hors page de test — présence faible post-config.

---

## 1. Cognitive walkthrough (tâches admin)

Légende statut : **OK** · **Friction** · **Échec** · **N/A décision**.

| # | But utilisateur | Étapes attendues | Résultat | Preuve |
|---|-----------------|------------------|----------|--------|
| CW1 | Ouvrir le hub org depuis n’importe quelle page métier | Nav « Organisation » → `/admin` | **OK** — onglet toujours visible (global admin) | `organisation_nav.go` (admin ⇒ true) ; `site_nav.html` |
| CW2 | Comprendre ce qu’on peut gérer depuis le hub | Lire H1 + cartes + CTAs | **Friction** — H1 « Organisation » + carte H2 « Organisation » (doublon) ; 6 liens tous `.button-secondary` sans priorité ; pas de `admin_nav` sur le hub | `admin_org_hub.html` ; `BCAdminOrgHub` |
| CW3 | Autoriser un collègue (whitelist) | Hub « Inviter » → formulaire email + rôle → Enregistrer | **Friction** — CTA **Inviter** mène à **Emails autorisés** (pas d’envoi d’invitation) ; le select « Administrateur » = rôle **global**, pas org admin (intro le dit, libellé hub non) | hub CTA ; `admin_users.html` « rôle global » ; `formatRole("admin")` → Administrateur |
| CW4 | Lister / créer des sujets org | Hub « Mes sujets » → `/admin/subjects` → Créer | **Friction** — « Mes » possessif trompeur (catalogue org) ; sinon parcours OK (`admin_nav`, CTA toolbar) | hub ; `subjects_list.html` `AdminSection=subjects` |
| CW5 | Changer le libellé des instances (`ui_run_label`) | Hub / nav « Libellé » → select instances → Enregistrer → voir nav | **Friction** — H1/BC/nav = « Libellé {sujet} » alors que le formulaire pilote **aussi** `ui_run_label` ; pas d’aperçu ; effet visible seulement après navigation (OK techniquement) | `BCAdminSubjectLabels` ; `admin_subject_labels.html` ; `labels.go` presets |
| CW6 | Configurer Notion puis l’utiliser | Hub Intégrations → Notion → save → import modèles / export revue | **Friction** — overview OK ; BC parent « Admin » ≠ hub Organisation ; après config, CTAs métier apparaissent (`NotionConfigured`) — **OK gating P3** ; si non config : import / messages « demandez à un admin » **sans** lien `/admin/integrations/notion` | `BCAdminNotion` ; `templates_index.html` ; `checklist_template_notion_import.html` ; `run_show.html` |
| CW7 | Configurer Jira puis lier un point NOK | Intégrations → Jira → save → fiche point | **Friction** — même îlot BC « Admin » ; si non config : « contactez un administrateur » sans deep link config (admin bloqué cognitivement) | `run_item_show.html` L38–39 ; `admin_jira.html` |
| CW8 | Configurer SMTP et vérifier l’effet | Intégrations → SMTP → save → test | **Friction** — save/test OK si `CanEncrypt` ; **aucune** surface métier n’expose un CTA « SMTP activé » ; notifications (clôture, assignation) sont silencieuses — succès perçu = seul email de test | `admin_smtp.html` ; `notifications/service.go` (pas d’UI statut) |
| CW9 | Revenir au hub depuis SMTP/Jira/Notion | Fil d’Ariane parent | **Échec partiel** — parent = « Admin » → `/admin/integrations` (overview), **pas** `/admin` ; sortie hub via onglet nav Organisation seulement | `PathAdmin=/admin/integrations` ; `BCAdminSMTP` etc. vs `BCAdminUsers` → Organisation `/admin` |
| CW10 | Déléguer pouvoirs leads | Politiques → checkboxes | **OK** — copie claire + field-hint admins non limités | `admin_lead_policies.html` |

---

## 2. Présence & cohérence (surfaces focus)

| Surface | Présence ressentie | Constats | Statut | Preuve |
|---------|-------------------|----------|--------|--------|
| Nav principale | Forte | Onglet Organisation actif (`ActiveTab=org`) sur tout le sous-arbre admin | Confirmé | handlers `ActiveTab = "org"` |
| Hub `/admin` | Moyenne | Lien farm horizontale ; pas de statut intégrations ni compteurs ; vocabulaire CTA décalé | Confirmé | `admin_org_hub.html` |
| `admin_nav` | Forte | Groupes Gestion / Intégrations + jump select + `aria-current` | Confirmé | `admin_nav.html` |
| Fil d’Ariane Gestion | Aligné | Parent « Organisation » → `/admin` | Confirmé | `BCAdminUsers`, `BCAdminSubjects`, labels, policies… |
| Fil d’Ariane Intégrations | Cassé | Parent « Admin » → `/admin/integrations` ; inventaire Wave 1 déjà signalé | Confirmé | `breadcrumbs.go` L17, L392–414 ; `inventory-screens.md` §5 |
| Labels | Partielle | Page riche (sujet + run) sous titre sujet-centré ; hint mobile « En cours » utile | Confirmé | `admin_subject_labels.html` ; décisions `ui_run_label` |
| `/admin/subjects` | OK | Intro + nav + toolbar Créer ; empty admin pointe emails puis créer | Confirmé | `subjects_list.html` |
| Intégrations overview | OK | Tableau statut Activé/Désactivé + Configurer | Confirmé | `admin_integrations.html` ; `service.go` descriptions |
| Formulaires SMTP/Jira/Notion | Inégale | SMTP/Jira alertes `CanEncrypt` longues ; Notion/Webhooks plus sèches ; Notion form très compact vs Jira | Confirmé | templates admin_* |
| CTA Notion (modèles) | OK gating | Toolbar Importer si `CanManage ∧ NotionConfigured ∧ !SimpleUI` — admin non SimpleUI | Confirmé | `templates_index.html` L40 ; décisions P3 |
| CTA Notion (revue done) | OK gating | Export si config + `CanExportNotion` (global admin = lead bypass) | Confirmé | `run_show.html` ; `CanLeadAccess` admin |
| CTA Jira (fiche point) | Gating OK, copy NOK | Carte toujours visible ; message non-config adresse « un administrateur » sans lien | Confirmé | `run_item_show.html` |
| CTA SMTP ailleurs | Absente | Attendu (pas de surface « envoyer mail ») — mais hub/overview ne le disent pas | Confirmé | pas de CTA métier SMTP hors admin |

---

## 3. Constats confirmés (passe confirmation)

| # | Constat | Passe | Statut | Preuve |
|---|---------|-------|--------|--------|
| A1 | BC intégrations : libellé **Admin** + URL parent = overview, pas hub Organisation | Parcours / présence | **Confirmé** | `PathAdmin`, `BCAdminIntegrations`… |
| A2 | Hub CTA **Inviter** ≠ page **Emails autorisés** (whitelist GitHub) | Libellés | **Confirmé** | `admin_org_hub.html` vs `admin_users.html` |
| A3 | Hub **Mes {sujets}** possessif incorrect pour catalogue org | Libellés | **Confirmé** | hub L17 |
| A4 | Nav/H1 **Libellé {sujet}** occulte le preset **instances** (`ui_run_label`) | Libellés / parcours | **Confirmé** | `BCAdminSubjectLabels` ; form run select |
| A5 | Messages « contactez / demandez à un administrateur » sans lien config pour l’admin lui-même | Parcours CTA | **Confirmé** | `run_item_show.html` ; `checklist_template_notion_import.html` |
| A6 | Chemins hétérogènes `/admin/settings/*` vs `/admin/integrations/*` | Présence IA | **Confirmé** | router + `integrationPath*` |
| A7 | Doublon H1 / H2 « Organisation » sur le hub | Composants | **Confirmé** | hub + `BCAdminOrgHub` |
| A8 | Rôle whitelist **Administrateur** = global admin (puissant) — hint page OK, CTA hub non | Libellés / risque cognitif | **Confirmé** | intro users ; RBAC globaux |
| A9 | SMTP : gating chiffrement OK ; feedback produit post-config faible (test only) | Parcours | **Confirmé** | smtp template ; notifications sans UI |
| A10 | Placeholder « Laisser vide pour conserver » (secrets) sans `.field-hint` — dette charte | Libellés / a11y | **Partiel** | smtp/jira/notion/webhooks placeholders |
| A11 | Hub : Intégrations derrière `CanManageOrgUsers` — redondant pour routes déjà `RequireOrgAdmin` | Composants | **Rejeté** comme bug — défense en profondeur template OK | hub L20–22 |
| A12 | Import Notion masqué SimpleUI — N/A pour devadmin | — | **N/A décision** | décisions P0 / matrice |

---

## 4. Critique adverse

### Ce qui tient (ne pas « corriger » à la légère)

- **Onglet Organisation toujours visible** pour global admin : intention explicite (`showOrganisationNav`) — pas un oubli SimpleUI.
- **Gating P3** Notion/Jira/preuve par config (pas par SimpleUI) : aligné `decisions.md` ; les CTA apparaissent dès que l’intégration est prête — bon modèle « unlock ».
- **`admin_nav` + jump select** : bonne densité pour power user ; `aria-current` correct.
- **Whitelist « rôle global »** : la page users le dit clairement ; le risque est surtout le CTA hub « Inviter », pas le formulaire.
- **SMTP sans CTA métier** : l’absence d’UI « emails activés » n’est pas un bug de gating — les notifs sont fire-and-forget. C’est un problème de **feedback de présence**, pas de droit manquant.
- **Tous les liens hub en secondary** : défendable (hub = lanceur, pas une action unique) — friction de hiérarchie, pas violation stricte « un `.button` plein » (aucun primary du tout).

### Ce qui est downgradé / faux positif possible

- **« Admin » dans BC** : peut être legacy pré-hub Organisation ; impact UX réel (perte du lien hub) > pur naming. Ne pas traiter comme simple typo.
- **Possessif « Mes sujets »** : cosmétique si le reste du parcours est fluide ; prioriser après Inviter / BC.
- **Placeholders secrets** : charte stricte, mais pattern répandu « leave blank to keep » — effort faible, priorité basse vs parcours.
- **A11 `CanManageOrgUsers` sur hub Intégrations** : ne pas ouvrir une PR « simplifier le if ».

### Pièges de l’audit

- Ne pas confondre **org owner editor** (peut admin sans être global) et **devadmin** : la critique « Administrateur = global » pèse surtout quand un owner non-global utilise le même écran.
- Ne pas exiger un CTA SMTP sur `/revues` — hors modèle produit.
- Ne pas relitiger SimpleUI / fuite `/admin` pour particulier (matrice #1) — hors persona.
- Chemins `settings` vs `integrations` : dette IA ; renommer URL = breaking bookmarks — préférer BC/labels d’abord.

---

## 5. Top issues actionnables

| Rang | Issue | Impact persona | Effort | Priorité |
|------|-------|----------------|--------|----------|
| 1 | Unifier BC intégrations : parent **Organisation** → `/admin`, courant Intégrations/SMTP/… (drop libellé « Admin ») | CW6–CW9, présence | S | P0 |
| 2 | Deep link « Configurer… » depuis messages Jira/Notion non configurés si `CanManageOrgUsers` | CW6–CW7 | S | P0 |
| 3 | Hub : renommer **Inviter** → libellé aligné page (ex. Emails autorisés) ; **Mes sujets** → **Sujets** (ou Labels.Subject.Plural) | CW3–CW4 | S | P1 |
| 4 | Labels : H1/nav/BC reflétant les **deux** presets (ex. « Libellés ») ; garder hint `ui_run_label` | CW5 | S | P1 |
| 5 | Hub : retirer H2 « Organisation » redondant ou le remplacer par statut org (membres / intégrations on) | Présence | S | P2 |
| 6 | Harmoniser alertes `CanEncrypt` (longueur + wording) Notion/Webhooks ↔ SMTP/Jira | Présence | S | P2 |
| 7 | Overview ou SMTP : une ligne de présence « utilisé notifications : assignation, clôture, rappels » + lien doc | CW8 | S | P2 |
| 8 | Secrets : `.field-hint` « Laisser vide pour conserver » hors placeholder | Charte | S | P3 |

---

## 6. Décisions produit ouvertes

| Question | Contexte |
|----------|----------|
| Le hub doit-il rester un **lanceur plat** ou devenir un **tableau de bord** (compteurs whitelist, statut 4 intégrations) ? | Présence hub moyenne ; décisions actuelles = hub minimal |
| Faut-il exposer **deux entrées** labels (sujet / instances) dans `admin_nav`, ou un seul « Libellés » ? | CW5 |
| Message non-config Jira/Notion : lien admin seulement si `CanManageOrgUsers`, sinon copy actuelle ? | A5 — recommandation technique claire, validation produit utile |

Sinon : pas de contradiction avec les décisions actées P0–P3 / placement CTA listes.

---

## 7. Plan PR suggéré (sans implémenter)

1. **PR A — Fil d’Ariane & copy admin** : BC Organisation sur intégrations ; hub Inviter/Mes sujets ; H1 Libellés ; optionnel dédoublonnage H2 hub.  
2. **PR B — CTA capability → config** : liens « Configurer Notion/Jira » pour org admin depuis empty/messages métier.  
3. **PR C (optionnel)** : présence SMTP/overview + field-hints secrets + alertes `CanEncrypt` homogènes.

---

Fin Wave 2 persona **devadmin** — read-only ; aucune correction implémentée.
