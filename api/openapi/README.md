# OpenAPI

Source de vérité du contrat HTTP `/api/v1/**`.

| Fichier | Rôle |
|---------|------|
| `openapi.yaml` | Spec OpenAPI 3 |
| `oapi-codegen.yaml` | Config codegen serveur (chi) |

## Génération

```bash
make api            # serveur Go
make frontend-api   # client TypeScript
```

### Serveur (oapi-codegen)

- Outil : **oapi-codegen** `v2.4.1` (`chi-server` + `models`)
- Script : `./scripts/generate-api.sh`
- Sortie commitée : `internal/api/v1/oapi.gen.go`
- Handlers métier (non générés) : `internal/api/v1/handler.go` (`Server`)

### Client front (openapi-typescript)

- Outil : **openapi-typescript** + wrappers `openapi-fetch`
- Script : `./scripts/generate-frontend-api.sh`
- Sortie commitée : `frontend/src/lib/api/schema.d.ts`
- Wrappers : `frontend/src/lib/api/client.ts`

Détails : [docs/API.md](../../docs/API.md) · [docs/FRONTEND.md](../../docs/FRONTEND.md).
