-- +goose Up
-- Preserve access for subjects that previously relied on org_member_legacy
-- (ungated normal subjects): grant every org member contributor access.
-- Org owner/admin already inherit visibility; grants remain harmless.
INSERT INTO subject_members (subject_id, user_id, role, created_at)
SELECT s.id, om.user_id, 'contributor', strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
FROM subjects s
INNER JOIN organization_members om ON om.organization_id = s.organization_id
WHERE s.visibility = 'normal'
  AND NOT EXISTS (SELECT 1 FROM subject_members sm WHERE sm.subject_id = s.id)
  AND NOT EXISTS (SELECT 1 FROM team_subject_roles tsr WHERE tsr.subject_id = s.id)
  AND NOT EXISTS (
      SELECT 1 FROM subject_members sm2
      WHERE sm2.subject_id = s.id AND sm2.user_id = om.user_id
  );

-- +goose Down
-- Irreversible data backfill; no-op down.
SELECT 1;
