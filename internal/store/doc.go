// Package store is the only package that issues SQL.
//
// SQL access rules (rewrite ADR-001 / WP-003):
//   - Prefer sqlc: queries in internal/store/queries/, generated code in
//     internal/store/sqlc/ (regenerate with `make sqlc`).
//   - Legacy hand-written SQL remains in this package until migrated; do not
//     put SQL in handlers or features.
//
// Aggregate layout (keep related queries in these files — do not grow a god file):
//
//	db.go, migrate — pool, pragmas, goose
//	users.go, sessions.go, allowed_emails.go — auth identity (users/sessions → sqlc)
//	organizations.go, organization_invitations.go, teams.go — multi-tenant
//	subjects.go, subject_access.go, tags.go — subjects & access
//	checklist_templates.go — models / versions / items
//	runs.go, run_items.go, run_item_events.go, run_export.go — reviews
//	attachments.go, integrations.go, integration_links.go — companions
//	settings.go — encrypted org settings
//	webhook_deliveries.go, email_deliveries.go — durable outbound queues
//	dashboard.go — hub aggregations
//
// Feature packages define small consumer interfaces; *Store satisfies them.
//
// Remaining hand-written SQL (migrate incrementally in follow-up area:data issues):
// allowed_emails, organizations*, teams, subjects*, tags, checklist_templates,
// runs*, attachments, integrations*, settings, webhook/email deliveries, dashboard.
package store
