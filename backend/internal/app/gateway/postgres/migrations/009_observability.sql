-- W3C trace context survives outbox publication, retries and worker restarts.
-- No baggage, credentials, prompts or user metadata are stored in these fields.
ALTER TABLE report_jobs ADD COLUMN trace_parent varchar(55) NOT NULL DEFAULT '';
ALTER TABLE report_jobs ADD COLUMN trace_state varchar(512) NOT NULL DEFAULT '';
ALTER TABLE report_events ADD COLUMN trace_parent varchar(55) NOT NULL DEFAULT '';
ALTER TABLE report_events ADD COLUMN trace_state varchar(512) NOT NULL DEFAULT '';
