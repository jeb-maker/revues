# Audit live UX/UI — persona **claire** (reader)

**Date** : 2026-07-21  
**Environnement** : `http://127.0.0.1:8080`, `REVUES_DEV_AUTH=1`, seed présent (non reset)  
**Auth** : `POST /auth/dev/login` `user_id=4` → session claire · badge **Lecteur**  
**Périmètre demandé** : **C1 nav** · **C2 run write CTAs** · **C3 launch** + deep-links + sense of agency  
**Méthode** : probes HTTP live (GET/POST) + lecture templates ; Confirmé / Partiel / Rejeté ; **read-only**.

---

## Synthèse

Live, claire est **bien plafonnée en écriture** : nav sans Modèles/Organisation, aucun CTA Lancer / cocher / clôturer, POST item·complete·launch·assign → **404**, wizard `GET /revues/nouvelle` → **303 `/revues`**. Le trou principal reste le **sense of agency** : UI « shell éditeur » sans signal lecture seule, conflit **Lecteur** vs **Contributeur** sur fiche sujet, et **fuite catalogue** `GET /modeles` 200 (nav masquée, page orpheline).

---

## C1 — Navigation

| Attendu (matrice reader) | Live claire | Statut |
|--------------------------|-------------|--------|
| Nav Revues | Présente (`/revues`) | **OK** |
| Nav Mes tâches (`ShowMyTasks`) | Présente ; `aria-current` sur `/mes-taches` | **OK** |
| Nav Modèles | Absente (`ne .User.Role "reader"`) | **OK** |
| Nav Organisation | Absente | **OK** |
| Badge rôle header | `Lecteur` | **OK** |

**Preuve** : `/revues` et `/mes-taches` — `site-nav` = « Revues · Mes tâches » uniquement ; `/admin` → **403**.

**Écart lié** : deep `GET /modeles` → **200** alors qu’aucun onglet n’est actif (`aria-current` vide) — page catalogue **orpheline** (Fuite #5).

---

## C2 — Run write CTAs (grille / clôture / item)

Surface testée : `GET /runs/101` (`in_progress`), `GET /runs/100` (`done`), `GET /runs/101/items/444`.

| Contrôle | Attendu reader | Live | Statut |
|----------|----------------|------|--------|
| Selects statut / commentaire HTMX | Masqué | Absents (`hx-post` = 0 ; badges `OK` / `En attente`) | **OK** |
| Colonne Assigné | Lecture (I2) | Texte `—` ; pas de select assign | **OK** |
| Colonne Actions / cocher | Masqué | Absente | **OK** |
| Section Clôturer | Masqué | Absente (`CanComplete`) | **OK** |
| Upload PJ / forms Jira | Masqué | Absents ; item montre « Aucune issue » / « Aucune pièce jointe » | **OK** (bruit P3) |
| Export CSV (done) | Faire | Lien `Exporter CSV` → `/runs/100/export.csv` | **OK** |
| Filtres GET section/statut | Lecture OK | Form `GET /runs/{id}` (pas mutation) | **OK** — ne pas confondre avec écriture |

**POST live** (CSRF session claire) :

| Requête | Code |
|---------|------|
| `POST /runs/101/items/444` (status=ok) | **404** |
| `POST /runs/101/complete` | **404** |
| `POST /runs/101/items/444/assign` | **404** |

**Sense of agency (C2)** : la grille ressemble à l’UI éditeur sans contrôles — **aucun** libellé « lecture seule » / `role="status"` sur `/runs/101`. Claire doit déduire l’impossibilité d’agir. Badge header « Lecteur » seul = signal faible.

---

## C3 — Launch

| Surface | Attendu | Live | Statut |
|---------|---------|------|--------|
| Toolbar `/revues` « Lancer une revue » | Masqué | `COUNT[Lancer]=0` | **OK** |
| Empty-state Lancer | Masqué (sinon message éditeur) | N/A seed (liste pleine) ; template `Demandez à un éditeur…` si `!CanLaunch` | **OK** (code) |
| Fiche sujet CTA Lancer | Masqué | Pas de `.button` / `for_run` | **OK** |
| Fiche modèle « Lancer avec ce modèle » | Masqué | `/modeles/9` : Lancer=0 | **OK** |
| « Lancer une autre revue » (run done) | Masqué (`CanLaunch`) | Absent sur `/runs/100` | **OK** |
| Deep `GET /revues/nouvelle` | Deny | **303** `Location: /revues` | **Partiel** vs matrice « 404 » |
| `POST /subjects/90/revues` | 404 | **404** | **OK** |
| `POST /revues/nouvelle` | 404 | **404** | **OK** |
| `GET …/modeles?for_run=1` | 404 | **404** | **OK** |

---

## Deep-links & surfaces annexes

| URL | Code | Note UX |
|-----|------|---------|
| `/revues`, `?status=in_progress\|done` | 200 | Hub OK ; pas de CTA write |
| `/runs/{id}`, `/runs/{id}/items/{id}` | 200 | Lecture ; shell éditeur |
| `/mes-taches` | 200 | Empty « Aucune tâche » → « Voir les revues » |
| `/subjects/90` | 200 | Pas Lancer/Modifier ; **Contributeur** vs badge Lecteur |
| `/subjects/90/modeles` | 200 | Copy « Gérez… onglet Modèles » + lien `/modeles` — **trompeur** |
| `/subjects/90/modeles?for_run=1` | 404 | Deny dur OK |
| `/revues/nouvelle` | 303 → `/revues` | Soft deny |
| `/modeles`, `/modeles/9` | 200 | **Fuite #5** ; CRUD/Lancer absents |
| `/admin` | 403 | Aligné |

---

## Constats (passes + confirmation)

| # | Passe | Constat | Statut | Preuve live / code |
|---|-------|---------|--------|--------------------|
| 1 | C1 / RBAC | Nav reader correcte (Revues · Mes tâches) | **Confirmé OK** | HTML nav ; `site_nav.html` L.13–16 |
| 2 | C1 / RBAC | `GET /modeles` 200 sans onglet (orphelin, pas d’`aria-current`) | **Confirmé** Fuite #5 | Live 200 ; title « Modèles — Revues » |
| 3 | C2 | Write CTAs run absents ; POST item/complete/assign 404 | **Confirmé OK** | Live probes |
| 4 | C3 | CTAs Lancer + « autre revue » absents ; POST launch 404 | **Confirmé OK** | Live ; `run_show.html` L.206 |
| 5 | C3 | Wizard GET → 303 `/revues` (≠ 404 matrice) | **Partiel** | Live `Location: /revues` |
| 6 | Parcours | Copy « onglet Modèles » sur picker sujet lecture | **Confirmé** | `/subjects/90/modeles` muted + lien |
| 7 | Présence | `rôles effectif : Contributeur` vs badge **Lecteur** | **Confirmé** | `/subjects/90` « Vos accès » |
| 8 | Présence / agency | Pas de signal lecture seule sur run `in_progress` | **Confirmé** | `/runs/101` |
| 9 | Parcours | Blocs Équipes/Membres vides **avant** Revues | **Confirmé** | h2 order sujet |
| 10 | Décisions | Masquer Modèles reader = intentionnel | **Confirmé** | `decisions.md` |
| 11 | Export | CSV done accessible reader | **Confirmé OK** | `/runs/100` |

---

## Critique adverse

### Ce qui tient

- **Plafond write réel** (UI + serveur) : C2/C3 ne sont pas du security-through-obscurity.
- **Filtres** section/statut = GET lecture — faux positif si on les compte comme « selects d’écriture ».
- **Fuite #5** = lecture catalogue, pas élévation : Créer / Lancer / edit restent absents / 404.
- Soft redirect wizard : moins brutal qu’un 404 « page cassée » pour un bookmark.

### Downgrades / hors scope

- Exiger le parcours alice (cocher / lancer) pour claire → **faux positif**.
- Traiter absence de CTA Lancer comme bug présence → **non** : contrat reader.
- Fuites SimpleUI (#1–4) → **hors persona claire**.

### Pièges

- Seed multi-revues : empty-state reader « Demandez à un éditeur » non exercé en live (template seulement).
- DevAuth switcher (`<select>` header) n’est pas un contrôle métier.

---

## Top issues actionnables

| Rang | Issue | Type | Effort | Priorité |
|------|-------|------|--------|----------|
| 1 | **Fuite #5** : deny `GET /modeles` (+ show) pour `reader` **ou** assumer lecture + entrée secondaire cohérente | UI↔RBAC | S–M | **P1** |
| 2 | Copy picker sujet : ne pas renvoyer vers « onglet Modèles » si `User.Role==reader` | Copy | S | **P1** |
| 3 | Fiche sujet : ne pas afficher « Contributeur » sous plafond global `reader` | Présence | S | **P1** |
| 4 | Signal lecture seule sur `/runs/{id}` quand `!CanCheck && !CanComplete` | Agency | S | **P2** |
| 5 | Ordre fiche sujet : Revues avant Équipes/Membres vides | Parcours | S | **P2** |
| 6 | Doc/UX deny wizard GET : 303 vs 404 matrice | Cohérence | XS | **P3** |
| 7 | Replier section Jira vide si `!CanLinkJira && !link` | Présence | XS | **P3** |

---

## Décisions produit ouvertes

1. Reader × catalogue `/modeles` : fermer serveur **ou** assumer deep-link (+ corriger copy) ?  
2. Affichage rôle effectif sous plafond `reader` : grant, plafond, ou les deux ?  
3. Bandeau « lecture seule » sur revue : oui / non ?

---

## Plan PR suggéré (ne pas implémenter ici)

1. **PR A — Reader catalogue** : décision 1 + copy `checklist_templates_list` + test live/reader.  
2. **PR B — Présence lecteur** : DisplayRole sous plafond reader ; optionnel bandeau run + reorder collab.

---

## Annexe — probes live (extrait)

```
POST /auth/dev/login user_id=4 → 303 Set-Cookie revues_session
GET  /revues                     → 200  (~25 Ko)  nav: Revues · Mes tâches · Lecteur
GET  /revues/nouvelle            → 303 Location: /revues
GET  /modeles                    → 200
GET  /admin                      → 403
GET  /subjects/90/modeles?for_run=1 → 404
POST /runs/101/items/444         → 404
POST /runs/101/complete          → 404
POST /subjects/90/revues         → 404
« Lancer une autre revue » sur /runs/100 → absent
```

Fin audit live **claire** — aucune correction, pas de reset DB / kill server / commit.
