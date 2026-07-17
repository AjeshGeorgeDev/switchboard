CREATE TYPE auth_type AS ENUM ('local', 'oidc');
CREATE TYPE access_type AS ENUM ('ip_port', 'url');
CREATE TYPE cve_severity AS ENUM ('critical', 'high', 'medium', 'low', 'unknown');
CREATE TYPE finding_source AS ENUM ('weekly_pull', 'webhook');
CREATE TYPE deploy_status AS ENUM ('success', 'failed', 'partial');
CREATE TYPE notification_channel AS ENUM ('email', 'teams', 'in_app');
CREATE TYPE notification_event_type AS ENUM ('weekly_digest', 'deployment_report', 'critical_cve');
CREATE TYPE notification_severity AS ENUM ('info', 'warning', 'critical');

CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    display_name TEXT,
    auth_type auth_type NOT NULL,
    password_hash TEXT,
    oidc_provider TEXT,
    oidc_subject TEXT,
    oidc_email TEXT,
    last_login_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (oidc_provider, oidc_subject)
);

CREATE TABLE oidc_providers (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    issuer_url TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT ARRAY['openid','profile','email'],
    auto_provision BOOLEAN NOT NULL DEFAULT TRUE,
    default_role_id UUID REFERENCES roles(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE user_invitations (
    id UUID PRIMARY KEY,
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

CREATE TABLE catalog_sections (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE applications (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    icon_url TEXT,
    access_type access_type NOT NULL,
    target_host TEXT NOT NULL,
    target_port INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    section_id UUID REFERENCES catalog_sections(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE app_role_access (
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (application_id, role_id)
);

CREATE TABLE cve_findings (
    id UUID PRIMARY KEY,
    image_name TEXT NOT NULL,
    image_tag TEXT NOT NULL,
    cve_id TEXT NOT NULL,
    severity cve_severity NOT NULL,
    package_name TEXT,
    installed_version TEXT,
    fixed_version TEXT,
    source finding_source NOT NULL,
    scan_date TIMESTAMPTZ NOT NULL,
    raw_payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (image_name, image_tag, cve_id, source)
);

CREATE TABLE deployment_reports (
    id UUID PRIMARY KEY,
    app_name TEXT NOT NULL,
    image_name TEXT NOT NULL,
    image_tag TEXT NOT NULL,
    triggered_by TEXT,
    status deploy_status NOT NULL,
    critical_count INT NOT NULL DEFAULT 0,
    high_count INT NOT NULL DEFAULT 0,
    medium_count INT NOT NULL DEFAULT 0,
    low_count INT NOT NULL DEFAULT 0,
    report_url TEXT,
    raw_payload JSONB,
    payload_hash TEXT UNIQUE,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    user_agent TEXT,
    ip_address TEXT
);

CREATE TABLE notification_preferences (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel notification_channel NOT NULL,
    event_type notification_event_type NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (user_id, channel, event_type)
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    event_type notification_event_type NOT NULL,
    severity notification_severity NOT NULL DEFAULT 'info',
    read BOOLEAN NOT NULL DEFAULT FALSE,
    link_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE teams_webhook_configs (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    webhook_url TEXT NOT NULL,
    event_types TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100),
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100)
);

CREATE TYPE webhook_source AS ENUM ('harbor', 'trivy');
CREATE TYPE webhook_event_status AS ENUM ('accepted', 'processed', 'failed');

CREATE TABLE webhook_events (
    id UUID PRIMARY KEY,
    source webhook_source NOT NULL,
    status webhook_event_status NOT NULL DEFAULT 'accepted',
    payload JSONB NOT NULL,
    payload_preview TEXT,
    error_message TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_username TEXT,
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    details JSONB,
    ip_address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
