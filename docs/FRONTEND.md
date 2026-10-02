# Front — SvelteKit + miniature-broccoli

Complète [ADR-001](./ADR-001-api-first-svelte.md).

## Stack

| Élément | Choix |
|---------|--------|
| Framework | SvelteKit (mode SPA / `adapter-static`) |
| UI | `@jeb-maker/mb` Web Components (Lit), tokens CSS |
| Données | Client **TypeScript généré** depuis OpenAPI |
| Build | Vite (via SvelteKit) |

## Emplacement

```
frontend/                 # app SvelteKit
  src/routes/             # pages
  src/lib/api/            # client généré + wrappers fins
  src/lib/components/     # composition mb + logique écran
static vendor mb          # bundlé ou package ; pas réinventer les CE
```

Go sert le build `frontend/build` (ou équivalent) en production ; en dev : proxy Vite → API `:8080`.

## Règles

1. **Pas de `html/template` Go** ni HTMX pour l’UI métier.
2. Composants visuels : préférer **mb-*** (`mb-button`, `mb-table`, `mb-input`, …) avant HTML custom.
3. Aucune règle RBAC « de confiance » uniquement côté client — le serveur tranche ; le front masque seulement.
4. CSRF : lire le token bootstrap, l’envoyer en `X-CSRF-Token` sur chaque mutation.
5. Pas de polling ni WebSocket (inchangé).
6. i18n hors scope sauf issue dédiée.

## Budgets (SPA)

Mesurés par `check.sh` (à brancher dans le WP CI) :

| Métrique | Cible indicative |
|----------|------------------|
| JS app (hors `vendor/` mb + reports) | à définir dans le WP CI ; viser sobriété |
| CSS app (hors tokens mb) | sobriété ; tokens mb = vendor |
| Vendor mb | mesuré, hors fail strict initial si documenté |

## Tests front

- Minimum : build `npm run check` / `npm run build` vert en CI.
- Tests e2e hors scope v1 rewrite sauf issue dédiée.
- Parité fonctionnelle validée manuellement ou via tests API pour la logique.

## Référence mb

Voir `web/static/vendor/jeb-maker-mb/README.md` (ou emplacement post-migration) et upstream [miniature-broccoli](https://github.com/jeb-maker/miniature-broccoli).
