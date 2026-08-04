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

Ouvrir `http://localhost:8080/login`.

## 3. Se connecter

**Instance migrée (organisation `default` existante)** : cliquer sur **Se connecter avec GitHub** avec le compte correspondant à `REVUES_BOOTSTRAP_ADMIN_EMAIL`. Au premier login, cet email reçoit le rôle global **admin** et devient **owner** de l'organisation `default`.

**Self-service (nouvelle installation)** : tout utilisateur sans organisation peut se connecter et créer sa première organisation via `/org/new`.

## 4. Autoriser les utilisateurs

Les administrateurs d'organisation (`owner` / `admin`) gèrent la liste blanche depuis **Emails autorisés** (`/admin/users`). Ajoutez les emails GitHub des personnes autorisées à rejoindre l'organisation active, avec leur rôle global (`reader`, `editor`, `admin`).

Une personne peut aussi rejoindre si elle est déjà membre d'une organisation, ou via une équipe / grant sujet selon le modèle décrit dans [RBAC.md](./RBAC.md).

## 5. Créer un sujet et lancer une revue

1. Créer un **sujet** (hub Organisation ou `/subjects`)
2. Créer ou rattacher un **modèle** de checklist (`/modeles`)
3. **Lancer une revue** via l'assistant (`/revues/nouvelle`)

Les lecteurs (`reader`) voient les sujets auxquels ils ont accès ; ils ne cochent pas.
