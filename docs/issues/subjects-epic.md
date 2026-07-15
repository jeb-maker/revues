# Épique — Sujets (remplacement projets, greenfield)

Issues à créer sur GitHub (repo `jeb-maker/revues`). **Ordre d'implémentation strict.**

```bash
./scripts/create-subjects-issues.sh
```

---

## Epic — `[Epic] Sujets — remplacement projets (greenfield)`

**Labels** : `epic`, `vague-subjects`, `area:data`, `area:core`

### Contexte produit

Remplacer **Projet** par **Sujet** (chose revue : site, matériel, app, usage solo).

- **Étiquettes** : classification libre (filtrer, retrouver) — jamais accès
- **Domaines** : matching modèles ↔ sujet
- **Accès v1** : membre org → voit tous sujets de l'org ; `project_members` supprimé
- **Libellé UI** : preset org (`sujet`, `cible`, …) — injection `{{.Labels.Subject.*}}` ; admin preset en issue 6 (v1b)
- **Greenfield** : base jetable, breaking webhooks OK

Spec plan : `.cursor/plans/` ou discussion produit Sujets/tags.

### Issues filles (ordre)

1. Schéma + canonical.sql
2. Store
3. RBAC + IDOR
4. Handlers + wizard
5. UI + intégrations + docs
6. (v1b) Preset libellé admin org

### Hors scope épique

- Séries / récurrence (`review_series`)
- Programmes multi-sujets (`campaigns`)
- `team_subject_roles`, sujets privés (rebasé sur épique équipes v2)
- Fusion sujets doublons

### Prompt agent type

```
Repo jeb-maker/revues. Implémente UNIQUEMENT l'issue #N.
Lis AGENTS.md, docs/CONVENTIONS.md, docs/GO.md, docs/DEFINITION_OF_DONE.md.
Spec épique : docs/issues/subjects-epic.md.
Branche : cursor/issue-N-subjects-<slug>-f21b
PR titre = titre issue, corps : Closes #N
./scripts/check.sh doit passer avant push.
```

---

## Issue 1 — `[data] Schéma sujets greenfield`

**Labels** : `vague-subjects`, `area:data`

### Contexte

Migration destructive greenfield : `projects` → `subjects`, séparation étiquettes/domaines.

### Critères d'acceptation

- [ ] `migrations/00014_subjects_greenfield.sql` :
  - `projects` → `subjects`
  - `project_tags` → `subject_domains`
  - `template_tags` → `template_domains`
  - Nouvelle table `subject_tags` (étiquettes descriptives)
  - `checklist_runs.project_id` → `subject_id`
  - `DROP project_members`
  - Retirer ou nullifier `organization_invitations.project_id` / `project_role`
- [ ] `docs/schema/canonical.sql` aligné (seule PR autorisée sur ce fichier)
- [ ] `./scripts/check.sh` vert (adapter tests DB minimalement si cassés — stubs OK, pas de store complet)

### Hors scope

- Handlers, UI, store Go (issue 2)
- RBAC

---

## Issue 2 — `[store] Store sujets et domaines`

**Labels** : `vague-subjects`, `area:data`, `area:core`

**Bloqué par** : Issue 1

### Critères d'acceptation

- [ ] `internal/store/subjects.go` — CRUD sujet, archivage, list avec filtres
- [ ] `internal/store/tags.go` — `ListSubjectTags`, `SetSubjectTags`, `ListSubjectDomains`, `SetSubjectDomains`, `ListTemplateDomains`, `SetTemplateDomains`
- [ ] Adapter `runs.go`, `dashboard.go`, `checklist_templates.go`, `organization_invitations.go` (`subject_id`)
- [ ] Supprimer `projects.go`, `projects_test.go`, `projects_idor_test.go`
- [ ] `ListChecklistTemplates(ctx, subjectID)` — intersection `subject_domains` ↔ `template_domains` ; sans domaine = tous modèles
- [ ] Tests store table-driven verts

### Hors scope

- Package `internal/features/subjects`
- UI

---

## Issue 3 — `[auth] RBAC sujets v1 + tests IDOR`

**Labels** : `vague-subjects`, `area:auth`, `area:core`

**Bloqué par** : Issue 2

### Critères d'acceptation

- [ ] `internal/features/subjects/service.go` :
  - `CanViewSubject` : admin global OU membre org du sujet (org active session)
  - `CanManageSubject` : admin global OU org owner/admin OU editor+ global
  - `CanLaunchRun` : CanView + editor+ global
- [ ] Adapter `runs/service.go`, `internal/templates/access.go`
- [ ] Pas de `project_members` / `subject_members` en v1
- [ ] Tests : `TestIDOR_CrossSubject`, adapter `rbac_test.go`, `runs/access_test.go`
- [ ] `./scripts/check.sh` vert

### Hors scope

- Handlers HTTP (issue 4)
- Mise à jour RBAC.md (issue 5)

---

## Issue 4 — `[features] Handlers sujets + wizard revue`

**Labels** : `vague-subjects`, `area:core`, `area:ui`

**Bloqué par** : Issue 3

### Critères d'acceptation

- [ ] Package `internal/features/subjects/` (remplace `projects/`)
- [ ] Routes `/subjects/...` ; supprimer `/projects/...` et routes membres
- [ ] Wizard :
  - `GET/POST /revues/nouvelle` — choix/création inline sujet (nom seul)
  - `GET /subjects/{id}/modeles?for_run=1`
  - `POST /subjects/{id}/revues`
- [ ] CRUD sujet complet
- [ ] Recherche sujets (LIKE name) à la création
- [ ] CSRF tous POST
- [ ] Adapter `runs/handlers.go`, `checklisttemplates/handlers.go`, `organizations/handlers.go`, `router.go`
- [ ] Tests handlers verts

### Hors scope

- Templates HTML renommés (issue 5) — handlers peuvent référencer noms templates finaux
- Webhooks payload

---

## Issue 5 — `[ui][integrations] Templates, intégrations, docs`

**Labels** : `vague-subjects`, `area:ui`, `area:integrations`, `area:core`

**Bloqué par** : Issue 4

### Critères d'acceptation

- [ ] Templates : `subjects_*`, `run_wizard_subjects`, nav « Sujets », run_show, etc.
- [ ] `SubjectUILabels` injecté layout — **aucun** libellé sujet en dur ; défaut preset `sujet`
- [ ] Domaines en `<details>` options avancées au wizard
- [ ] Webhooks : `subject_id`, `subject_name` ; notifications, notion export, CSV
- [ ] Docs : `PLAN.md`, `RBAC.md` (section Sujets v1), `CONVENTIONS.md`, `AGENTS.md`
- [ ] `cmd/seed/main.go` — sujets démo
- [ ] Note dans `docs/issues/access-teams-epic.md` : rebaser sur `subjects`
- [ ] `grep -r project_id internal/ web/` → 0 hors migrations historiques
- [ ] `./scripts/check.sh` vert

### Hors scope

- Écran admin preset libellé (issue 6)

---

## Issue 6 — `[ui] Preset libellé sujet (admin org)` (v1b)

**Labels** : `vague-subjects`, `area:ui`, `area:admin`

**Bloqué par** : Issue 5

### Critères d'acceptation

- [ ] Stockage org : `ui_subject_label` ∈ {`sujet`,`cible`,`entite`,`asset`}
- [ ] Écran admin org : select preset
- [ ] Layout résout preset → `SubjectUILabels` (singulier, pluriel, hint)
- [ ] Tests preset `cible` affiché en nav

### Hors scope

- Texte libre custom
- i18n EN

---
