# Webhooks sortants — Revues

See admin `/admin/settings/webhooks`. Payload JSON with `event_id`, `event_type`, `occurred_at`, `data`.

Signature header: `X-Revues-Signature: sha256=<hmac-sha256 hex of raw body>`.

Events: `review.completed`, `review.item.nok`, `webhook.test`.

Delivery: 3 retries, 5s timeout, anti-SSRF (block private/metadata IPs, https only except localhost in dev).

## review.completed

```json
{"event_id":"uuid","event_type":"review.completed","occurred_at":"2026-06-28T12:00:00Z","data":{"review":{"id":42,"title":"…","status":"done","project_id":3,"project_name":"…","closing_note":"…","completed_at":"…"},"items":{"total":10,"ok":8,"nok":1,"na":1,"pending":0}}}
```

## review.item.nok

```json
{"event_id":"uuid","event_type":"review.item.nok","occurred_at":"…","data":{"review":{"id":42,"title":"…","status":"in_progress","project_id":3,"project_name":"…"},"item":{"id":101,"section":"…","label":"…","status":"nok","comment":"…"}}}
```

## webhook.test

```json
{"event_id":"uuid","event_type":"webhook.test","occurred_at":"…","data":{"message":"Ceci est un événement de test depuis Revues."}}
```
