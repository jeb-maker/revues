# Questions ouvertes — backlog phase 0

Liste de questions **sans réponse attendue dans ce dossier**. À cocher quand une réponse émerge en atelier ou en phase ultérieure.

---

## Produit & usages

- [ ] Quels parcours doivent fonctionner **sans réseau** du tout ?
- [ ] Offline **lecture seule** suffit-il pour une large part du terrain ?
- [ ] Combien de temps maximum sans sync (minutes, heures, jours) ?
- [ ] Collaboration offline sur **la même** revue est-elle un besoin réel ?
- [ ] Qui tranche en cas de conflit (dernier gagnant, manuel, règle métier) ?

---

## SPA éco-conçue

- [ ] Quels écrans exigent une interactivité que le web classique ne couvre pas ?
- [ ] Une SPA **partielle** (un seul module) est-elle suffisante ?
- [ ] Quel budget JS (Ko gzip) est acceptable pour l'éco-conception visée ?
- [ ] SSR + hydratation minimale est-il dans le périmètre « SPA » ?
- [ ] Comment mesurer l'éco (Lighthouse, transfert par parcours, autre) ?

---

## Offline sans SPA

- [ ] Quel niveau cible : L0, L1, L2, L3 ou L4 ? (voir fiche 02)
- [ ] Service Worker : périmètre de cache (assets, pages, API) ?
- [ ] Outbox : ordre des opérations, idempotence API ?
- [ ] Pack offline **par revue** vs application entière ?
- [ ] Comment gérer auth et CSRF après une longue période offline ?

---

## Sync SQLite navigateur ↔ serveur

- [ ] Sync **continue** ou **à la demande** (bouton synchroniser) ?
- [ ] Granularité : par revue, projet, organisation ?
- [ ] Même schéma SQL des deux côtés : gain ou coût (migrations × 2) ?
- [ ] CRDT / event sourcing / last-write-wins : quel modèle pour les statuts de points ?
- [ ] Pièces jointes : hors SQL ou dans le même flux ?
- [ ] Solutions tierces (ElectricSQL, PowerSync, etc.) : dans le périmètre d'étude marché ?

---

## SQLite par organisation

- [ ] Isolation **juridique** exige-t-elle un fichier par client ?
- [ ] Ordre de grandeur : nombre d'organisations, taille DB médiane ?
- [ ] Utilisateur dans N orgs : N connexions / N caches — acceptable ?
- [ ] Backup/restore par org : qui opère (client, hébergeur, les deux) ?
- [ ] Registry central (users, sessions) + DB métier par org : modèle retenu ?
- [ ] Seuil de bascule vers moteur autre que SQLite (si un jour pertinent) ?

---

## Sécurité & conformité

- [ ] Données dans le navigateur (IndexedDB, .db local) : classification RGPD ?
- [ ] Export d'un fichier .db au client : risque acceptable ?
- [ ] Révocation de droits pendant une session offline : comportement attendu ?
- [ ] Audit trail : les actions offline doivent-elles porter l'heure locale ou serveur au sync ?

---

## Opérations & coût

- [ ] Une VM, un binaire : contrainte durable ou hypothèse provisoire ?
- [ ] Migrations appliquées sur N fichiers .db : procédure acceptable ?
- [ ] Monitoring et support : visibilité cross-org nécessaire ?

---

## Méthode

- [ ] Phase 0 considérée comme suffisante quand ?
- [ ] Une phase 1 (décision) aura-t-elle besoin de POC **hors** repo principal ?
- [ ] Entretiens utilisateurs prévus avant toute décision ?

---

## Réponses émergentes (à compléter en atelier)

| # | Question | Réponse / orientation | Date | Qui |
|---|----------|----------------------|------|-----|
| | | | | |
