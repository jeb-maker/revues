# Audit UX/UI live — persona **devadmin** (D1–D3)

**Date** : 2026-07-21  
**Méthode** : walkthrough HTTP + captures Chrome (DevAuth) — pas de reset DB durable, serveur intact  
**Persona** : DevAuth `devadmin` (header « Admin démo » · badge Administrateur · org Default)  
**App** : `http://127.0.0.1:8080` · `REVUES_DEV_AUTH=1`  
**Focus** : D1 hub `/admin` · D2 `ui_run_label` (change réversible + restore) · D3 gating intégrations · sens des CTA  
**Preuves visuelles** : `.cursor/skills/revues-ui-audit/shots-admin-live/*.png`  
**État final** : `organizations.ui_run_label='revues'` (Default) — restauré après test

---

## Synthèse

En live, le hub Organisation est joignable et dense, mais les CTA **Inviter** / **Mes sujets** / **Libellé sujet** sous-décrivent ou trompent la tâche. Le preset **`ui_run_label` fonctionne** (nav + H1 + « Lancer un audit » après passage à `audits`, puis restore). Les intégrations sont correctement **capability-gated** (overview Désactivé, `CanEncrypt` désactive Enregistrer), mais le fil d’Ariane **Admin** et les messages métier « contactez un administrateur » **sans lien config** cassent le sens d’action pour l’admin lui-même.

---

## Constats par passe (D1–D3)

| # | Passe | Constat | Statut | Preuve |
|---|-------|---------|--------|--------|
| L1 | Parcours D1 | Nav **Organisation** → `/admin` ; H1 « Organisation » ; compte `devadmin` | **Confirmé** | `01-admin-hub.png` ; GET `/admin` 200 |
| L2 | Composants D1 | Doublon H1 + carte H2 « Organisation » ; hub = ferme de `.button-secondary` ; **pas** de `admin_nav` | **Confirmé** | hub HTML + shot |
| L3 | Libellés D1 | CTA **Inviter** mène à H1 **Emails autorisés** (whitelist, pas d’envoi) | **Confirmé** | hub → `/admin/users` ; intro « rôle global » |
| L4 | Libellés D1 | CTA **Mes sujets** (possessif) alors que la liste admin = catalogue org « Sujets » | **Confirmé** | hub vs `/admin/subjects` H1 |
| L5 | Libellés D1/D2 | CTA / H1 / BC **Libellé sujet** alors que le formulaire pilote aussi **`ui_run_label`** ; `admin_nav` = « Libellé » | **Confirmé** | `02-labels.png` ; `BCAdminSubjectLabels` |
| L6 | Parcours D2 | POST `ui_run_label=audits` → nav/H1 **Audits**, CTA liste **Lancer un audit** ; flash « Libellés mis à jour » | **Confirmé** | POST 302 `?msg=…` ; HTML `/revues` |
| L7 | Parcours D2 | Restore `ui_run_label=revues` → nav/H1 **Revues** ; DB Default = `revues` | **Confirmé** | POST restore ; `SELECT ui_run_label` |
| L8 | Parcours D3 | Overview : 4 intégrations **désactivé** + liens **Configurer** (SMTP `/admin/settings/smtp`, Jira/Notion `/admin/integrations/…`) | **Confirmé** | `03-integrations.png` ; hrefs |
| L9 | Parcours D3 | BC parent **Admin** → `/admin/integrations` (pas Organisation `/admin`) sur overview / SMTP / Jira / Notion | **Confirmé** | crumbs live ; `breadcrumbs.go` `PathAdmin` |
| L10 | Parcours D3 | Sans `REVUES_ENCRYPTION_KEY` : alerte rouge + **Enregistrer** / test **disabled** (SMTP, Jira, Notion, Webhooks) — gating clair | **Confirmé** | `04-jira-admin.png` ; `05-smtp.png` |
| L11 | CTA D3 | Fiche point : « Jira n'est pas configuré — **contactez un administrateur** » **sans** lien `/admin/integrations/jira` alors que l’utilisateur **est** admin | **Confirmé** | `06-item-jira.png` ; `run_item_show.html` |
| L12 | CTA D3 | Toolbar Modèles : pas d’**Importer** si Notion off (OK) ; deep link `/modeles/notion-import` → redirect flash « Notion n'est pas configuré » **sans** CTA config | **Confirmé** | GET 303→`/modeles?msg=…` ; flash sans lien admin |
| L13 | Présence D3 | Chemins hétérogènes `settings/{smtp,webhooks,labels}` vs `integrations/{jira,notion}` — overview Compense via Configurer | **Confirmé** | router + overview |
| L14 | Composants | Hub Intégrations derrière `CanManageOrgUsers` | **Rejeté** comme bug (défense template) | `admin_org_hub.html` |

---

## Critique adverse

### Ce qui tient

- **Unlock P3** : CTA Notion import absents tant que non configuré ; overview statut + Configurer — bon modèle « capability ».
- **`ui_run_label` bout en bout** : changement immédiat sur nav, H1 liste, CTA lancer — pas de dette technique sur l’effet produit.
- **`CanEncrypt`** : message + boutons disabled = gating honnête (pas de faux succès save).
- **Onglet Organisation** toujours visible pour global admin — aligné décisions / matrice.
- **Restore** effectué : aucune dérive vocabulaire laissée sur Default.

### Downgradé / non bloquant

- **« Mes sujets »** : friction de copy, parcours sinon OK (`admin_nav` + H1 Sujets).
- **Placeholders secrets** / dette charte : hors focus D1–D3 live (env sans clé = forms gelés).
- **SMTP sans CTA métier** : attendu (notifs silencieuses) — présence faible, pas un trou de droit.
- **Tous secondary sur le hub** : défendable (lanceur) — hiérarchie faible, pas violation « deux primary ».

### Pièges

- Ne pas traiter le message Jira « contactez un admin » comme correct pour **claire** et le laisser tel quel pour **devadmin** : le fix est conditionnel (`CanManageOrgUsers` → lien config).
- Ne pas relitiger SimpleUI / fuite `/admin` particulier.
- Autre agent peut toucher les mêmes runs seed (`/runs/101`) — les captures métier restent valides pour le copy Jira.

---

## Top issues actionnables

| Rang | Action | Effort | Priorité |
|------|--------|--------|----------|
| 1 | BC intégrations : parent **Organisation** → `/admin` (drop libellé « Admin ») | S | P0 |
| 2 | Messages Jira / Notion non configurés : lien **Configurer…** si `CanManageOrgUsers` (fiche point, flash `/modeles`, template import) | S | P0 |
| 3 | Hub : **Inviter** → aligné page (ex. Emails autorisés) ; **Mes sujets** → **Sujets** | S | P1 |
| 4 | Labels : H1 / nav / BC **Libellés** (sujet + instances) — le form le dit déjà | S | P1 |
| 5 | Hub : retirer H2 « Organisation » redondant ou le remplacer par statut (whitelist / intégrations on) | S | P2 |
| 6 | Overview ou SMTP : une ligne de présence « emails : assignation, clôture… » | S | P2 |

---

## Décisions produit ouvertes

| Question | Contexte live |
|----------|----------------|
| Hub = lanceur plat ou mini tableau de bord (compteurs / 4 statuts intégrations) ? | Accès rapide sans statut alors que overview les a |
| Flash Notion redirect vs page import avec erreur + lien : quel pattern ? | Aujourd’hui redirect sec sans deep link admin |

Sinon : aucune contradiction avec `decisions.md` P0–P3.

---

## Plan PR suggéré

1. **PR A** — BC Organisation + copy hub (Inviter / Mes sujets / Libellés)  
2. **PR B** — Deep links config depuis empty/messages Jira & Notion (org admin)  
3. **PR C (opt.)** — présence SMTP/overview + dédoublonnage H2 hub

---

## Protocole live exécuté

1. Session DevAuth auto → `devadmin` (switcher header).  
2. D1 : GET `/admin`, users, subjects ; shots hub.  
3. D2 : POST labels `ui_run_label=audits` → vérifier `/revues` ; POST restore `revues` ; vérifier DB.  
4. D3 : GET integrations, smtp, jira, notion ; item `/runs/101/items/444` ; `/modeles/notion-import` ; shots.  
5. **Aucune** migration / reset DB / kill serveur / commit.

Fin audit live **devadmin** D1–D3 — read-only sauf POST labels réversible.
