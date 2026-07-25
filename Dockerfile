# Revues — binaire Go (CGO off, SQLite modernc) + assets embarqués.
# Bind localhost uniquement via compose ; Caddy hôte reverse_proxy.

FROM golang:1.22-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/revues ./cmd/revues

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

USER revues:revues
EXPOSE 8080
VOLUME ["/app/data"]

HEALTHCHECK --interval=15s --timeout=5s --retries=5 --start-period=10s \
	CMD curl -sf http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["/app/revues"]
