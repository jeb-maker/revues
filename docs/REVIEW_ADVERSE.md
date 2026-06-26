# Revue adverse — synthèse (juin 2026)

Revue critique du plan, de la roadmap et des issues **avant** tout code métier.
Quatre angles : sécurité, architecture, produit, harness agents.

**Verdict global** : vision solide, mais le plan mélange MVP et produit mature. Sans harness et corrections ci-dessous, les agents produiront des migrations correctives, des failles RBAC et du scope creep.

---

## Bloquants — à traiter avant issue #5

| # | Problème | Action |
|---|----------|--------|
| B1 | Pas de harness (CI, conventions, DoD) | ✅ PR harness (ce commit) |
| B2 | Schéma incomplet (`sessions`, `section`, `due_date`) | ✅ `docs/schema/canonical.sql` |
| B3 | RBAC global × projet non défini → IDOR | ✅ `docs/RBAC.md` + tests exigés en DoD |
| B4 | Numérotation ROADMAP ≠ issues GitHub | ✅ ROADMAP aligné sur #5–#29 |
| B5 | Webhooks = risque SSRF sous-estimé | Spec ajoutée dans PLAN + CONVENTIONS |
| B6 | Pas d'export → ne remplace pas Excel | Issue #31 (export CSV vague 1a) |

---

## Sécurité — points critiques

### MUST FIX (vague 1, avant intégrations)

1. **Matrice RBAC** testée : chaque route × chaque rôle → 401/403/200. Voir [RBAC.md](./RBAC.md).
2. **IDOR** : contrôle objet-par-objet (projet, revue, point, pièce jointe) — le middleware global ne suffit pas.
3. **Sessions** : rotation post-OAuth, révocation en base, TTL inactivité + absolu.
4. **Whitelist** : email GitHub **vérifié** (`email_verified=true`) avant activation.
5. **CSRF** : tous POST/PUT/PATCH/DELETE, y compris `hx-post` HTMX (`hx-headers`).
6. **Secrets** : AES-256-GCM, clé `REVUES_ENCRYPTION_KEY` en env uniquement, jamais en DB.
7. **Webhooks** (vague 2) : anti-SSRF (blocklist IP privées, timeout 5s, max 1 redirect), `event_id` unique, HMAC-SHA256 documenté.
8. **Uploads** (vague 3) : magic bytes, noms UUID, `Content-Disposition: attachment`, auth sur chaque GET.

### CAN DEFER

CSP stricte, scan antivirus, rotation clés, rate limiting global, audit admin complet, queue emails persistante, OAuth Jira Server.

### Tests sécurité minimum (exigés par harness)

```
TestRBAC_Matrix, TestIDOR_CrossProject, TestCSRF_MissingToken,
TestSession_Fixation, TestUnauthenticated, TestWebhook_HMAC (v2),
TestWebhook_SSRF_Block (v2), TestUpload_Rejects (v3)
```

---

## Architecture — points critiques

### Sur-ingénierie précoce

- Double RBAC (6 rôles) dès v1 — acceptable si matrice figée et testée.
- Jira Cloud **et** Server/DC en vague 2 — **recadrer** : Cloud d'abord, Server si demande avérée.
- Versionnement strict des modèles en vague 1 — **assouplir** : versionner au premier snapshot ; édition libre tant qu'aucune revue n'existe.

### Sous-ingénierie

- Pas de `sessions` dans schéma initial → corrigé dans `canonical.sql`.
- Sections modèles : colonne `section` sur items (KISS) plutôt que table séparée en v1.
- Pas de CI → corrigé par harness.
- Goroutines email/webhook sans file → acceptable v1, documenter perte au restart.
- Concurrence HTMX : ajouter `updated_at` sur `run_items` pour détecter écrasements (v1.1).

### Issues mal découpées

| Issue | Problème | Recommandation |
|-------|----------|----------------|
| #7 Auth | Trop large (OAuth+session+CSRF) | Agent unique, revue humaine obligatoire |
| #11 Modèles | Versionnement + sections + UI | Ne pas ajouter WYSIWYG ; formulaires structurés |
| #13 Audit | Séparé de #12 statuts | Implémenter audit **dans** #12 |
| #19–#23 Jira | Cloud+Server couplé | Scinder : Cloud v2a, Server v2b |

---

## Produit — points critiques

### Contradiction simple / riche

Le plan affiche trois piliers égaux mais la vague 1 est déjà riche. **Recadrage proposé** :

| Vague 1a (MVP) | Vague 1b (enrichissement) |
|----------------|---------------------------|
| Auth + whitelist | Audit trail complet |
| Projets + 2 rôles locaux (`contributor`/`viewer`) | Assignation par point |
| Modèles (édition libre → version au snapshot) | Versionnement strict UI |
| Lancer / cocher / clôturer | HTMX avancé |
| Commentaire obligatoire si `nok` | |
| **Export CSV** | |
| **due_date** sur revue | |

### Manques critiques adoption

- **Export CSV** — sans export, Excel reste le fallback.
- **due_date** — mentionné pour emails J-1 mais absent du schéma → ajouté.
- **Backup SQLite** — issue infra à planifier.
- **Onboarding** — premier admin, états vides : spec UX manquante.
- **Demande d'accès** — OAuth OK puis whitelist refusée : prévoir message + contact admin.

### YAGNI — reporter

- Import Notion (garder export seul en vague 3)
- Jira Server/DC (après Cloud)
- Webhooks avant validation SMTP en pilote
- Pièces jointes avant export CSV

---

## Harness agents — décision

Un **harness** est indispensable avant délégation :

| Fichier | Rôle |
|---------|------|
| [AGENTS.md](../AGENTS.md) | Contrat agents Cloud |
| [CONVENTIONS.md](./CONVENTIONS.md) | Code, SQL, routes |
| [DEFINITION_OF_DONE.md](./DEFINITION_OF_DONE.md) | Critères merge |
| [RBAC.md](./RBAC.md) | Matrice permissions |
| [schema/canonical.sql](./schema/canonical.sql) | Schéma normatif |
| `scripts/check.sh` | Gate local + CI |
| `.github/workflows/ci.yml` | Pipeline |
| PR template + checklist | Revue humaine |

**Règle d'or** : 1 issue = 1 PR = `./scripts/check.sh` vert = checklist cochée.

---

## Actions issues GitHub recommandées

À créer après merge harness :

| Issue | Titre |
|-------|-------|
| #31 | `[Vague 1a] Export CSV revue clôturée` |
| #32 | `[Vague 1a] Échéance revue (due_date)` |
| #33 | `[infra] Backup SQLite + doc restauration` |
| #34 | `[Vague 1a] Tests RBAC transversaux` |

Modifier les corps d'issues existantes pour référencer `docs/RBAC.md` et `canonical.sql`.

---

## Critères de succès produit (ajout)

| KPI | Cible v1 |
|-----|----------|
| Temps modèle → revue clôturée | < 30 min (équipe pilote) |
| Revues sans retour Excel | > 80 % après 1 mois |
| Routes POST avec RBAC testé | 100 % |
| p95 page détail revue | < 500 ms |
| `check.sh` vert sur main | toujours |

---

## Conclusion

**Ne pas démarrer #5 (bootstrap)** avant merge de cette PR harness.

Ordre recommandé :
1. Merge harness + corrections plan
2. Créer issues #30–#33
3. Déléguer #5 (bootstrap) avec `AGENTS.md` en tête de prompt
4. Revue humaine obligatoire sur #7 (auth) et toute issue `area:integrations`
