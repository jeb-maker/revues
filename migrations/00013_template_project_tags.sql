-- +goose Up
CREATE TABLE project_tags (
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (project_id, tag)
);

CREATE TABLE template_tags (
    template_id INTEGER NOT NULL REFERENCES checklist_templates(id) ON DELETE CASCADE,
    tag         TEXT NOT NULL,
    PRIMARY KEY (template_id, tag)
);

ALTER TABLE checklist_templates ADD COLUMN organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE;

UPDATE checklist_templates
SET organization_id = (
    SELECT p.organization_id FROM projects p WHERE p.id = checklist_templates.project_id
)
WHERE project_id IS NOT NULL;

UPDATE checklist_templates
SET organization_id = (SELECT id FROM organizations WHERE slug = 'default')
WHERE organization_id IS NULL;

INSERT INTO project_tags (project_id, tag)
SELECT DISTINCT project_id, 'legacy-' || project_id
FROM checklist_templates
WHERE project_id IS NOT NULL;

INSERT INTO template_tags (template_id, tag)
SELECT id, 'legacy-' || project_id
FROM checklist_templates
WHERE project_id IS NOT NULL;

UPDATE checklist_templates SET project_id = NULL;

CREATE INDEX idx_checklist_templates_organization ON checklist_templates(organization_id);

-- +goose Down
UPDATE checklist_templates
SET project_id = CAST(substr(tag, 8) AS INTEGER)
FROM template_tags
WHERE checklist_templates.id = template_tags.template_id
  AND template_tags.tag LIKE 'legacy-%';

DROP INDEX IF EXISTS idx_checklist_templates_organization;
ALTER TABLE checklist_templates DROP COLUMN organization_id;

DROP TABLE IF EXISTS template_tags;
DROP TABLE IF EXISTS project_tags;
