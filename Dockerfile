# Revues — binaire Go (CGO off, SQLite modernc) + SPA SvelteKit construite dans l'image.
# Bind localhost uniquement via compose ; Caddy hôte reverse_proxy.

# --- Front : build SvelteKit (adapter-static) -> frontend/build ---------------
FROM node:22-bookworm-slim AS front
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build && test -f build/index.html

# --- API : binaire Go statique --------------------------------------------------
FROM golang:1.22-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/revues ./cmd/revues

# --- Image finale ------------------------------------------------------------------
FROM debian:bookworm-slim
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates curl \
	&& rm -rf /var/lib/apt/lists/* \
	&& groupadd --gid 10001 revues \
	&& useradd --uid 10001 --gid revues --home-dir /app --shell /usr/sbin/nologin revues \
	&& mkdir -p /app/data/attachments \
	&& chown -R revues:revues /app

WORKDIR /app
COPY --from=build /out/revues /app/revues
# Le binaire sert ce dossier (REVUES_SPA_DIR) ; sans lui il répond 503 « SPA non construite ».
COPY --from=front --chown=revues:revues /src/frontend/build /app/frontend/build
ENV REVUES_SPA_DIR=/app/frontend/build

USER revues:revues
EXPOSE 8080
VOLUME ["/app/data"]

# API vivante ET SPA servie (le stub répond 503 → conteneur unhealthy).
HEALTHCHECK --interval=15s --timeout=5s --retries=5 --start-period=10s \
	CMD curl -sf http://127.0.0.1:8080/healthz >/dev/null \
	&& curl -sf -o /dev/null http://127.0.0.1:8080/login || exit 1

ENTRYPOINT ["/app/revues"]
