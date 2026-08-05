# Webhooks sortants — Revues

See admin `/admin/settings/webhooks`. Payload JSON with `event_id`, `event_type`, `occurred_at`, `data`.

Signature header: `X-Revues-Signature: sha256=<hmac-sha256 hex of raw body>`.

Events: `review.completed`, `review.item.nok`, `webhook.test`.

## Delivery & durable retry

- HTTP timeout 5s, max 1 redirect, anti-SSRF (block private/metadata IPs; `https` only except `http://localhost` in dev).
- **Anti-SSRF is re-checked on every attempt** (URL scheme + DNS/IP + dial).
- Queue table: `webhook_deliveries` (payload + `state` / `attempts` / `next_attempt_at` / `expires_at`).
- Drain: in-process cron every **1 minute** (`webhooks.StartDrainScheduler`) and opportunistic drain after emit — **same binary**, no Redis / separate worker.
- Secret HMAC is loaded from settings at attempt time (not stored on the row).

### Backoff

After each failed attempt `n` (1-based count of failures so far), next try is delayed by:

| After attempt | Delay |
|---------------|-------|
| 1 | 1 min |
| 2 | 2 min |
| 3 | 4 min |
| 4 | 8 min |
| … | … capped at **30 min** |

Formula: `min(1m << (n-1), 30m)`.

### TTL

A delivery expires **24 hours** after enqueue (`DeliveryTTL`). If the next backoff would land past TTL, or a due row is past `expires_at`, the row is marked **poison**.

### Max attempts & poison

- Hard cap: **5** HTTP attempts (`MaxAttempts`), including the first try right after enqueue.
- **Poison** (`state = poison`): no further retries. Triggers:
  - `attempts >= MaxAttempts`
  - TTL expired
  - permanent policy errors (scheme forbidden, blocked IP, localhost in prod)
- Successful delivery: `state = done`, `success = 1`.
- Pending: `state = pending` with `next_attempt_at` in the future.

Poison rows are kept for ops inspection; there is no automatic replay UI in this issue.

## review.completed

```json
{"event_id":"uuid","event_type":"review.completed","occurred_at":"2026-06-28T12:00:00Z","data":{"review":{"id":42,"title":"…","status":"done","subject_id":3,"subject_name":"…","closing_note":"…","completed_at":"…"},"items":{"total":10,"ok":8,"nok":1,"na":1,"pending":0}}}
```

## review.item.nok

```json
{"event_id":"uuid","event_type":"review.item.nok","occurred_at":"…","data":{"review":{"id":42,"title":"…","status":"in_progress","subject_id":3,"subject_name":"…"},"item":{"id":101,"section":"…","label":"…","status":"nok","comment":"…"}}}
```

## webhook.test

```json
{"event_id":"uuid","event_type":"webhook.test","occurred_at":"…","data":{"message":"Ceci est un événement de test depuis Revues."}}
```
