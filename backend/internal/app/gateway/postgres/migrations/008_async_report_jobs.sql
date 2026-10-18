CREATE TABLE report_jobs (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id),
    request_key uuid NOT NULL,
    request jsonb NOT NULL,
    status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','processing','completed','failed')),
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    max_attempts integer NOT NULL DEFAULT 3 CHECK (max_attempts BETWEEN 1 AND 10),
    available_at timestamptz NOT NULL DEFAULT now(),
    lease_token uuid,
    lease_until timestamptz,
    progress_agent text NOT NULL DEFAULT 'fila',
    progress_message text NOT NULL DEFAULT 'Aguardando processamento',
    error_message text NOT NULL DEFAULT '',
    report_id uuid REFERENCES generated_reports(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    UNIQUE(user_id, request_key)
);
CREATE UNIQUE INDEX report_jobs_one_active_per_user ON report_jobs(user_id) WHERE status IN ('queued','processing');
CREATE INDEX report_jobs_owner_created ON report_jobs(user_id,created_at DESC,id DESC);
CREATE INDEX report_jobs_recovery ON report_jobs(status,lease_until,available_at);

-- Immutable lifecycle history; requested events also act as the transactional outbox.
CREATE TABLE report_events (
    id uuid PRIMARY KEY,
    job_id uuid NOT NULL REFERENCES report_jobs(id),
    event_type text NOT NULL,
    event_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    available_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    publish_token uuid,
    publish_lease_until timestamptz
);
CREATE INDEX report_events_job_created ON report_events(job_id,created_at,id);
CREATE INDEX report_events_pending ON report_events(available_at,created_at)
    WHERE event_type='report.requested' AND published_at IS NULL;
