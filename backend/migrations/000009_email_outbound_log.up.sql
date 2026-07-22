CREATE TABLE email_outbound_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    subject TEXT NOT NULL,
    body_preview TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('sent', 'failed', 'skipped')),
    error_message TEXT,
    triggered_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_email_outbound_log_created_at ON email_outbound_log (created_at DESC);
CREATE INDEX idx_email_outbound_log_event_type ON email_outbound_log (event_type);

CREATE TABLE email_outbound_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    log_id UUID NOT NULL REFERENCES email_outbound_log(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status TEXT NOT NULL CHECK (status IN ('sent', 'failed', 'skipped')),
    error_message TEXT
);

CREATE INDEX idx_email_outbound_recipients_log_id ON email_outbound_recipients (log_id);
