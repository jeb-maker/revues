# OpenAPI

Source de vérité du contrat HTTP `/api/v1/**`.

| Fichier | Rôle |
|---------|------|
| `openapi.yaml` | Spec OpenAPI 3 |
| `oapi-codegen.yaml` | Config codegen serveur (chi) |

## Génération

```bash
make api
# équivalent : ./scripts/generate-api.sh
```

- Outil : **oapi-codegen** `v2.4.1` (`chi-server` + `models`)
- Sortie commitée : `internal/api/v1/oapi.gen.go`
- Handlers métier (non générés) : `internal/api/v1/handler.go` (`Server`)

Détails : [docs/API.md](../../docs/API.md).
