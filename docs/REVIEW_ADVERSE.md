# Revue adverse — pièges connus

Checklist vivante pour agents et relecteurs. Ne pas élargir le scope d'une issue pour « corriger » un point listé ici sans issue dédiée.

---

## Sécurité — MUST FIX

1. **Matrice RBAC** testée : chaque route × chaque rôle → 401/403/200. Voir [RBAC.md](./RBAC.md).
2. **IDOR** : contrôle objet-par-objet (sujet, revue, point, pièce jointe) — le middleware global ne suffit pas.
3. **Sessions** : rotation post-OAuth, révocation en base, TTL inactivité + absolu.
4. **Whitelist** : email GitHub **vérifié** (`email_verified=true`) avant activation.
5. **CSRF** : tous POST/PUT/PATCH/DELETE, y compris `hx-post` HTMX (`hx-headers`).
6. **Secrets** : AES-256-GCM, clé `REVUES_ENCRYPTION_KEY` en env uniquement, jamais en DB.
7. **Webhooks** : anti-SSRF (blocklist IP privées, timeout 5s, max 1 redirect), `event_id` unique, HMAC-SHA256 documenté.
8. **Uploads** : magic bytes, noms UUID, `Content-Disposition: attachment`, auth sur chaque GET.

### CAN DEFER

CSP stricte, scan antivirus, rotation clés, rate limiting global, audit admin complet, queue emails persistante, OAuth Jira Server.

### Tests sécurité minimum

```
TestRBAC_Matrix, TestIDOR_CrossProject, TestCSRF_MissingToken,
TestSession_Fixation, TestUnauthenticated, TestWebhook_HMAC,
TestWebhook_SSRF_Block, TestUpload_Rejects
```

---

## Architecture — pièges

- Ne pas fusionner les feature stores ni remettre le SQL dans les handlers (voir [CONVENTIONS.md](./CONVENTIONS.md)).
- Jira Cloud d'abord ; Server/DC seulement si demande avérée.
- Versionnement modèles : versionner au premier snapshot ; édition libre tant qu'aucune revue n'existe.
- Goroutines email/webhook sans file : acceptable v1, documenter perte au restart.
- Concurrence HTMX : `updated_at` sur `run_items` pour détecter écrasements.

---

## Produit — pièges

- Ne pas confondre marque « Revues » et libellés preset (`Labels.Run` / `Labels.Subject`).
- SimpleUI / disclosure progressive : deep links peuvent rester ouverts alors que la nav est masquée — ce n'est pas forcément un bug RBAC.
- Deny métier → **404** (trade-off IDOR acté), pas 403 révélateur.
- Pas de SPA, pas de polling, pas de WebSocket.
- Budgets éco : HTML ≤ 50 Ko/page ; CSS core ≤ 24 Ko / 8 Ko gzip ; CSS total ≤ 40 Ko / 12 Ko gzip ; JS ≤ 15 Ko ; ≤ 8 requêtes/page.

---

## Harness agents

- **1 issue = 1 PR** — jamais de fourre-tout ni « tant qu'on y est ».
- Critère ambigu → commenter sur l'issue, ne pas deviner.
- Revue humaine obligatoire : auth OAuth, `area:integrations`, chiffrement, webhooks.
- Fichiers sacrés (issue dédiée) : `docs/schema/canonical.sql`, `docs/RBAC.md`, `AGENTS.md`, `.github/workflows/ci.yml`.
