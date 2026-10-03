# Déploiement (Docker + Caddy hôte)

Cible : VPS multi-apps. Caddy sur l’hôte (`:80`/`:443`) reverse-proxy vers le conteneur bindé en `127.0.0.1`.

## Fichiers

| Chemin | Rôle |
|--------|------|
| [`Dockerfile`](../Dockerfile) | Build multi-stage : SPA SvelteKit (`node:22`, `npm run build`) + binaire Go (CGO off) ; l'image sert `frontend/build` via `REVUES_SPA_DIR` |
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
curl -sf http://127.0.0.1:8088/login | grep -q '_app/' && echo SPA OK   # 503 = build front absent de l'image
```

DNS requis : `A revues.betafly.ovh` → IP publique du VPS. Callback OAuth GitHub : `https://revues.betafly.ovh/auth/github/callback`.

## Ports

| Service | Bind hôte | Conteneur |
|---------|-----------|-----------|
| revues  | `127.0.0.1:8088` | `:8080` |

Ne pas exposer `8088` publiquement : seul Caddy doit proxyfier.

## SPA absente = 503

Le binaire ne contient pas le front : il sert `REVUES_SPA_DIR` (défaut `frontend/build`). Si le dossier manque, **toutes** les routes UI répondent `503` avec une page « SPA SvelteKit non construite » (l'API `/api/v1` et `/healthz` restent en 200). Le `HEALTHCHECK` Docker et `deploy/update.sh` vérifient `/login` pour qu'un tel déploiement soit visible (`unhealthy` / échec du script) au lieu de passer pour sain.
