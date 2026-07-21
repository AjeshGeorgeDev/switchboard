ALTER TABLE oidc_providers
    ADD COLUMN claim_email TEXT NOT NULL DEFAULT 'email',
    ADD COLUMN claim_name TEXT NOT NULL DEFAULT 'name',
    ADD COLUMN claim_subject TEXT NOT NULL DEFAULT 'sub',
    ADD COLUMN claim_groups TEXT NOT NULL DEFAULT 'groups',
    ADD COLUMN group_role_mappings JSONB NOT NULL DEFAULT '[]'::jsonb;
