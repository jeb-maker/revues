# Backup & restauration — Revues

Procédure opérationnelle pour sauvegarder la base SQLite et les pièces jointes, puis restaurer en cas d'incident.

## Contenu sauvegardé

| Élément | Chemin par défaut | Artefact |
|---------|-------------------|----------|
| Base SQLite | `data/revues.db` (`REVUES_DATABASE_PATH`) | `revues.sql` (dump SQL) |
| Pièces jointes | `data/attachments/` | `attachments.tar.gz` (si non vide) |

Chaque exécution crée un répertoire horodaté UTC sous `backups/` (ou `REVUES_BACKUP_DIR`) avec un fichier `manifest.txt` (métadonnées).

## Fréquence

| Environnement | Fréquence recommandée | Mécanisme |
|---------------|----------------------|-----------|
| Production | **Quotidienne** (nuit, faible charge) | `cron` ou `systemd.timer` |
| Pré-production | Hebdomadaire | Idem |
| Développement local | À la demande | Exécution manuelle |

Exemple cron (02:15 UTC, tous les jours) :

```cron
15 2 * * * cd /opt/revues && ./scripts/backup.sh >> /var/log/revues-backup.log 2>&1
```

**Recommandation** : arrêter brièvement l'application ou s'assurer qu'aucune écriture concurrente n'a lieu pendant le dump (SQLite WAL). En pratique, un dump nocturne avec charge faible suffit pour v1. Le `.dump` via `sqlite3` gère un snapshot cohérent sans arrêt obligatoire.

## Exécution

Prérequis : client `sqlite3` installé sur l'hôte.

```bash
# Depuis la racine du dépôt / répertoire d'installation
./scripts/backup.sh
```

Variables optionnelles :

| Variable | Défaut | Description |
|----------|--------|-------------|
| `REVUES_DATABASE_PATH` | `data/revues.db` | Fichier SQLite source |
| `REVUES_ATTACHMENTS_DIR` | `data/attachments` | Répertoire des pièces jointes |
| `REVUES_BACKUP_DIR` | `backups` | Répertoire de destination |
| `REVUES_BACKUP_RETENTION_DAYS` | `90` | Rétention locale (jours) |

Copier les sauvegardes **hors de la VM** (object storage, NAS, autre zone) — le script ne gère que la rétention locale.

## Rétention 90 jours

- **Politique** : conserver **90 jours** d'historique de sauvegardes.
- **Local** : `scripts/backup.sh` supprime automatiquement les répertoires de backup de plus de 90 jours (`REVUES_BACKUP_RETENTION_DAYS`).
- **Distant** : appliquer la même règle sur le stockage externe (lifecycle S3, rotation rsync, etc.).
- **Minimum** : garder au moins la dernière sauvegarde quotidienne **et** la dernière sauvegarde hebdomadaire du mois en cours.

## Restauration

### Prérequis

- Sauvegarde cible (répertoire horodaté sous `backups/` ou copie externe).
- Application **arrêtée** (`systemctl stop revues` ou équivalent).
- Sauvegarde de l'état actuel (renommer `data/` existant) avant écrasement.

### Procédure

1. **Identifier la sauvegarde**

   ```bash
   ls -lt backups/
   BACKUP=backups/2026-06-28T021500Z   # adapter
   ```

2. **Sauvegarder l'état courant (rollback possible)**

   ```bash
   mv data/revues.db "data/revues.db.bak.$(date -u +%Y%m%dT%H%M%SZ)" 2>/dev/null || true
   mv data/attachments "data/attachments.bak.$(date -u +%Y%m%dT%H%M%SZ)" 2>/dev/null || true
   ```

3. **Restaurer la base**

   ```bash
   mkdir -p data
   sqlite3 data/revues.db < "${BACKUP}/revues.sql"
   ```

4. **Restaurer les pièces jointes** (si présentes)

   ```bash
   if [[ -f "${BACKUP}/attachments.tar.gz" ]]; then
     tar -xzf "${BACKUP}/attachments.tar.gz" -C data
   fi
   ```

5. **Vérifier les permissions**

   ```bash
   chmod 640 data/revues.db
   chmod -R u=rwX,go= data/attachments 2>/dev/null || true
   ```

6. **Redémarrer et contrôler**

   ```bash
   systemctl start revues   # ou go run ./cmd/revues en dev
   curl -sf http://localhost:8080/healthz   # → ok
   ```

7. **Contrôle fonctionnel** : connexion admin, ouverture d'un projet, consultation d'une revue récente, téléchargement d'une pièce jointe si applicable.

En cas d'échec, remettre les fichiers `.bak` et investiguer (`manifest.txt`, logs applicatifs).

## Test trimestriel

**Objectif** : prouver que les backups sont restaurables, pas seulement créés.

| Étape | Action | Critère de succès |
|-------|--------|-------------------|
| 1 | Planifier un créneau (1× par trimestre calendaire) | Date notée dans le calendrier ops |
| 2 | Choisir une sauvegarde aléatoire des 30 derniers jours | Répertoire horodaté + `manifest.txt` |
| 3 | Restaurer sur un **environnement de test** (pas la prod) | Procédure ci-dessus |
| 4 | Lancer l'application contre la base restaurée | `GET /healthz` → `ok` |
| 5 | Vérifier un échantillon métier | ≥ 1 projet, 1 revue, comptage cohérent avec la prod |
| 6 | Documenter le résultat | Ticket / note ops : date, backup ID, OK/KO, durée |

En cas d'échec du test trimestriel : traiter comme incident P2 — corriger la procédure ou l'infrastructure de backup avant la prochaine fenêtre.

## Dépannage

| Symptôme | Cause probable | Action |
|----------|----------------|--------|
| `sqlite3 absent` | Client non installé | `apt install sqlite3` (Debian/Ubuntu) |
| Dump vide ou incomplet | DB verrouillée / chemin erroné | Vérifier `REVUES_DATABASE_PATH`, réessayer hors pic |
| `attachments.tar.gz` absent | Dossier vide ou inexistant | Normal si vague 3 non déployée |
| Restauration : contraintes FK | Dump partiel | Restaurer sur base vide ; ne pas réutiliser un `.db` existant |

## Références

- [CONVENTIONS.md](./CONVENTIONS.md) — chemins `data/` et variables `REVUES_*`
- [REVIEW_ADVERSE.md](./REVIEW_ADVERSE.md) — risque identifié sur l'absence de backup
- Script : [`scripts/backup.sh`](../scripts/backup.sh)
