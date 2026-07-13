# Exploration architecturale — phase 0

Étude large **sans décision**, **sans prise en compte de l'existant**, **sans code**.

Objectif : cartographier l'espace des possibles avant toute implémentation ou choix technique.

---

## Contenu du dossier

| Document | Rôle |
|----------|------|
| [scenarios.md](./scenarios.md) | Histoires utilisateur à simuler sur chaque piste |
| [matrice-comparative.md](./matrice-comparative.md) | Grille vide à remplir en atelier |
| [couplages.md](./couplages.md) | Interactions entre les quatre pistes |
| [questions-ouvertes.md](./questions-ouvertes.md) | Backlog de questions — sans réponse attendue ici |
| [synthese-phase0.md](./synthese-phase0.md) | Gabarit de clôture de la phase d'étude |
| [fiches/](./fiches/) | Une fiche par piste architecturale |

### Fiches

1. [SPA éco-conçue](./fiches/01-spa-eco.md)
2. [Mode déconnecté sans SPA](./fiches/02-offline-sans-spa.md)
3. [Synchronisation SQLite navigateur ↔ serveur](./fiches/03-sync-sqlite.md)
4. [Une base SQLite par organisation](./fiches/04-sqlite-par-organisation.md)

---

## Méthode d'atelier proposée

### Étape 1 — Scénarios (≈ 1 h)

Lire [scenarios.md](./scenarios.md). Pour chaque histoire, noter les contraintes réseau, sécurité, audit et multi-tenant **sans** choisir de techno.

### Étape 2 — Fiches en parallèle (≈ 2 h)

Quatre relais ou quatre créneaux : remplir les sections « À documenter en atelier » de chaque fiche.

### Étape 3 — Matrice (≈ 1 h)

Remplir [matrice-comparative.md](./matrice-comparative.md) qualitativement : `++`, `+`, `0`, `−`, `−−`. Définir les poids des critères **ensemble** si besoin.

### Étape 4 — Couplages (≈ 30 min)

Parcourir [couplages.md](./couplages.md) : quelles combinaisons sont cohérentes, exclusives ou indifférentes ?

### Étape 5 — Clôture (≈ 30 min)

Compléter [synthese-phase0.md](./synthese-phase0.md) : état des lieux, zones grises, suites **éventuelles** (hors scope de ce dossier).

---

## Ce que cette phase ne produit pas

- Pas de décision architecture
- Pas de modification du plan produit existant
- Pas de code, POC ou script
- Pas d'issues GitHub d'implémentation
- Pas de benchmarks mesurés (réservés à une phase ultérieure si un jour elle est ouverte)

---

## Légende matrice

| Symbole | Signification |
|---------|---------------|
| `++` | Très favorable |
| `+` | Plutôt favorable |
| `0` | Neutre / dépend du contexte |
| `−` | Plutôt défavorable |
| `−−` | Très défavorable |
| `?` | Pas encore évalué |
