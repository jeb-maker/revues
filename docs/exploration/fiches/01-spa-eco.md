# Fiche 01 — SPA éco-conçue

**Statut** : exploration · **Décision** : aucune · **Code** : aucun

---

## Définition

Application web dont la navigation et une partie importante de l'interactivité s'exécutent **côté navigateur**, avec chargement initial d'un shell JavaScript, tout en visant des budgets de ressources **volontairement bas** (transfert, CPU, mémoire).

« Éco-conçue » n'est pas synonyme de « SPA » : c'est une **contrainte** qu'on peut appliquer (ou non) à différentes architectures.

---

## Variantes dans le spectre

| Variante | Description | Ordre de grandeur JS |
|----------|-------------|----------------------|
| SPA classique | Framework complet + routeur + state global | 200 Ko – 2 Mo+ |
| SPA minimaliste | Preact, Solid, Svelte compilé, Alpine ciblé | 30 – 80 Ko |
| SPA shell | Un HTML, routes client, données via `fetch` | 50 – 150 Ko |
| Islands / partial hydration | HTML serveur + îlots interactifs | Variable |
| PWA + SPA | Shell + Service Worker + cache | + SW |

---

## Problèmes que cette piste pourrait adresser

- Listes longues (centaines de points) : filtres, tri, recherche **sans** rechargement complet
- Éditeur de modèle riche : réordonnancement, sections, prévisualisation
- Formulaires multi-étapes avec état conservé entre étapes
- Transitions fluides, moins de « flash » blanc entre vues
- Base naturelle pour **état local** si offline ou sync client envisagés plus tard
- API stable → clients multiples (web, mobile, intégrations)

---

## Risques et coûts

| Domaine | Risque |
|---------|--------|
| Éco réelle | Bundle mal maîtrisé → plus lourd qu'un HTML serveur pour un simple « cocher » |
| Sécurité | Tentation de logique métier côté client ; RBAC doit rester serveur |
| Ops | Tooling build (Vite, TS), CI E2E, versions dépendances |
| Accessibilité | Focus, annonces SR, états loading à concevoir explicitement |
| SEO | Peu pertinent pour app authentifiée |
| Équipe | Compétences front modernes vs templates serveur |

---

## Leviers « éco » si SPA retenue un jour

1. **Budget chiffré** : JS initial gzip < ? Ko ; lazy-load par route
2. **Pas de lib UI lourde** : CSS maison ou micro-lib
3. **Données** : pagination stricte, pas de chargement intégral revue
4. **SSR ou prerender** : premier paint rapide, pas page blanche
5. **Mesure** : transfert par parcours, Lighthouse, CPU machine modeste

---

## Ce que cette piste ne résout pas seule

- Offline (nécessite SW, stockage local, sync — voir fiches 02 et 03)
- Isolation multi-tenant (voir fiche 04)
- Audit serveur (les écritures passent toujours par l'API)

---

## Scénarios [scenarios.md](../scenarios.md) — notes atelier

| Scénario | Pertinence de la SPA | Notes |
|----------|---------------------|-------|
| S1 Zone blanche | | |
| S2 Gros modèle | | |
| S3 Multi-org | | |
| S4 Restauration | | |
| S5 Conflit offline | | |
| S6 Export auditeur | | |
| S7 Onboarding | | |

---

## Comparatif rapide : SPA complète vs SPA partielle vs pas de SPA

| | SPA complète | SPA partielle (1 module) | Pas de SPA |
|--|--------------|--------------------------|------------|
| Coût initial | Élevé | Moyen | Faible |
| UX édition riche | Forte | Ciblée | Dépend HTMX / forms |
| Éco parcours simple | Faible | Moyenne | Forte |
| Offline | Plus facile | Module isolé | Plus difficile |

---

## À documenter en atelier

- [ ] Quels écrans **nécessitent** une SPA dans notre métier (revues / modèles / admin) ?
- [ ] Budget JS acceptable pour « éco » ?
- [ ] SPA partielle suffisante ?
- [ ] Comment mesurer l'éco avant / après ?

---

## Références conceptuelles (hors repo)

- Progressive enhancement vs client-only
- « Resilient Web Design » (offline, couches)
- Budgets performance web (Core Web Vitals comme inspiration, pas comme objectif produit)

---

## Notes libres

-
