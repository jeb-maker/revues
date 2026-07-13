# Fiche 02 — Mode déconnecté sans SPA

**Statut** : exploration · **Décision** : aucune · **Code** : aucun

---

## Définition

Permettre une utilisation **utile** sans connexion réseau (ou avec réseau intermittent), en restant sur une architecture **sans framework SPA** : pages HTML, requêtes fragment (HTMX ou équivalent), JavaScript minimal, Service Worker, stockage navigateur.

---

## Niveaux d'offline (échelle L0 → L4)

| Niveau | Nom | Comportement | Complexité typique |
|--------|-----|--------------|-------------------|
| **L0** | Résilience | Retry, message « hors ligne », pas de perte de saisie **en cours** | Faible |
| **L1** | Lecture | Consulter contenu déjà vu / pré-téléchargé | Moyenne |
| **L2** | Écriture différée | Cocher, commenter ; envoi au retour réseau | Moyenne–élevée |
| **L3** | Création offline | Nouvelle revue / modèle sans serveur | Élevée |
| **L4** | Autonomie longue | Jours sans sync, résolution conflits | Très élevée |

**Question centrale atelier** : quel niveau minimum est **utile** vs marketing ?

---

## Briques techniques (sans SPA)

```
┌────────────────────────────────────────────────────────────┐
│ Navigateur                                                 │
├────────────────────────────────────────────────────────────┤
│ HTML + fragments    │ Service Worker (cache GET)           │
│ Petit JS vanilla    │ IndexedDB / localStorage             │
│ Background Sync API │ File d'attente « outbox » (POST)     │
│ Web Locks           │ Éviter double envoi                    │
└────────────────────────────────────────────────────────────┘
```

### Pattern A — Cache de pages (Service Worker)

- Intercepter les GET, servir depuis le cache si offline
- **Limite** : pages dynamiques, personnalisation, auth, CSRF

### Pattern B — Outbox HTTP

- Sérialiser les mutations (formulaires) en IndexedDB
- Au événement `online`, rejouer dans l'ordre vers le serveur
- **Nécessite** : API idempotente, gestion conflits, feedback utilisateur

### Pattern C — Pack offline explicite

- L'utilisateur **télécharge** un bundle pour une revue (HTML + JSON + assets)
- Pas de sync temps réel : reimport ou sync manuelle au retour
- Plus simple juridiquement et opérationnellement

### Pattern D — Client hors navigateur

- App desktop ou mobile native — hors scope « sans SPA » web mais comparable fonctionnellement

---

## Tensions transverses

| Tension | Description |
|---------|-------------|
| **Auth** | Session cookie expire ; jeton longue durée = risque |
| **CSRF** | Jeton périmé après longue offline |
| **Conflits** | Deux acteurs modifient le même point |
| **Audit** | Horodatage et auteur des actions différées |
| **RBAC** | Droits révoqués pendant l'absence réseau |
| **Cache** | Taille revue + pièces jointes |
| **UX** | Chaque page = unité de cache vs shell unifié SPA |

---

## Offline sans SPA : avantages hypothétiques

- Pas de gros runtime client (WASM SQL, framework)
- Aligné avec sobriété transfert si pages légères
- Moins de surface « état global » à synchroniser
- Déploiement serveur inchangé (pas de build front lourd)

---

## Offline sans SPA : difficultés hypothétiques

- Cache de pages **personnalisées** (RBAC) complexe
- Outbox à concevoir par type d'action métier
- UX moins fluide qu'un shell client unique
- Tests E2E offline multi-pages plus laborieux

---

## Lien avec les autres fiches

| Fiche | Lien |
|-------|------|
| [01 SPA éco](./01-spa-eco.md) | SPA facilite état local unifié ; sans SPA = outbox + cache par page |
| [03 Sync SQLite](./03-sync-sqlite.md) | L2+ offline écriture pousse vers sync structurée ou outbox métier |
| [04 DB / org](./04-sqlite-par-organisation.md) | Pack offline pourrait être **un export** par org ou par revue |

---

## Scénarios [scenarios.md](../scenarios.md) — notes atelier

| Scénario | Niveau L minimal ? | Pattern A/B/C ? | Notes |
|----------|-------------------|-----------------|-------|
| S1 Zone blanche | | | |
| S2 Gros modèle | | | |
| S3 Multi-org | | | |
| S4 Restauration | | | |
| S5 Conflit offline | | | |
| S6 Export auditeur | | | |
| S7 Onboarding | | | |

---

## À documenter en atelier

- [ ] Niveau L cible (0–4) par type d'utilisateur
- [ ] Offline **par revue** vs application entière
- [ ] Comportement conflit (S5) acceptable côté produit
- [ ] Pack exportable (pattern C) : suffisant pour le terrain ?

---

## Notes libres

-
