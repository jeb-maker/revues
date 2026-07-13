# Matrice comparative — à remplir en atelier

Comparer les **cinq colonnes** sur les mêmes critères. Ne pas chercher une « gagnante » en phase 0 — documenter les écarts et les dépendances au contexte.

**Légende** : `++` · `+` · `0` · `−` · `−−` · `?` (non évalué)

---

## Pondération des critères (optionnel)

À définir **en groupe** si une synthèse chiffrée est utile plus tard. Laisser vide si non pertinent en phase 0.

| Critère | Poids (1–5) | Commentaire |
|---------|-------------|-------------|
| | | |
| | | |

---

## Grille principale

| Critère | Baseline web classique | SPA éco-conçue | Offline sans SPA | Sync SQLite client↔serveur | SQLite par organisation |
|---------|------------------------|----------------|------------------|----------------------------|-------------------------|
| **Sobriété** (transfert réseau, CPU client) | ? | ? | ? | ? | ? |
| **Sobriété** (empreinte serveur / hébergement) | ? | ? | ? | ? | ? |
| **Simplicité opérationnelle** (déploiement, 1 VM, ops) | ? | ? | ? | ? | ? |
| **Time-to-market** (première version utile) | ? | ? | ? | ? | ? |
| **UX** — parcours simples (cocher, commenter) | ? | ? | ? | ? | ? |
| **UX** — édition riche (gros modèles, filtres) | ? | ? | ? | ? | ? |
| **UX terrain** — latence perçue | ? | ? | ? | ? | ? |
| **Résilience réseau** (L0 retry) | ? | ? | ? | ? | ? |
| **Offline lecture** | ? | ? | ? | ? | ? |
| **Offline écriture** | ? | ? | ? | ? | ? |
| **Cohérence données** (conflits multi-utilisateurs) | ? | ? | ? | ? | ? |
| **Audit & traçabilité** | ? | ? | ? | ? | ? |
| **Sécurité** (surface d'attaque, secrets client) | ? | ? | ? | ? | ? |
| **Isolation multi-tenant** | ? | ? | ? | ? | ? |
| **Conformité** (effacement org, export, backup) | ? | ? | ? | ? | ? |
| **Évolutivité** (10 → 10 000 organisations) | ? | ? | ? | ? | ? |
| **Maintenabilité** (équipe, stack, dette) | ? | ? | ? | ? | ? |
| **Testabilité** (E2E, charge, offline) | ? | ? | ? | ? | ? |
| **Accessibilité** | ? | ? | ? | ? | ? |
| **Multi-plateforme** (navigateurs, mobile) | ? | ? | ? | ? | ? |

---

## Critères additionnels (lignes libres)

| Critère (à ajouter) | Baseline | SPA éco | Offline | Sync SQLite | DB / org |
|---------------------|----------|---------|---------|-------------|----------|
| | ? | ? | ? | ? | ? |
| | ? | ? | ? | ? | ? |
| | ? | ? | ? | ? | ? |

---

## Scénarios × pistes (rappel qualitatif)

Après [scenarios.md](./scenarios.md), résumer en une ligne par scénario :

| Scénario | Mieux servi par (hypothèse) | Pire cas (hypothèse) | Notes |
|----------|----------------------------|----------------------|-------|
| S1 Zone blanche | | | |
| S2 Gros modèle | | | |
| S3 Multi-org | | | |
| S4 Restauration | | | |
| S5 Conflit offline | | | |
| S6 Export auditeur | | | |
| S7 Onboarding | | | |

---

## Observations transverses (texte libre)

### Points de convergence

-

### Points de divergence forte

-

### Critères qui ne discriminent pas

-

### Zones où « ça dépend » du périmètre produit

-
