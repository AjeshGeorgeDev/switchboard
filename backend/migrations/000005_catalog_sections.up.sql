CREATE TABLE catalog_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE applications
    ADD COLUMN section_id UUID REFERENCES catalog_sections(id) ON DELETE SET NULL;

CREATE INDEX idx_applications_section_id ON applications(section_id);
