# Tests de charge Revues (k6)

Cible : l’API Go (`/api/v1`), pas le front Svelte.
Objectif : mesurer latences / erreurs sous charge, surtout la contention **SQLite WAL**
et le **lancement de revue** (snapshot transactionnel).

## Prérequis

1. API locale avec `REVUES_DEV_AUTH=0` (recommandé) :

```bash
set -a && source .env && set +a
# s'assurer que REVUES_DEV_AUTH n'est pas à 1
go run ./cmd/revues
```

2. k6 sur le `PATH`, ou binaire dans `.tools/k6` (gitignored) :

```bash
mkdir -p .tools
curl -fsSL https://github.com/grafana/k6/releases/download/v0.57.0/k6-v0.57.0-linux-amd64.tar.gz \
  | tar -xz -C .tools --strip-components=1 k6-v0.57.0-linux-amd64/k6
```

## Lancer

```bash
./scripts/load/run.sh read    # lecture seule
./scripts/load/run.sh mix     # lecture + PATCH + lancement revue
./scripts/load/run.sh runs    # lancement de revues uniquement (snapshot)
```

Variables utiles :

| Variable | Défaut | Rôle |
|----------|--------|------|
| `BASE_URL` | `http://127.0.0.1:8080` | API |
| `VUS` | `20` | utilisateurs virtuels |
| `DURATION` | `30s` | plateau (hors ramp) |
| `WRITE_RATIO` | `0.2` | part PATCH/POST subject (`mix`) |
| `RUN_RATIO` | `0.1` | part lancement revue (`mix`) |
| `SESSION_POOL` | `10` | comptes pré-auth en setup (anti rate-limit) |
| `LOAD_EMAIL_PREFIX` | `loadtest-vu` | emails `loadtest-vu{N}@example.com` |
| `LOAD_PASSWORD` | `loadtest-pass-12` | mot de passe des comptes load |

Exemples :

```bash
VUS=50 DURATION=1m ./scripts/load/run.sh mix
VUS=20 DURATION=1m ./scripts/load/run.sh runs
```

## Auth

Les sessions sont créées **en setup** (séquentiel, ~2 s entre chaque) pour ne pas
saturer le rate-limit `/auth/login|register` (30/min). Chaque compte reçoit une org
+ un sujet/modèle/revue de seed si besoin.

**Recommandé : `REVUES_DEV_AUTH=0`**. DevAuth partage un admin et rotate les sessions → CSRF 403.

## Scénarios

| Fichier | Contenu |
|---------|---------|
| `read.js` | `/me`, sujets, modèles, revues + détail |
| `mix.js` | ~70 % read / ~20 % PATCH item / ~10 % `POST …/runs` |
| `runs.js` | uniquement lancement de revue + détail post-création |

Seuils par défaut : échecs HTTP `< 2 %`, checks `> 95 %`, `p95` selon scénario
(800 ms read, 1.5 s mix, 2 s runs).

## Interprétation

1. `read` → plafond lectures.
2. `mix` → contention réaliste.
3. `runs` → goulot snapshot SQLite (écritures lourdes).

Comparer avec : `go test ./internal/store/ -bench=BenchmarkConcurrentListSubjects -benchmem`.

Ne pas lancer contre une instance de production.
