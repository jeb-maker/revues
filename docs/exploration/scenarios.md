# Scénarios utilisateur — simulation multi-architecture

À utiliser en atelier : pour chaque scénario, décrire le déroulé sur **chaque** piste (voir fiches) et sur une **baseline** « application web classique » (pages serveur, en ligne).

Ne pas trancher pendant l'exercice — noter seulement ce qui marche, ce qui casse, ce qui reste flou.

---

## S1 — Revue en zone blanche réseau

**Acteurs** : Thomas (terrain), revue de 80 points.

**Contexte** : sous-sol / site industriel, réseau intermittent puis absent pendant 2 h.

**Actions** :
- Ouvrir la revue en cours
- Cocher 40 points, ajouter 3 commentaires `nok`
- Perdre le réseau au point 41
- Terminer la saisie offline
- Retrouver le réseau en fin de journée

**Contraintes à observer** :
- Perte de saisie ?
- Conflit si un collègue a coché en ligne pendant ce temps ?
- Horodatage et auteur des actions différées ?
- Message utilisateur compréhensible ?

| Piste | Déroulé (à remplir) | Bloquants | Questions |
|-------|---------------------|-----------|-----------|
| Baseline web classique | | | |
| SPA éco | | | |
| Offline sans SPA | | | |
| Sync SQLite | | | |
| DB SQLite / org | | | |

---

## S2 — Création et exécution d'un gros modèle

**Acteurs** : Marie (référente qualité).

**Contexte** : modèle de 200 points, 12 sections, filtres et réorganisation fréquents à l'édition.

**Actions** :
- Créer le modèle
- Réordonner des sections
- Lancer une revue (snapshot)
- Vérifier que la revue figée ne change pas si le modèle est encore édité ailleurs

**Contraintes** :
- Fluidité UI vs sobriété transfert
- Temps de chargement perçu
- Cohérence version / snapshot

| Piste | Déroulé (à remplir) | Bloquants | Questions |
|-------|---------------------|-----------|-----------|
| Baseline web classique | | | |
| SPA éco | | | |
| Offline sans SPA | | | |
| Sync SQLite | | | |
| DB SQLite / org | | | |

---

## S3 — Utilisateur membre de plusieurs organisations

**Acteurs** : Alex, membre de « Acme » et « Beta Consulting ».

**Contexte** : même navigateur, deux clients distincts, données strictement séparées.

**Actions** :
- Se connecter
- Basculer d'Acme à Beta
- Vérifier qu'aucune donnée Acme n'apparaît sous Beta
- Recevoir une invitation projet sur Beta pendant une session Acme

**Contraintes** :
- Isolation des données
- Session et cache navigateur
- Performance du switch

| Piste | Déroulé (à remplir) | Bloquants | Questions |
|-------|---------------------|-----------|-----------|
| Baseline web classique | | | |
| SPA éco | | | |
| Offline sans SPA | | | |
| Sync SQLite | | | |
| DB SQLite / org | | | |

---

## S4 — Restauration après incident

**Acteurs** : Admin infra, DPO client.

**Contexte** : ransomware ou suppression accidentelle ; obligation de restaurer **uniquement** l'organisation Acme à l'état J-1.

**Actions** :
- Identifier le périmètre de backup
- Restaurer sans impact sur les autres organisations
- Prouver l'intégrité des revues clôturées et de l'audit

**Contraintes** :
- RPO / RTO acceptables
- Preuve légale (audit trail)
- Procédure documentée

| Piste | Déroulé (à remplir) | Bloquants | Questions |
|-------|---------------------|-----------|-----------|
| Baseline web classique | | | |
| SPA éco | | | |
| Offline sans SPA | | | |
| Sync SQLite | | | |
| DB SQLite / org | | | |

---

## S5 — Deux tablettes sur la même revue offline

**Acteurs** : Thomas (tablette A), Sophie (tablette B), même revue.

**Contexte** : réunion sans Wi-Fi ; les deux travaillent en parallèle sur des sections différentes puis sur le **même** point.

**Actions** :
- Tous deux passent le point #12 en `nok` avec des commentaires différents
- Retour réseau ; synchronisation

**Résultat attendu (à définir en produit, pas ici)** :
- Dernier gagnant ?
- Fusion ?
- Blocage avec message ?

| Piste | Déroulé (à remplir) | Bloquants | Questions |
|-------|---------------------|-----------|-----------|
| Baseline web classique | | | |
| SPA éco | | | |
| Offline sans SPA | | | |
| Sync SQLite | | | |
| DB SQLite / org | | | |

---

## S6 — Export pour auditeur externe

**Acteurs** : Auditeur (lecture seule), Sophie (interne).

**Contexte** : revue clôturée en 2024 ; export complet hors plateforme.

**Actions** :
- Exporter historique + états finaux + métadonnées (qui / quand)
- Consulter sans compte applicatif si possible

**Contraintes** :
- Format ouvrable (CSV, PDF, archive)
- Chaîne de preuve
- Données sensibles

| Piste | Déroulé (à remplir) | Bloquants | Questions |
|-------|---------------------|-----------|-----------|
| Baseline web classique | | | |
| SPA éco | | | |
| Offline sans SPA | | | |
| Sync SQLite | | | |
| DB SQLite / org | | | |

---

## S7 — Première visite, onboarding léger

**Acteurs** : Nouvel utilisateur invité par email.

**Contexte** : première connexion, liste blanche, aucun projet encore.

**Actions** :
- OAuth / login
- Comprendre où agir en < 5 min
- Créer ou rejoindre un espace de travail

**Contraintes** :
- Pas de surcharge JS / temps de chargement
- Clarté sans formation

| Piste | Déroulé (à remplir) | Bloquants | Questions |
|-------|---------------------|-----------|-----------|
| Baseline web classique | | | |
| SPA éco | | | |
| Offline sans SPA | | | |
| Sync SQLite | | | |
| DB SQLite / org | | | |

---

## Synthèse atelier scénarios

| Scénario | Le plus impacté par… | Zone la plus floue |
|----------|----------------------|-------------------|
| S1 | | |
| S2 | | |
| S3 | | |
| S4 | | |
| S5 | | |
| S6 | | |
| S7 | | |
