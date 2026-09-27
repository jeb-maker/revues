// Package store is the only package that issues SQL.
//
// Aggregate layout (keep related queries in these files — do not grow a god file):
//
//	db.go, migrate — pool, pragmas, goose
//	users.go, sessions.go, allowed_emails.go — auth identity
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
package store
