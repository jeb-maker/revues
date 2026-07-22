# Épique — Roadmap thématique post-cœur

Backlog produit révisé (adoption → SimpleUI → preuve → pilote → vision).  
**Icebox** (séries/campagnes/gouvernance avancée/Slack/Google OAuth/Postgres) : ne pas créer d’issues tant que signal d’usage.

```bash
# Prérequis : gh auth login
./scripts/create-thematic-roadmap-issues.sh          # lots 0–5
./scripts/create-thematic-roadmap-issues.sh --lot 0  # épiques seules
./scripts/create-thematic-roadmap-issues.sh --lot 1  # gates sécu
```

Ordre d’implémentation : **Lot 1 → 2 → 3 → 4 (#66 humain) → 5**.  
Ne pas paralléliser A1↔B*, C1↔C2↔C3, ni toute PR `canonical.sql`.

---

## Priorisation (normative)

### Must-ship

1. Gates sécu : A4 (#62), A5 (tests CSRF), G2 rate limit, A2 whitelist
2. SimpleUI mince : B1 → B2 → B3 → A1 (scindé)
3. Preuve : C1 → C2 (C3 hashes-first ensuite)
4. A3 / #66 PASS (humain)
5. Post-#66 : D1, E3', E6, D6 gated, B4 scindé

### Icebox (pas d’issue GitHub)

D2/D3 moteurs, D4 fusion, D5 rapport, E2/E4/E5, F1–F4, C5/C6, G3/G7, S2 N-1.

### Definition of Eco (chaque issue UI / intégration)

- [ ] HTML/fragment < 50 Ko ; Δ `app.css` ≤ 0 ; CSS total budgets
- [ ] JS parcours ≤ 15 Ko ; ≤ 8 req SQL ; ≤ 8 req navigateur
- [ ] Pas de polling/WS ; pas d’API externe au premier paint
- [ ] Listes paginées ≤ 25 ; ZIP/upload plafonnés
- [ ] SimpleUI = masquage serveur ; matrice paliers si UI

### Template corps

```markdown
## Contexte / Objectif / In / Out (obligatoire)
## Critères d'acceptation + ./scripts/check.sh
## Dépendances + ne pas paralléliser avec
## Revue humaine si area:auth | integrations | chiffrement
### Definition of Eco
(voir checklist ci-dessus)
```

---

## Epic A — `[Epic] Adoption — parcours pilote`

**Labels** : `epic`, `vague-thematic`, `area:ui`, `area:auth`  
**Lot** : 0 (épique) + issues lots 1 / 2 / 4

### Contexte

Débloquer le premier usage réel. Gate officielle : #66 PASS.

### Issues filles

- Lot 1 : A2, A4 (#62), A5 (#64)
- Lot 2 : A1a, A1b, A1c (après B1–B2)
- Lot 4 : A3 (#66) checklist terrain

### Hors scope

- Fancy onboarding (coach marks, tours)
- Second produit « Lite »

---

## Epic B — `[Epic] Progressive disclosure (SimpleUI)`

**Labels** : `epic`, `vague-thematic`, `area:ui`  
**Lot** : 0 + 2 (+ B4 en lot 5, B5 parallèle)

### Contexte

Même produit ; complexité révélée par structure (membres / sujets / capabilities).  
Spec : `.cursor/skills/revues-ui-audit/decisions.md`.

### Issues filles

B1, B2, B3, B0, B5, B6 (lot 2) ; B4a/b/c (lot 5).

### Hors scope

- Toggle mode pro / fork Lite
- Masquer intégrations parce que SimpleUI

---

## Epic C — `[Epic] Conformité & preuve`

**Labels** : `epic`, `vague-thematic`, `area:core`  
**Lot** : 0 + 3

### Contexte

Revue clôturée = artefact traçable. WIP branche `cursor/issue-evidence-zip-f21b`.

### Issues filles

C0 (si besoin), C1, C2 (lot 3) ; C3, C4 (lot 6+ sur signal).

### Hors scope / icebox

- C5 audit admin, C6 concurrency (jusqu’à besoin)

---

## Epic D — `[Epic] Opérationnel qualité`

**Labels** : `epic`, `vague-thematic`, `area:core`  
**Lot** : 0 + 5 (D1, D6) ; D7 opportuniste

### Contexte

Remplacer le reste d’Excel (échéances, vues) sans usine à gaz.

### Issues filles créées

D1, D6 (lot 5) ; D7 optionnel.

### Icebox

D2/D3 moteurs, D4 fusion, D5 rapport — préférer D2' Relancer / D3' wizard ≤20 plus tard.

---

## Epic E — `[Epic] Intégrations v2+ (légères)`

**Labels** : `epic`, `vague-thematic`, `area:integrations`  
**Lot** : 0 + 5  
**Gate** : #66 PASS + rate limit + anti-SSRF verts.  
**Revue humaine** obligatoire.

### Issues filles créées

E3', E6 (lot 5) ; E1' plus tard.

### Icebox

E2 Jira Server, E4 Slack/Teams, E5 Google OAuth.

---

## Epic F — `[Epic] Gouvernance suite (icebox)`

**Labels** : `epic`, `vague-thematic`, `area:auth`, `area:admin`  
**Lot** : 0 seulement — **aucune issue fille** jusqu’à org multi-équipes réelle.

### Hors scope actuel

F1 bulk invite, F2 matrix UX, F3 ownership, F4 demande accès privé.

---

## Epic G — `[Epic] Hardening & dette`

**Labels** : `epic`, `vague-thematic`, `area:infra`, `area:auth`  
**Lot** : 0 + G2 en lot 1 ; G1/G4/G5/G6 opportunistes

### Issues filles créées maintenant

G2 rate limit (lot 1).

### Icebox

G3 antivirus, G7 PostgreSQL.

---

## Issue A2 — `[auth] Message post-OAuth si email non autorisé`

**Labels** : `area:auth`, `vague-thematic`  
**Lot** : 1  
**Bloqué par** : #7

### Objectif

Afficher un message clair quand OAuth réussit mais l’email n’est pas whitelisté, avec moyen de contacter un admin — **sans énumérer** si un email est présent/absent en base.

### In

- Page ou flash post-callback dédié
- Hint contact admin générique (pas « cet email n’existe pas »)
- Test : réponse identique timing/message pour email inconnu vs non whitelisté (autant que possible)

### Out

- Self-service demande d’accès
- Fuite d’existence d’emails / orgs

### Critères d'acceptation

- [ ] Utilisateur non whitelisté voit un message actionnable
- [ ] Pas d’énumération d’emails dans le message ni les logs applicatifs
- [ ] Test table-driven
- [ ] `./scripts/check.sh` vert

### Definition of Eco

Checklist standard (page HTML légère).

---

## Issue A4 — `[auth] Tests OAuth GitHub mockés`

**Labels** : `area:auth`, `vague-thematic`  
**Lot** : 1  
**Note** : correspond à #62 si déjà ouverte — lier, ne pas dupliquer.

### Objectif

Suite de tests mockés : PKCE, `state`, email non vérifié refusé, rotation session, pas de fixation.

### Critères d'acceptation

- [ ] Tests mock HTTP GitHub couvrent happy path + refus email non vérifié
- [ ] Test fixation / rotation session au login
- [ ] `./scripts/check.sh` vert
- [ ] Revue humaine `area:auth`

---

## Issue A5 — `[core] Couverture CSRF HTMX toutes routes mutantes`

**Labels** : `area:core`, `vague-thematic`  
**Lot** : 1  
**Note** : affine #64 — **tests d’abord**, refactor deps ensuite (issue séparée si besoin).

### Objectif

Table-driven `TestCSRF_MissingToken` sur toutes les routes POST/PUT/PATCH/DELETE (y compris HTMX).

### In

- Inventaire des routes mutantes
- Tests qui échouent sans token CSRF

### Out

- Gros refactor « centraliser deps » dans la même PR (faire après si nécessaire)

### Critères d'acceptation

- [ ] Toute route mutante absente du test = échec CI ou liste explicite justifiée
- [ ] `./scripts/check.sh` vert

---

## Issue G2 — `[infra] Rate limit auth, invite, export, webhook`

**Labels** : `area:infra`, `area:auth`, `vague-thematic`  
**Lot** : 1  
**Revue humaine** : auth

### Objectif

Rate limit par IP (+ user/org si authentifié) sur `/auth/*`, invitations whitelist, export evidence, test webhook.

### Critères d'acceptation

- [ ] Limites documentées (env ou constantes)
- [ ] 429 avec message sobre
- [ ] Tests
- [ ] `./scripts/check.sh` vert
- [ ] Ne bloque pas le parcours pilote nominal

### Out

- WAF externe, Redis obligatoire

---

## Issue B1 — `[ui] Flags SimpleUI runtime + tests middleware`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 2

### Objectif

Stabiliser `SimpleUI`, `ShowAssign`, `ShowMyTasks`, `ShowCollab`, `ShowSubjectColumn` + tests des 4 coins (solo, duo mono-sujet, multi-sujet, admin global).

### Critères d'acceptation

- [ ] Flags documentés / alignés `decisions.md`
- [ ] Tests middleware table-driven (4 coins)
- [ ] `./scripts/check.sh` vert

### Ne pas paralléliser avec

A1*, B2, B3 (templates).

---

## Issue B2 — `[ui] Vocabulaire Listes vs Modèles (ShowSubjectColumn)`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 2  
**Bloqué par** : B1

### Objectif

Aligner nav, formulaires, empty states, CTA : « Listes » si `!ShowSubjectColumn`, « Modèles » sinon.

### Critères d'acceptation

- [ ] Matrice paliers P0–P2 dans la PR
- [ ] Zéro « Modèle » visible en mono-sujet
- [ ] `./scripts/check.sh` vert
- [ ] Definition of Eco cochée (Δ app.css ≤ 0)

---

## Issue B3 — `[ui] Hub org solo minimal vs onglet Organisation`

**Labels** : `area:ui`, `area:admin`, `vague-thematic`  
**Lot** : 2  
**Bloqué par** : B1

### Objectif

Solo : lien Organisation → hub Inviter + Mes sujets. Onglet complet après 2ᵉ email whitelisté. Moment d’unlock explicite.

### Critères d'acceptation

- [ ] Comportement conforme `decisions.md`
- [ ] Micro-copy unlock whitelist→org (B0 peut fusionner si petit)
- [ ] Tests ou scénarios documentés
- [ ] `./scripts/check.sh` vert

---

## Issue B0 — `[ui] Moments d'unlock P0→P1 et P1→P2`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 2  
**Bloqué par** : B1

### Objectif

Flash / bandeau one-shot quand le 2ᵉ membre ou le 2ᵉ sujet apparaît : ce qui se débloque.

### Out

- Tours guidés persistants, checklists onboarding

### Critères d'acceptation

- [ ] Affiché une fois (cookie ou pref user/org)
- [ ] Copy courte, pas de widgets
- [ ] `./scripts/check.sh` vert

---

## Issue B5 — `[docs] Documenter paliers UI P0–P3 dans PLAN.md`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 2 (parallèle)

### Objectif

Section « Paliers UI » dans `docs/PLAN.md` + lien depuis `docs/ROADMAP.md`.

### Critères d'acceptation

- [ ] P0–P3 + flags décrits
- [ ] Pas de code applicatif requis
- [ ] Lien ROADMAP

---

## Issue B6 — `[ui] Wizard : préremplir sujet si SimpleSubjectID`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 2  
**Bloqué par** : B1

### Objectif

Si un seul sujet visible : sauter ou préremplir l’étape sujet du wizard `/revues/nouvelle`.

### Critères d'acceptation

- [ ] Routes inchangées (pas de fork)
- [ ] Matching domaines respecté
- [ ] `./scripts/check.sh` vert

---

## Issue A1a — `[ui] Empty states — sujets`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 2  
**Bloqué par** : B1, B2  
**Note** : scinde #63

### Critères d'acceptation

- [ ] Empty 0 sujet : CTA guidé selon palier
- [ ] Matrice P0/P1 vocabulaire
- [ ] `./scripts/check.sh` vert
- [ ] Definition of Eco

---

## Issue A1b — `[ui] Empty states — listes / modèles`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 2  
**Bloqué par** : B1, B2  
**Note** : scinde #63

### Critères d'acceptation

- [ ] Empty index modèles/listes avec CTA « Créer… »
- [ ] Libellés Listes vs Modèles cohérents
- [ ] `./scripts/check.sh` vert

---

## Issue A1c — `[ui] Empty states — revues`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 2  
**Bloqué par** : B1, B2  
**Note** : scinde #63

### Critères d'acceptation

- [ ] Empty `/revues` : CTA lancer revue + copy Bienvenue adaptée palier
- [ ] Distinct de « 0 résultat filtré »
- [ ] `./scripts/check.sh` vert

---

## Issue C0 — `[attachments] Harden PJ — IDOR, magic bytes, disposition`

**Labels** : `area:attachments`, `area:auth`, `vague-thematic`  
**Lot** : 3 (si gaps)  
**Revue humaine** si toucher auth download

### Objectif

Avant d’embarquer des PJ dans un ZIP preuve : GET IDOR, magic bytes, `Content-Disposition: attachment`, path UUID.

### Critères d'acceptation

- [ ] Tests IDOR cross-org / cross-run
- [ ] Rejet type non autorisé
- [ ] `./scripts/check.sh` vert

---

## Issue C1 — `[core] Export preuve ZIP (CSV + manifest + sha256)`

**Labels** : `area:core`, `vague-thematic`  
**Lot** : 3  
**Note** : WIP `cursor/issue-evidence-zip-f21b`

### Objectif

Téléchargement ZIP pour revue `done` : `revue.csv`, `manifest.json`, `sha256sum.txt`.

### Out

- Inclusion binaires PJ (C3)
- Prévisualisation riche HTML

### Critères d'acceptation

- [ ] RBAC / `ResolveSubjectAccess` identique à la fiche run
- [ ] Hash CSV dans manifest = contenu servi
- [ ] Tests `evidence_test.go`
- [ ] `./scripts/check.sh` vert
- [ ] Definition of Eco (téléchargement, pas preview)

---

## Issue C2 — `[ui] Afficher hash preuve + téléchargement sur revue done`

**Labels** : `area:ui`, `area:core`, `vague-thematic`  
**Lot** : 3  
**Bloqué par** : C1

### Objectif

Sur fiche revue `done` : hash visible + bouton télécharger ZIP preuve. Capability P3 (pas masqué par SimpleUI).

### Critères d'acceptation

- [ ] Visible dès qu’une preuve peut être générée
- [ ] Copy compréhensible sans admin intégrations
- [ ] Matrice P0/P3
- [ ] `./scripts/check.sh` vert

---

## Issue C3 — `[core] Preuve ZIP — hashes PJ (binaires optionnels plafonnés)`

**Labels** : `area:core`, `area:attachments`, `vague-thematic`  
**Lot** : 6+ (créée pour tracker ; implémenter après C2 stable)  
**Bloqué par** : C1, C0

### Objectif

Défaut : chemins + hashes PJ dans le manifeste **sans** binaires. Option explicite « Inclure fichiers » avec plafond + stream disque.

### Critères d'acceptation

- [ ] Pas de path traversal / ZIP-slip
- [ ] Plafond taille/nombre documenté
- [ ] Tests
- [ ] `./scripts/check.sh` vert
- [ ] Revue humaine si PJ embarquées

---

## Issue A3 — `[meta] Pilote vague 1a — checklist PASS/FAIL`

**Labels** : `vague-thematic`  
**Lot** : 4  
**Note** : = #66 — **humain**, pas agent code. Mettre à jour le corps de #66.

### Objectif

Valider le parcours Marie / Thomas / Sophie + critères terrain.

### Critères succès terrain (6–8 semaines / 5–20 pers.)

- [ ] ≥ 80 % des revues planifiées du pilote faites dans Revues
- [ ] Export (CSV ou preuve) utilisé ≥ 1× pour un destinataire réel
- [ ] Chaque NOK a un commentaire ; PJ si le métier l’exige
- [ ] Moins de relances manuelles grâce aux échéances (si D1 livré)
- [ ] Décision go/no-go extension **sans** attendre Jira/Slack
- [ ] Scénario Marie modèle / Thomas exécute / Sophie exporte = PASS

### Out

- Implémentation code par un agent Cloud

---

## Issue D1 — `[notifications] Rappels échéance J-1 + badge en retard`

**Labels** : `area:notifications`, `area:ui`, `vague-thematic`  
**Lot** : 5  
**Bloqué par** : #66 PASS, SMTP (#18)

### Objectif

Email J-1 + badge/filtre « en retard » sur `/revues`.

### Critères d'acceptation

- [ ] Job/cron ou déclenchement documenté dans le même binaire
- [ ] Badge liste + filtre simple
- [ ] Tests
- [ ] `./scripts/check.sh` vert
- [ ] Definition of Eco

---

## Issue D6 — `[ui] Filtres /revues gated par palier`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 5  
**Bloqué par** : B1, #66 PASS

### Objectif

Filtres statut / échéance / domaine / assigné **uniquement** si la colonne ou capability équivalente est visible (pas d’annulation SimpleUI).

### Critères d'acceptation

- [ ] Matrice P0–P2 : quels contrôles
- [ ] Empty filtré ≠ empty onboarding
- [ ] Pagination 25 conservée
- [ ] Max 4 filtres ; 1 fragment HTMX
- [ ] `./scripts/check.sh` vert
- [ ] Definition of Eco

---

## Issue E3prime — `[integrations] Webhooks — retry durable léger`

**Labels** : `area:integrations`, `vague-thematic`  
**Lot** : 5  
**Bloqué par** : #66 PASS, G2  
**Revue humaine** obligatoire

### Objectif

Retry borné via table `webhook_deliveries` + drain au request ou cron 1′ **dans le même binaire** (pas de Redis / worker séparé).

### Critères d'acceptation

- [ ] Re-check anti-SSRF à chaque tentative
- [ ] Backoff + TTL + poison message documentés
- [ ] Tests
- [ ] `./scripts/check.sh` vert

### Out

- Queue externe, Slack natif

---

## Issue E6 — `[integrations] Notion — erreurs UX import/export`

**Labels** : `area:integrations`, `area:ui`, `vague-thematic`  
**Lot** : 5  
**Bloqué par** : #66 PASS  
**Revue humaine**

### Objectif

Messages d’erreur actionnables sur import modèle / export revue (pas de bi-sync).

### Critères d'acceptation

- [ ] Erreurs utilisateur compréhensibles
- [ ] Pas d’appel Notion au premier paint hors action user
- [ ] `./scripts/check.sh` vert

---

## Issue B4a — `[ui] Matrice capability P3 (doc + PageData)`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 5  
**Bloqué par** : B1, C2 (pour gate preuve)

### Objectif

Documenter et exposer les flags capability (intégration configurée / preuve scellée).

### Critères d'acceptation

- [ ] Flags dans PageData
- [ ] Doc matrice
- [ ] Tests
- [ ] `./scripts/check.sh` vert

---

## Issue B4b — `[ui] Gates UI intégrations (Notion/Jira/webhooks)`

**Labels** : `area:ui`, `area:integrations`, `vague-thematic`  
**Lot** : 5  
**Bloqué par** : B4a

### Objectif

Surfaces admin/métier intégrations visibles si config — jamais masquées par SimpleUI seul.

### Critères d'acceptation

- [ ] Matrice P0 vs P3
- [ ] `./scripts/check.sh` vert

---

## Issue B4c — `[ui] Gate UI preuve scellée`

**Labels** : `area:ui`, `vague-thematic`  
**Lot** : 5  
**Bloqué par** : B4a, C2

### Objectif

Chrome preuve (hash/ZIP) capability-gated, visible en SimpleUI si preuve existe.

### Critères d'acceptation

- [ ] Cohérent avec C2
- [ ] `./scripts/check.sh` vert

---

## Issue D7 — `[admin] Preset ui_subject_label org`

**Labels** : `area:admin`, `area:ui`, `vague-thematic`  
**Lot** : opportuniste (créée avec lot 5 si absente)

### Objectif

Preset admin `sujet|cible|entite|asset` + injection `Labels.Subject.*`.

### Critères d'acceptation

- [ ] Écran admin + défaut `sujet`
- [ ] `./scripts/check.sh` vert

---

## Prompt agent type

```
Repo jeb-maker/revues. Implémente UNIQUEMENT l'issue #N.
Lis AGENTS.md, docs/CONVENTIONS.md, docs/GO.md, docs/DEFINITION_OF_DONE.md.
Spec : docs/issues/thematic-roadmap-epic.md.
Branche : cursor/issue-N-thematic-<slug>-f21b
PR titre = titre issue, corps : Closes #N
./scripts/check.sh doit passer avant push.
Definition of Eco cochée si UI/intégration.
```
