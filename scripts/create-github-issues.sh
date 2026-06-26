#!/usr/bin/env bash
# Crée labels, milestones et issues GitHub pour Revues
set -euo pipefail

REPO="jeb-maker/revues"

labels=(
  "epic:7B68EE"
  "vague-1:0E8A16"
  "vague-2:1D76DB"
  "vague-3:5319E7"
  "area:infra:EDEDED"
  "area:data:EDEDED"
  "area:auth:BFD4F2"
  "area:core:FEF2C0"
  "area:ui:C2E0C6"
  "area:admin:F9D0C4"
  "area:integrations:D4C5F9"
  "area:notifications:FBCA04"
  "area:attachments:FBCA04"
  "good first issue:7057FF"
)

for entry in "${labels[@]}"; do
  name="${entry%%:*}"
  color="${entry##*:}"
  gh label create "$name" --color "$color" --repo "$REPO" 2>/dev/null || true
done

m1=$(gh api repos/$REPO/milestones -f title="Vague 1 — Cœur métier" -f description="Revues complètes, auth, audit" -f state=open --jq .number)
m2=$(gh api repos/$REPO/milestones -f title="Vague 2 — Admin & intégrations" -f description="SMTP, Jira, webhooks" -f state=open --jq .number)
m3=$(gh api repos/$REPO/milestones -f title="Vague 3 — Companion & fichiers" -f description="Notion, pièces jointes" -f state=open --jq .number)

create_issue() {
  local title="$1" body="$2" labels="$3" milestone="$4"
  gh issue create --repo "$REPO" --title "$title" --body "$body" --label "$labels" --milestone "$milestone"
}

# Epics
E1=$(create_issue "Epic: Vague 1 — Cœur métier" "Regroupe toutes les issues de la vague 1.

Voir docs/ROADMAP.md

## Critère de fin
Marie crée un modèle, Thomas exécute une revue, Sophie consulte l'historique." "epic,vague-1" "$m1")
E2=$(create_issue "Epic: Vague 2 — Admin & intégrations" "SMTP, Jira (Cloud + Server/DC), webhooks.

Voir docs/ROADMAP.md" "epic,vague-2" "$m2")
E3=$(create_issue "Epic: Vague 3 — Companion & fichiers" "Export/import Notion, pièces jointes compressées.

Voir docs/ROADMAP.md" "epic,vague-3" "$m3")

echo "Epics: $E1 $E2 $E3"

# Vague 1
create_issue "Bootstrap projet Go (chi, templates, static, healthz)" "## Objectif
Initialiser le squelette applicatif Go.

## Critères d'acceptation
- [ ] \`cmd/revues/main.go\` démarre un serveur HTTP
- [ ] Router chi avec route \`GET /healthz\` → 200
- [ ] Servir \`web/static/\` et templates de base
- [ ] README avec instructions \`go run\`

## Dépendances
Aucune

## Epic
$E1" "vague-1,area:infra,good first issue" "$m1"

create_issue "Schéma DB SQLite + migrations goose" "## Objectif
Créer le schéma initial et le système de migrations.

## Critères d'acceptation
- [ ] Tables : users, projects, project_members, checklist_templates, template_versions, template_items, checklist_runs, run_items, run_item_events
- [ ] Migrations goose numérotées, exécution au démarrage
- [ ] PRAGMA foreign_keys + WAL activés
- [ ] Index sur project_id, run_id, version_id

## Dépendances
Bloqué par bootstrap Go

## Epic
$E1" "vague-1,area:data" "$m1"

create_issue "Auth GitHub OAuth + sessions + CSRF" "## Objectif
Connexion via GitHub, sessions serveur sécurisées.

## Critères d'acceptation
- [ ] Flux OAuth Authorization Code + PKCE
- [ ] Table sessions en base, cookie HttpOnly/Secure/SameSite
- [ ] Token CSRF sur formulaires POST
- [ ] Routes /login, /auth/github/callback, /logout

## Dépendances
Bootstrap + schéma DB

## Epic
$E1" "vague-1,area:auth" "$m1"

create_issue "RBAC global + middleware RequireRole" "## Objectif
Contrôle d'accès admin/editor/reader côté serveur.

## Critères d'acceptation
- [ ] Rôle stocké sur users, défaut reader
- [ ] Middleware RequireAuth + RequireRole(\"editor\")
- [ ] Tests table-driven sur matrice permissions

## Dépendances
Auth GitHub

## Epic
$E1" "vague-1,area:auth" "$m1"

create_issue "Admin liste blanche utilisateurs" "## Objectif
Gérer les emails autorisés et leurs rôles.

## Critères d'acceptation
- [ ] Écran admin : ajouter/retirer email + rôle
- [ ] Première connexion OAuth : refus si email non autorisé
- [ ] Premier admin bootstrap via env ou migration

## Dépendances
Auth + RBAC

## Epic
$E1" "vague-1,area:admin" "$m1"

create_issue "CRUD projets + membres + rôles locaux" "## Objectif
Gérer projets et appartenance (lead/contributor/viewer).

## Critères d'acceptation
- [ ] CRUD projet (nom, description, archivage)
- [ ] Ajouter/retirer membres avec rôle local
- [ ] Vérification permissions sur chaque action

## Dépendances
RBAC

## Epic
$E1" "vague-1,area:core" "$m1"

create_issue "Modèles versionnés (templates, sections, items)" "## Objectif
CRUD modèles de check-list avec versionnement.

## Critères d'acceptation
- [ ] Créer/modifier modèle = nouvelle template_version
- [ ] Sections et items ordonnés
- [ ] Éditeur HTML serveur
- [ ] Archivage modèle

## Dépendances
Schéma DB + RBAC

## Epic
$E1" "vague-1,area:core" "$m1"

create_issue "Lancer revue (snapshot SQL des items)" "## Objectif
Instancier une revue sur un projet à partir d'un modèle versionné.

## Critères d'acceptation
- [ ] Assistant 3 étapes : projet → modèle → lancer
- [ ] INSERT run_items depuis template_items (snapshot)
- [ ] template_version_id figé sur checklist_runs
- [ ] Statuts revue : draft → in_progress → done

## Dépendances
Projets + modèles

## Epic
$E1" "vague-1,area:core" "$m1"

create_issue "Statuts ok/nok/na + commentaire obligatoire si nok" "## Objectif
Gérer les points d'une revue avec règles métier.

## Critères d'acceptation
- [ ] Statuts pending/ok/nok/na
- [ ] Commentaire requis si nok (validation serveur)
- [ ] Note globale sur la revue à la clôture
- [ ] Résumé nok non résolus avant clôture

## Dépendances
Lancer revue

## Epic
$E1" "vague-1,area:core" "$m1"

create_issue "Assignation par point + vue Mes tâches" "## Objectif
Assigner des points et afficher les tâches de l'utilisateur.

## Critères d'acceptation
- [ ] Champ assigned_to sur run_items
- [ ] Écran /mes-taches listant les points assignés
- [ ] Filtre par projet et statut

## Dépendances
Statuts points

## Epic
$E1" "vague-1,area:core" "$m1"

create_issue "Audit trail (run_item_events)" "## Objectif
Historiser chaque changement de statut.

## Critères d'acceptation
- [ ] Table run_item_events peuplée à chaque changement
- [ ] Affichage historique sur le détail d'un point
- [ ] Transaction atomique update + insert event

## Dépendances
Statuts points

## Epic
$E1" "vague-1,area:core" "$m1"

create_issue "UI HTMX (cocher, commenter sans reload)" "## Objectif
Interactions fluides sans SPA.

## Critères d'acceptation
- [ ] HTMX chargé (< 15 Ko)
- [ ] Cocher/changer statut sans rechargement complet
- [ ] Édition commentaire inline
- [ ] Barre de progression mise à jour

## Dépendances
Statuts points

## Epic
$E1" "vague-1,area:ui" "$m1"

create_issue "Tableau de bord + fiche projet" "## Objectif
Vues synthèse projets et revues en cours.

## Critères d'acceptation
- [ ] Dashboard : projets, revues en cours, % complété
- [ ] Fiche projet : revues, membres, points bloquants (nok)
- [ ] 3 onglets : Projets | Mes tâches | Modèles

## Dépendances
Lancer revue + assignation

## Epic
$E1" "vague-1,area:ui" "$m1"

# Vague 2
create_issue "Écran admin SMTP (config chiffrée + test email)" "## Objectif
Permettre à l'admin de configurer le relais SMTP.

## Critères d'acceptation
- [ ] Champs : hôte, port, TLS, user, password chiffré, from
- [ ] Table settings, secret via env
- [ ] Bouton envoyer email de test
- [ ] App fonctionne sans SMTP (notifications désactivées)

## Epic
$E2" "vague-2,area:admin" "$m2"

create_issue "Notifications email (revue terminée, assignation, échéance)" "## Objectif
Emails déclenchés par événements métier.

## Critères d'acceptation
- [ ] Email revue terminée → membres projet
- [ ] Email point assigné → assigné
- [ ] Email échéance J-1 → responsable
- [ ] Envoi async (goroutine), log si échec

## Dépendances
SMTP admin

## Epic
$E2" "vague-2,area:notifications" "$m2"

create_issue "Config Jira admin (Cloud vs Server/DC)" "## Objectif
Configurer Jira Cloud ou Server/Data Center.

## Critères d'acceptation
- [ ] Choix type instance (cloud/server)
- [ ] Cloud : URL + email + API token
- [ ] Server/DC : URL + PAT
- [ ] Credentials chiffrés, bouton tester connexion

## Epic
$E2" "vague-2,area:integrations" "$m2"

create_issue "Jira : lier une issue sur un point" "## Objectif
Associer PROJ-123 ou URL Jira à un run_item.

## Critères d'acceptation
- [ ] Champ lien issue sur point nok ou tout point
- [ ] Table integration_links
- [ ] Affichage lien cliquable vers Jira
- [ ] Client API adaptateur Cloud/Server

## Dépendances
Config Jira

## Epic
$E2" "vague-2,area:integrations" "$m2"

create_issue "Jira : créer ticket depuis point nok" "## Objectif
Créer une issue Jira pré-remplie depuis un nok.

## Critères d'acceptation
- [ ] Bouton \"Créer ticket Jira\" sur point nok
- [ ] Titre/description pré-remplis (projet, revue, commentaire)
- [ ] Lien issue stocké automatiquement
- [ ] Gestion erreurs API

## Dépendances
Config Jira + lier issue

## Epic
$E2" "vague-2,area:integrations" "$m2"

create_issue "Webhooks sortants (review.completed + review.item.nok)" "## Objectif
Notifier des URLs externes sur événements.

## Critères d'acceptation
- [ ] Événements review.completed et review.item.nok
- [ ] Payload JSON documenté, signature HMAC
- [ ] Config : URLs, secret, cases par événement
- [ ] Retry 3x, log échecs, bouton test

## Epic
$E2" "vague-2,area:integrations" "$m2"

create_issue "Admin intégrations (UI unifiée)" "## Objectif
Écran admin regroupant SMTP, Jira, webhooks.

## Critères d'acceptation
- [ ] Page /admin/integrations
- [ ] Statut activé/désactivé par intégration
- [ ] Liens vers sous-configs
- [ ] Réservé admin

## Dépendances
SMTP, Jira, webhooks

## Epic
$E2" "vague-2,area:admin" "$m2"

# Vague 3
create_issue "Config Notion admin (token, workspace)" "## Objectif
Configurer l'accès API Notion.

## Critères d'acceptation
- [ ] Token integration Notion chiffré
- [ ] Bouton tester connexion
- [ ] Documentation mapping champs

## Epic
$E3" "vague-3,area:integrations" "$m3"

create_issue "Export revue clôturée vers Notion" "## Objectif
Archiver une revue terminée en page Notion.

## Critères d'acceptation
- [ ] Bouton export à la clôture ou depuis détail revue done
- [ ] Page Notion structurée (projet, date, points, statuts)
- [ ] URL Notion stockée sur la revue

## Dépendances
Config Notion

## Epic
$E3" "vague-3,area:integrations" "$m3"

create_issue "Import modèle depuis DB Notion" "## Objectif
Créer un template Revues depuis une database Notion.

## Critères d'acceptation
- [ ] Saisie URL/database ID Notion
- [ ] Mapping colonnes → sections/items
- [ ] Crée template_version v1
- [ ] Preview avant import

## Dépendances
Config Notion + modèles versionnés

## Epic
$E3" "vague-3,area:integrations" "$m3"

create_issue "Upload pièces jointes + compression images" "## Objectif
Joindre des fichiers à un point, images compressées.

## Critères d'acceptation
- [ ] Upload max 5 Mo, types jpeg/png/webp/pdf
- [ ] Compression images (max 1920px, webp/jpeg)
- [ ] Stockage data/attachments/
- [ ] 1 pièce jointe par point

## Epic
$E3" "vague-3,area:attachments" "$m3"

create_issue "Affichage pièces jointes dans détail revue" "## Objectif
Voir et télécharger les pièces jointes.

## Critères d'acceptation
- [ ] Affichage miniature images
- [ ] Lien téléchargement PDF
- [ ] Permissions alignées sur accès revue

## Dépendances
Upload pièces jointes

## Epic
$E3" "vague-3,area:attachments" "$m3"

echo "Done."
