# Déploiement (Docker + Caddy hôte)

Cible : VPS multi-apps. Caddy sur l’hôte (`:80`/`:443`) reverse-proxy vers le conteneur bindé en `127.0.0.1`.

## Fichiers

| Chemin | Rôle |
|--------|------|
| [`Dockerfile`](../Dockerfile) | Build multi-stage du binaire Go (CGO off) |
| [`docker-compose.yml`](../docker-compose.yml) | Service `app` + volume SQLite/attachments |
| [`generate-env.sh`](generate-env.sh) | Génère `.env` (secrets) sur l’hôte |
| [`caddy/revues.betafly.ovh.caddy`](caddy/revues.betafly.ovh.caddy) | Snippet Caddy prod |

## Déployer

```bash
# Sur l'hôte, depuis la racine du dépôt (ex. /opt/revues)
bash deploy/generate-env.sh
# Éditer .env : REVUES_GITHUB_CLIENT_* + REVUES_BOOTSTRAP_ADMIN_EMAIL

# Log Caddy (user caddy)
touch /var/log/caddy/revues.log
chown caddy:caddy /var/log/caddy/revues.log
chmod 640 /var/log/caddy/revues.log

install -m 644 deploy/caddy/revues.betafly.ovh.caddy /etc/caddy/conf.d/
caddy validate --config /etc/caddy/Caddyfile
caddy reload --config /etc/caddy/Caddyfile --force

docker compose up -d --build
curl -sf http://127.0.0.1:8088/healthz   # → ok
```

DNS requis : `A revues.betafly.ovh` → IP publique du VPS. Callback OAuth GitHub : `https://revues.betafly.ovh/auth/github/callback`.

## Ports

| Service | Bind hôte | Conteneur |
|---------|-----------|-----------|
| revues  | `127.0.0.1:8088` | `:8080` |

Ne pas exposer `8088` publiquement : seul Caddy doit proxyfier.
