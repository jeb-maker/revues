# Notifications email — Revues

Emails métier : revue terminée, point assigné, échéance J-1.

Config SMTP : admin org (`/admin/settings/smtp`), credentials AES-GCM.

## File durable

Même modèle que les webhooks — **pas de Redis** :

- Table `email_deliveries` (`to_address`, `subject`, `body`, `state`, `attempts`, `next_attempt_at`, `expires_at`, `organization_id`)
- Enqueue à l’émission + **première tentative immédiate**
- Drain in-process chaque **1 minute** (`notifications.StartEmailDrainScheduler`) + drain opportuniste après enqueue
- Backoff : `min(1m << (n-1), 30m)` ; TTL **24 h** ; max **5** tentatives puis `poison`
- Secret SMTP relu à chaque tentative (pas stocké sur la ligne)

Survie au restart : les lignes `pending` dues sont reprises au prochain drain.
