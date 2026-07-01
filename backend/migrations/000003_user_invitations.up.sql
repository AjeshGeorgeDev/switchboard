CREATE TABLE user_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    username TEXT NOT NULL,
    display_name TEXT,
    role_ids UUID[] NOT NULL DEFAULT '{}',
    token_hash TEXT NOT NULL UNIQUE,
    invited_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX user_invitations_pending_email_idx
    ON user_invitations (lower(email))
    WHERE accepted_at IS NULL;

CREATE UNIQUE INDEX user_invitations_pending_username_idx
    ON user_invitations (lower(username))
    WHERE accepted_at IS NULL;
