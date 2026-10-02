.PHONY: api sqlc

api:
	./scripts/generate-api.sh

# Régénère le package Go typé (internal/store/sqlc) depuis internal/store/queries.
sqlc:
	./scripts/generate-sqlc.sh

