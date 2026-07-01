-- Seed default roles
INSERT INTO roles (name, description) VALUES
    ('viewer', 'Can view the app launcher catalog'),
    ('deployer', 'Can view the app launcher catalog'),
    ('security-team', 'Can view security dashboard and notifications'),
    ('admin', 'Full administrative access');

-- Casbin policies (ptype, v0=role, v1=object, v2=action)
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', '*', '*'),
    ('p', 'security-team', 'security', 'read'),
    ('p', 'security-team', 'notifications', 'read'),
    ('p', 'deployer', 'catalog', 'read'),
    ('p', 'viewer', 'catalog', 'read'),
    ('p', 'admin', 'admin', '*'),
    ('p', 'admin', 'catalog', '*'),
    ('p', 'admin', 'security', '*'),
    ('p', 'admin', 'notifications', '*');

-- Bootstrap admin user is created via the first-run /api/setup flow
