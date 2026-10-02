.PHONY: api frontend-api sqlc

api:
	./scripts/generate-api.sh

# Client TypeScript (openapi-typescript) → frontend/src/lib/api/schema.d.ts
frontend-api:
	./scripts/generate-frontend-api.sh

# Régénère le package Go typé (internal/store/sqlc) depuis internal/store/queries.
sqlc:
	./scripts/generate-sqlc.sh

