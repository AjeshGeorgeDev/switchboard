ALTER TABLE oidc_providers
    DROP COLUMN IF EXISTS claim_email,
    DROP COLUMN IF EXISTS claim_name,
    DROP COLUMN IF EXISTS claim_subject,
    DROP COLUMN IF EXISTS claim_groups,
    DROP COLUMN IF EXISTS group_role_mappings;
