# Onboarding — Revues

Guide en 5 étapes pour une première installation.

## 1. Configurer l'environnement

Copier [.env.example](../.env.example) et renseigner au minimum :

- `REVUES_SESSION_SECRET` — secret aléatoire (32+ octets)
- `REVUES_GITHUB_CLIENT_ID` et `REVUES_GITHUB_CLIENT_SECRET` — application OAuth GitHub
- `REVUES_BOOTSTRAP_ADMIN_EMAIL` — email GitHub **vérifié** du premier administrateur (instances migrées)

Exporter les variables avant de lancer l'application (le binaire ne charge pas `.env` automatiquement).

## 2. Démarrer l'application

```bash
go run ./cmd/revues
```

Ouvrir l’UI (SPA SvelteKit servie par Go, ou `npm run dev` en dev — [FRONTEND.md](./FRONTEND.md)).

Stack cible : [ADR-001](./ADR-001-api-first-svelte.md). Rewrite agents : [rewrite/README.md](./rewrite/README.md).

## 3. Se connecter

Deux voies d'auth (complémentaires) :

- **GitHub OAuth** — « Se connecter avec GitHub » (email GitHub **vérifié** requis)
- **Email + mot de passe** — inscription puis login

**Instance migrée (organisation `default` existante)** : se connecter avec le compte correspondant à `REVUES_BOOTSTRAP_ADMIN_EMAIL`. Au premier login, cet email reçoit le rôle global **admin** (legacy — voir #310) et devient **owner** de l'organisation `default`.

**Self-service (nouvelle installation)** : tout utilisateur sans organisation peut se connecter et créer sa première organisation via `/org/new`.

Avec `REVUES_LOGIN_REQUIRE_WHITELIST=1`, l'inscription / OAuth n'est acceptée que pour un email déjà membre d'une org, avec une **invitation en attente**, ou égal au bootstrap (message générique anti-énumération).

## 4. Inviter des utilisateurs

Les administrateurs d'organisation (`owner` / `admin`) invitent depuis **Membres** (`/admin/members`) : email + rôle org (`member` / `admin` / `owner`). L'invité s'inscrit ou se connecte avec cet email, puis accepte l'invitation (parcours `/org/select`).

Le modèle d'accès complet (org = gouvernance, projet = droits métier) est décrit dans [RBAC.md](./RBAC.md) ; la bascule hors rôle global produit est suivie dans #310.

## 5. Créer un sujet et lancer une revue

1. Créer un **sujet** (`/subjects` ou `/subjects/new`)
2. Créer ou rattacher un **modèle** (`/modeles`)
3. **Lancer une revue** depuis le sujet (`/subjects/{id}/launch`) puis suivre la progression dans `/runs`

Les lecteurs (`reader`) voient les sujets auxquels ils ont accès ; ils ne cochent pas.

Admin org : `/admin` (membres, invitations, politiques, SMTP, hub intégrations) — réservé owner/admin.
